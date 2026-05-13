package com.crondispatch.service;

import com.crondispatch.enums.ExecutionResult;
import com.crondispatch.enums.TaskStatus;
import com.crondispatch.model.CallbackLog;
import com.crondispatch.model.ExecutionHistory;
import com.crondispatch.model.Task;
import com.crondispatch.store.TaskStore;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.fasterxml.jackson.datatype.jsr310.JavaTimeModule;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.stereotype.Service;

import java.io.IOException;
import java.net.ConnectException;
import java.net.SocketTimeoutException;
import java.net.URI;
import java.net.http.HttpClient;
import java.net.http.HttpRequest;
import java.net.http.HttpResponse;
import java.time.Duration;
import java.time.Instant;
import java.util.HashMap;
import java.util.Map;
import java.util.Queue;
import java.util.concurrent.ConcurrentHashMap;
import java.util.concurrent.ExecutorService;
import java.util.concurrent.Executors;
import java.util.concurrent.TimeoutException;

@Service
public class TaskExecutorService {
    private static final Logger logger = LoggerFactory.getLogger(TaskExecutorService.class);
    private static final int MAX_QUEUE_SIZE = 5;
    private static final int RETRY_BASE_DELAY = 5;
    private static final int HTTP_TIMEOUT = 30;

    private final TaskStore taskStore;
    private final HistoryService historyService;
    private final HttpClient httpClient;
    private final ObjectMapper objectMapper;
    private final ExecutorService executorService;
    private final Map<String, Boolean> runningTasks = new ConcurrentHashMap<>();
    private final Map<String, Queue<String>> taskQueues = new ConcurrentHashMap<>();

    public TaskExecutorService(TaskStore taskStore, HistoryService historyService) {
        this.taskStore = taskStore;
        this.historyService = historyService;
        this.httpClient = HttpClient.newBuilder()
                .connectTimeout(Duration.ofSeconds(HTTP_TIMEOUT))
                .build();
        this.objectMapper = new ObjectMapper();
        this.objectMapper.registerModule(new JavaTimeModule());
        this.executorService = Executors.newCachedThreadPool();
    }

    public void executeTask(String taskId) {
        Task task = taskStore.get(taskId);
        if (task == null) {
            logger.warn("任务 [{}] 不存在，跳过执行", taskId);
            return;
        }

        Boolean isRunning = runningTasks.get(taskId);
        if (Boolean.TRUE.equals(isRunning)) {
            handleMisfire(task);
            return;
        }

        runningTasks.put(taskId, true);
        logger.info("开始执行任务 [{}] '{}'", task.getId(), task.getName());
        executorService.submit(() -> runTask(task));
    }

    private void handleMisfire(Task task) {
        if (task.getMisfireStrategy() == null || task.getMisfireStrategy() == com.crondispatch.enums.MisfireStrategy.SKIPPED) {
            ExecutionHistory history = historyService.createHistory(task.getId());
            history.setResult(ExecutionResult.SKIPPED);
            history.setErrorMessage("任务正在执行中，触发跳过策略");
            history.setEndTime(Instant.now());
            historyService.saveHistory(history);
            logger.warn("任务 [{}] '{}' 正在执行中，本次触发已跳过 (SKIPPED策略)", 
                task.getId(), task.getName());
        } else {
            Queue<String> queue = taskQueues.computeIfAbsent(task.getId(), k -> new java.util.LinkedList<>());
            if (queue.size() < MAX_QUEUE_SIZE) {
                queue.offer(task.getId());
                logger.info("任务 [{}] '{}' 正在执行中，加入队列 (队列大小: {}, QUEUED策略)", 
                    task.getId(), task.getName(), queue.size());
            } else {
                ExecutionHistory history = historyService.createHistory(task.getId());
                history.setResult(ExecutionResult.SKIPPED);
                history.setErrorMessage("任务正在执行中，队列已满");
                history.setEndTime(Instant.now());
                historyService.saveHistory(history);
                logger.warn("任务 [{}] '{}' 正在执行中，队列已满，跳过本次触发", 
                    task.getId(), task.getName());
            }
        }
    }

    public void cleanupTask(String taskId) {
        runningTasks.remove(taskId);
        taskQueues.remove(taskId);
        logger.info("已清理任务 [{}] 的状态信息", taskId);
    }

    private void runTask(Task task) {
        ExecutionHistory history = historyService.createHistory(task.getId());
        
        try {
            task.setStatus(TaskStatus.RUNNING);
            taskStore.save(task);

            executeWithRetry(task, history);

            if (history.getResult() == ExecutionResult.SUCCESS) {
                task.setStatus(TaskStatus.SUCCESS);
            } else {
                task.setStatus(TaskStatus.FAILED);
            }
            task.setCurrentRetryCount(0);
            taskStore.save(task);

        } finally {
            history.setEndTime(Instant.now());
            historyService.saveHistory(history);
            runningTasks.put(task.getId(), false);
            
            processQueuedTasks(task.getId());
        }
    }

    private void executeWithRetry(Task task, ExecutionHistory history) {
        int retryCount = 0;
        boolean success = false;
        String lastError = null;
        String lastErrorType = null;
        CallbackLog lastLog = null;

        while (!success && retryCount <= (task.isRetryEnabled() ? task.getMaxRetryCount() : 0)) {
            if (retryCount > 0) {
                long delay = RETRY_BASE_DELAY * (long) Math.pow(2, retryCount - 1);
                try {
                    Thread.sleep(delay * 1000);
                } catch (InterruptedException e) {
                    Thread.currentThread().interrupt();
                    break;
                }
            }

            try {
                CallbackLog log = doCallback(task);
                lastLog = log;
                historyService.saveLog(log);
                history.setRetryCount(retryCount);

                if (log.getResponseStatus() >= 200 && log.getResponseStatus() < 300) {
                    success = true;
                    history.setResult(ExecutionResult.SUCCESS);
                } else {
                    lastError = "HTTP " + log.getResponseStatus();
                    lastErrorType = "HTTP_NON_2XX";
                }
            } catch (ConnectException e) {
                lastError = e.getMessage();
                lastErrorType = "CONNECTION_TIMEOUT";
            } catch (SocketTimeoutException e) {
                lastError = e.getMessage();
                lastErrorType = "SOCKET_TIMEOUT";
            } catch (TimeoutException e) {
                lastError = e.getMessage();
                lastErrorType = "TIMEOUT";
            } catch (IOException e) {
                lastError = e.getMessage();
                lastErrorType = "IO_ERROR";
                if (e.getMessage() != null && e.getMessage().contains("UnknownHost")) {
                    lastErrorType = "DNS_FAILURE";
                }
            } catch (Exception e) {
                lastError = e.getMessage();
                lastErrorType = "UNKNOWN_ERROR";
            }

            if (!success) {
                retryCount++;
                if (lastLog != null) {
                    lastLog.setErrorType(lastErrorType);
                }
            }
        }

        if (!success) {
            history.setResult(ExecutionResult.FAILED);
            history.setErrorMessage(lastError);
            if (lastLog != null) {
                historyService.saveLog(lastLog);
            }
        }
    }

    private CallbackLog doCallback(Task task) throws IOException, InterruptedException, TimeoutException {
        CallbackLog log = new CallbackLog();
        log.setTaskId(task.getId());

        Map<String, Object> requestBody = new HashMap<>();
        requestBody.put("taskId", task.getId());
        requestBody.put("taskName", task.getName());
        if (task.getParams() != null) {
            requestBody.putAll(task.getParams());
        }

        String jsonBody = objectMapper.writeValueAsString(requestBody);
        log.setRequestBody(jsonBody);

        HttpRequest request = HttpRequest.newBuilder()
                .uri(URI.create(task.getCallbackUrl()))
                .timeout(Duration.ofSeconds(HTTP_TIMEOUT))
                .header("Content-Type", "application/json")
                .POST(HttpRequest.BodyPublishers.ofString(jsonBody))
                .build();

        HttpResponse<String> response = httpClient.send(request, HttpResponse.BodyHandlers.ofString());
        log.setResponseStatus(response.statusCode());
        log.setResponseBody(response.body());

        return log;
    }

    private void processQueuedTasks(String taskId) {
        Queue<String> queue = taskQueues.get(taskId);
        if (queue != null && !queue.isEmpty()) {
            String queuedTaskId = queue.poll();
            if (queuedTaskId != null) {
                executeTask(queuedTaskId);
            }
        }
    }
}

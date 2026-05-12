package com.crondispatch.service;

import com.crondispatch.enums.MisfireStrategy;
import com.crondispatch.enums.TaskType;
import com.crondispatch.model.Task;
import com.crondispatch.store.TaskStore;
import com.crondispatch.util.CronExpressionParser;
import jakarta.annotation.PostConstruct;
import jakarta.annotation.PreDestroy;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.stereotype.Service;

import java.time.Duration;
import java.time.Instant;
import java.util.LinkedList;
import java.util.List;
import java.util.Map;
import java.util.Queue;
import java.util.concurrent.ConcurrentHashMap;
import java.util.concurrent.Executors;
import java.util.concurrent.ScheduledExecutorService;
import java.util.concurrent.ScheduledFuture;
import java.util.concurrent.TimeUnit;

@Service
public class TaskSchedulerService {
    private static final Logger logger = LoggerFactory.getLogger(TaskSchedulerService.class);
    private static final int MAX_QUEUE_SIZE = 5;

    private final TaskStore taskStore;
    private final TaskExecutorService taskExecutorService;
    private final ScheduledExecutorService scheduler;
    private final Map<String, ScheduledFuture<?>> scheduledTasks = new ConcurrentHashMap<>();
    private final Map<String, Queue<String>> taskQueues = new ConcurrentHashMap<>();
    private final Map<String, Boolean> runningTasks = new ConcurrentHashMap<>();

    public TaskSchedulerService(TaskStore taskStore, TaskExecutorService taskExecutorService) {
        this.taskStore = taskStore;
        this.taskExecutorService = taskExecutorService;
        this.scheduler = Executors.newScheduledThreadPool(Runtime.getRuntime().availableProcessors());
    }

    @PostConstruct
    public void restoreScheduledTasks() {
        List<Task> tasks = taskStore.getAll();
        for (Task task : tasks) {
            try {
                scheduleTask(task);
            } catch (Exception e) {
                logger.error("恢复任务失败: {}", task.getId(), e);
            }
        }
    }

    @PreDestroy
    public void shutdown() {
        scheduler.shutdown();
        try {
            if (!scheduler.awaitTermination(60, TimeUnit.SECONDS)) {
                scheduler.shutdownNow();
            }
        } catch (InterruptedException e) {
            scheduler.shutdownNow();
        }
    }

    public Task registerTask(Task task) {
        if (task.getType() == TaskType.CRON) {
            if (task.getCronExpression() == null || task.getCronExpression().isEmpty()) {
                throw new IllegalArgumentException("Cron任务必须指定cron表达式");
            }
            if (!CronExpressionParser.isValid(task.getCronExpression())) {
                throw new IllegalArgumentException("无效的cron表达式: " + task.getCronExpression());
            }
        } else if (task.getType() == TaskType.DELAYED) {
            if (task.getDelaySeconds() == null || task.getDelaySeconds() <= 0) {
                throw new IllegalArgumentException("延迟任务必须指定正的延迟秒数");
            }
        } else {
            throw new IllegalArgumentException("未知的任务类型: " + task.getType());
        }

        if (task.getCallbackUrl() == null || task.getCallbackUrl().isEmpty()) {
            throw new IllegalArgumentException("回调URL不能为空");
        }

        if (task.getMisfireStrategy() == null) {
            task.setMisfireStrategy(MisfireStrategy.SKIPPED);
        }

        if (task.getType() == TaskType.CRON) {
            CronExpressionParser parser = new CronExpressionParser(task.getCronExpression());
            task.setNextExecutionTime(parser.getNextExecutionTime(Instant.now()));
        } else {
            task.setNextExecutionTime(Instant.now().plusSeconds(task.getDelaySeconds()));
        }

        taskStore.save(task);
        scheduleTask(task);
        return task;
    }

    public boolean deleteTask(String taskId) {
        cancelScheduledTask(taskId);
        taskQueues.remove(taskId);
        return taskStore.delete(taskId);
    }

    public List<Task> getAllTasks() {
        return taskStore.getAll();
    }

    public Task getTask(String taskId) {
        return taskStore.get(taskId);
    }

    private void scheduleTask(Task task) {
        cancelScheduledTask(task.getId());

        if (task.getType() == TaskType.CRON) {
            scheduleCronTask(task);
        } else {
            scheduleDelayedTask(task);
        }
    }

    private void scheduleCronTask(Task task) {
        CronExpressionParser parser = new CronExpressionParser(task.getCronExpression());
        Instant nextExecution = task.getNextExecutionTime();
        
        if (nextExecution == null || nextExecution.isBefore(Instant.now())) {
            nextExecution = parser.getNextExecutionTime(Instant.now());
            task.setNextExecutionTime(nextExecution);
            taskStore.save(task);
        }

        long delay = Duration.between(Instant.now(), nextExecution).toMillis();
        if (delay < 0) {
            delay = 0;
        }

        ScheduledFuture<?> future = scheduler.schedule(() -> {
            try {
                triggerTask(task.getId());
            } catch (Exception e) {
                logger.error("执行Cron任务失败: {}", task.getId(), e);
            }
            
            Task updatedTask = taskStore.get(task.getId());
            if (updatedTask != null) {
                Instant newNextExecution = parser.getNextExecutionTime(Instant.now());
                updatedTask.setNextExecutionTime(newNextExecution);
                taskStore.save(updatedTask);
                scheduleCronTask(updatedTask);
            }
        }, delay, TimeUnit.MILLISECONDS);

        scheduledTasks.put(task.getId(), future);
    }

    private void scheduleDelayedTask(Task task) {
        long delay = Duration.between(Instant.now(), task.getNextExecutionTime()).toMillis();
        if (delay < 0) {
            delay = 0;
        }

        ScheduledFuture<?> future = scheduler.schedule(() -> {
            try {
                triggerTask(task.getId());
            } catch (Exception e) {
                logger.error("执行延迟任务失败: {}", task.getId(), e);
            }
            taskStore.delete(task.getId());
            scheduledTasks.remove(task.getId());
        }, delay, TimeUnit.MILLISECONDS);

        scheduledTasks.put(task.getId(), future);
    }

    private void cancelScheduledTask(String taskId) {
        ScheduledFuture<?> future = scheduledTasks.remove(taskId);
        if (future != null) {
            future.cancel(false);
        }
    }

    public void triggerTask(String taskId) {
        Task task = taskStore.get(taskId);
        if (task == null) {
            return;
        }

        Boolean isRunning = runningTasks.get(taskId);
        if (Boolean.TRUE.equals(isRunning)) {
            handleMisfire(task);
            return;
        }

        taskExecutorService.executeTask(taskId);
    }

    private void handleMisfire(Task task) {
        if (task.getMisfireStrategy() == MisfireStrategy.QUEUED) {
            Queue<String> queue = taskQueues.computeIfAbsent(task.getId(), k -> new LinkedList<>());
            if (queue.size() < MAX_QUEUE_SIZE) {
                queue.offer(task.getId());
                logger.info("任务 {} 正在执行中，加入队列 (队列大小: {})", task.getId(), queue.size());
            } else {
                logger.warn("任务 {} 正在执行中，队列已满，跳过本次触发", task.getId());
            }
        } else {
            logger.info("任务 {} 正在执行中，触发跳过策略", task.getId());
        }
    }
}

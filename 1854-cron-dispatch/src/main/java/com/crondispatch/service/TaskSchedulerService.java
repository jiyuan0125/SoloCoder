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
import java.util.List;
import java.util.Map;
import java.util.concurrent.ConcurrentHashMap;
import java.util.concurrent.Executors;
import java.util.concurrent.ScheduledExecutorService;
import java.util.concurrent.ScheduledFuture;
import java.util.concurrent.TimeUnit;

@Service
public class TaskSchedulerService {
    private static final Logger logger = LoggerFactory.getLogger(TaskSchedulerService.class);
    private static final long MIN_SCHEDULE_DELAY_MS = 100;

    private final TaskStore taskStore;
    private final TaskExecutorService taskExecutorService;
    private final ScheduledExecutorService scheduler;
    private final Map<String, ScheduledFuture<?>> scheduledTasks = new ConcurrentHashMap<>();
    private final Map<String, Object> schedulingLocks = new ConcurrentHashMap<>();

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
        taskExecutorService.cleanupTask(taskId);
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
            doScheduleCronTask(task);
        } else {
            scheduleDelayedTask(task);
        }
    }

    private void doScheduleCronTask(Task task) {
        Object lock = schedulingLocks.computeIfAbsent(task.getId(), k -> new Object());
        synchronized (lock) {
            Instant now = Instant.now();
            CronExpressionParser parser = new CronExpressionParser(task.getCronExpression());
            Instant nextExecution = task.getNextExecutionTime();
            
            if (nextExecution == null || !nextExecution.isAfter(now)) {
                nextExecution = parser.getNextExecutionTime(now);
                if (nextExecution == null) {
                    logger.error("无法计算任务 [{}] '{}' 的下次执行时间", task.getId(), task.getName());
                    return;
                }
            }

            long delay = Duration.between(now, nextExecution).toMillis();
            
            if (delay <= 0) {
                logger.warn("任务 [{}] '{}' 计算出的延迟 {}ms 为非正值，从下次执行时间点重新计算", 
                    task.getId(), task.getName(), delay);
                Instant nextMinute = nextExecution.plusMillis(1);
                nextExecution = parser.getNextExecutionTime(nextMinute);
                if (nextExecution == null) {
                    logger.error("无法计算任务 [{}] '{}' 的下次执行时间", task.getId(), task.getName());
                    return;
                }
                delay = Duration.between(now, nextExecution).toMillis();
            }
            
            if (delay < MIN_SCHEDULE_DELAY_MS) {
                logger.warn("任务 [{}] '{}' 计算出的延迟 {}ms < {}ms，跳过本次调度，等待下一轮", 
                    task.getId(), task.getName(), delay, MIN_SCHEDULE_DELAY_MS);
                Instant nextMinute = nextExecution.plusMillis(1);
                nextExecution = parser.getNextExecutionTime(nextMinute);
                if (nextExecution == null) {
                    logger.error("无法计算任务 [{}] '{}' 的下次执行时间", task.getId(), task.getName());
                    return;
                }
                delay = Duration.between(now, nextExecution).toMillis();
            }

            task.setNextExecutionTime(nextExecution);
            taskStore.save(task);

            logger.info("调度Cron任务 [{}] '{}', cron: {}, 下次执行: {}, 延迟: {}ms", 
                task.getId(), task.getName(), task.getCronExpression(), nextExecution, delay);

            final Instant scheduledNextExecution = nextExecution;
            ScheduledFuture<?> future = scheduler.schedule(() -> {
                logger.info("===== 触发Cron任务 [{}] '{}', 计划执行时间: {} =====", 
                    task.getId(), task.getName(), scheduledNextExecution);
                try {
                    taskExecutorService.executeTask(task.getId());
                } catch (Exception e) {
                    logger.error("执行Cron任务失败: {}", task.getId(), e);
                }
                
                Task updatedTask = taskStore.get(task.getId());
                if (updatedTask != null) {
                    logger.info("准备重新调度Cron任务 [{}] '{}'", updatedTask.getId(), updatedTask.getName());
                    scheduleTask(updatedTask);
                } else {
                    logger.warn("任务 [{}] 已被删除，跳过重新调度", task.getId());
                }
            }, delay, TimeUnit.MILLISECONDS);

            scheduledTasks.put(task.getId(), future);
            logger.debug("任务 [{}] 已加入调度队列，future: {}", task.getId(), future);
        }
    }

    private void scheduleDelayedTask(Task task) {
        long delay = Duration.between(Instant.now(), task.getNextExecutionTime()).toMillis();
        if (delay < 0) {
            delay = 0;
        }

        logger.info("调度延迟任务 [{}] '{}', 延迟: {}ms, 计划执行时间: {}", 
            task.getId(), task.getName(), delay, task.getNextExecutionTime());

        ScheduledFuture<?> future = scheduler.schedule(() -> {
            logger.info("触发延迟任务 [{}] '{}'", task.getId(), task.getName());
            try {
                taskExecutorService.executeTask(task.getId());
            } catch (Exception e) {
                logger.error("执行延迟任务失败: {}", task.getId(), e);
            }
            taskStore.delete(task.getId());
            scheduledTasks.remove(task.getId());
            logger.info("延迟任务 [{}] '{}' 已从存储中移除", task.getId(), task.getName());
        }, delay, TimeUnit.MILLISECONDS);

        scheduledTasks.put(task.getId(), future);
    }

    private void cancelScheduledTask(String taskId) {
        ScheduledFuture<?> future = scheduledTasks.remove(taskId);
        if (future != null) {
            future.cancel(false);
            logger.info("已取消任务 [{}] 的调度", taskId);
        }
    }
}

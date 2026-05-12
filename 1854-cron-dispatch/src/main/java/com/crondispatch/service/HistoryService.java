package com.crondispatch.service;

import com.crondispatch.model.CallbackLog;
import com.crondispatch.model.ExecutionHistory;
import com.fasterxml.jackson.core.type.TypeReference;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.fasterxml.jackson.databind.SerializationFeature;
import com.fasterxml.jackson.datatype.jsr310.JavaTimeModule;
import jakarta.annotation.PostConstruct;
import org.springframework.stereotype.Component;

import java.io.File;
import java.io.IOException;
import java.util.ArrayList;
import java.util.List;
import java.util.concurrent.locks.ReentrantReadWriteLock;

@Component
public class HistoryService {
    private static final int MAX_HISTORY = 1000;
    private static final String DATA_DIR = "data";
    private static final String HISTORY_FILE = DATA_DIR + "/history.json";
    private static final String LOGS_FILE = DATA_DIR + "/logs.json";

    private final List<ExecutionHistory> histories = new ArrayList<>();
    private final List<CallbackLog> logs = new ArrayList<>();
    private final ReentrantReadWriteLock historyLock = new ReentrantReadWriteLock();
    private final ReentrantReadWriteLock logLock = new ReentrantReadWriteLock();
    private final ObjectMapper objectMapper;

    public HistoryService() {
        this.objectMapper = new ObjectMapper();
        this.objectMapper.registerModule(new JavaTimeModule());
        this.objectMapper.disable(SerializationFeature.WRITE_DATES_AS_TIMESTAMPS);
    }

    @PostConstruct
    public void init() {
        loadHistory();
        loadLogs();
    }

    public ExecutionHistory createHistory(String taskId) {
        ExecutionHistory history = new ExecutionHistory();
        history.setTaskId(taskId);
        return history;
    }

    public void saveHistory(ExecutionHistory history) {
        historyLock.writeLock().lock();
        try {
            histories.add(history);
            while (histories.size() > MAX_HISTORY) {
                histories.remove(0);
            }
            persistHistory();
        } finally {
            historyLock.writeLock().unlock();
        }
    }

    public List<ExecutionHistory> getTaskHistory(String taskId) {
        historyLock.readLock().lock();
        try {
            List<ExecutionHistory> result = new ArrayList<>();
            for (ExecutionHistory h : histories) {
                if (taskId.equals(h.getTaskId())) {
                    result.add(h);
                }
            }
            return result;
        } finally {
            historyLock.readLock().unlock();
        }
    }

    public void saveLog(CallbackLog log) {
        logLock.writeLock().lock();
        try {
            logs.add(log);
            while (logs.size() > MAX_HISTORY * 3) {
                logs.remove(0);
            }
            persistLogs();
        } finally {
            logLock.writeLock().unlock();
        }
    }

    public List<CallbackLog> getTaskLogs(String taskId) {
        logLock.readLock().lock();
        try {
            List<CallbackLog> result = new ArrayList<>();
            for (CallbackLog log : logs) {
                if (taskId.equals(log.getTaskId())) {
                    result.add(log);
                }
            }
            return result;
        } finally {
            logLock.readLock().unlock();
        }
    }

    private void persistHistory() {
        try {
            File dir = new File(DATA_DIR);
            if (!dir.exists()) {
                dir.mkdirs();
            }
            objectMapper.writerWithDefaultPrettyPrinter().writeValue(new File(HISTORY_FILE), histories);
        } catch (IOException e) {
            throw new RuntimeException("保存历史记录失败", e);
        }
    }

    private void loadHistory() {
        File file = new File(HISTORY_FILE);
        if (!file.exists()) {
            return;
        }

        try {
            List<ExecutionHistory> loaded = objectMapper.readValue(file, new TypeReference<List<ExecutionHistory>>() {});
            histories.clear();
            histories.addAll(loaded);
        } catch (IOException e) {
            throw new RuntimeException("加载历史记录失败", e);
        }
    }

    private void persistLogs() {
        try {
            File dir = new File(DATA_DIR);
            if (!dir.exists()) {
                dir.mkdirs();
            }
            objectMapper.writerWithDefaultPrettyPrinter().writeValue(new File(LOGS_FILE), logs);
        } catch (IOException e) {
            throw new RuntimeException("保存日志失败", e);
        }
    }

    private void loadLogs() {
        File file = new File(LOGS_FILE);
        if (!file.exists()) {
            return;
        }

        try {
            List<CallbackLog> loaded = objectMapper.readValue(file, new TypeReference<List<CallbackLog>>() {});
            logs.clear();
            logs.addAll(loaded);
        } catch (IOException e) {
            throw new RuntimeException("加载日志失败", e);
        }
    }
}

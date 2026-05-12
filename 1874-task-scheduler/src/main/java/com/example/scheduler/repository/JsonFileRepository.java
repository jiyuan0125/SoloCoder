package com.example.scheduler.repository;

import com.example.scheduler.model.CallbackLog;
import com.example.scheduler.model.ExecutionHistory;
import com.example.scheduler.model.Task;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.fasterxml.jackson.databind.SerializationFeature;
import com.fasterxml.jackson.datatype.jsr310.JavaTimeModule;
import jakarta.annotation.PostConstruct;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.stereotype.Repository;

import java.io.IOException;
import java.nio.file.Files;
import java.nio.file.Path;
import java.nio.file.Paths;
import java.util.ArrayList;
import java.util.List;
import java.util.Map;
import java.util.concurrent.ConcurrentHashMap;
import java.util.stream.Collectors;

@Repository
public class JsonFileRepository {

    @Value("${scheduler.data-dir:data}")
    private String dataDir;

    private final ObjectMapper objectMapper;
    private final Map<String, Task> tasks = new ConcurrentHashMap<>();
    private final Map<String, List<ExecutionHistory>> executionHistory = new ConcurrentHashMap<>();
    private final Map<String, List<CallbackLog>> callbackLogs = new ConcurrentHashMap<>();

    private Path tasksFile;
    private Path historyFile;
    private Path callbacksFile;

    public JsonFileRepository() {
        this.objectMapper = new ObjectMapper();
        this.objectMapper.registerModule(new JavaTimeModule());
        this.objectMapper.disable(SerializationFeature.WRITE_DATES_AS_TIMESTAMPS);
        this.objectMapper.enable(SerializationFeature.INDENT_OUTPUT);
    }

    @PostConstruct
    public void init() throws IOException {
        Path dataPath = Paths.get(dataDir);
        if (!Files.exists(dataPath)) {
            Files.createDirectories(dataPath);
        }
        this.tasksFile = dataPath.resolve("tasks.json");
        this.historyFile = dataPath.resolve("execution-history.json");
        this.callbacksFile = dataPath.resolve("callbacks.json");

        loadTasks();
        loadExecutionHistory();
        loadCallbackLogs();
    }

    @SuppressWarnings("unchecked")
    private void loadTasks() throws IOException {
        if (!Files.exists(tasksFile)) {
            return;
        }
        String content = Files.readString(tasksFile);
        if (content.isEmpty()) {
            return;
        }
        List<Task> taskList = objectMapper.readValue(
            content,
            objectMapper.getTypeFactory().constructCollectionType(List.class, Task.class)
        );
        for (Task task : taskList) {
            tasks.put(task.getId(), task);
        }
    }

    @SuppressWarnings("unchecked")
    private void loadExecutionHistory() throws IOException {
        if (!Files.exists(historyFile)) {
            return;
        }
        String content = Files.readString(historyFile);
        if (content.isEmpty()) {
            return;
        }
        List<ExecutionHistory> allHistory = objectMapper.readValue(
            content,
            objectMapper.getTypeFactory().constructCollectionType(List.class, ExecutionHistory.class)
        );
        for (ExecutionHistory history : allHistory) {
            executionHistory.computeIfAbsent(history.getTaskId(), k -> new ArrayList<>()).add(history);
        }
    }

    @SuppressWarnings("unchecked")
    private void loadCallbackLogs() throws IOException {
        if (!Files.exists(callbacksFile)) {
            return;
        }
        String content = Files.readString(callbacksFile);
        if (content.isEmpty()) {
            return;
        }
        List<CallbackLog> allLogs = objectMapper.readValue(
            content,
            objectMapper.getTypeFactory().constructCollectionType(List.class, CallbackLog.class)
        );
        for (CallbackLog log : allLogs) {
            callbackLogs.computeIfAbsent(log.getTaskId(), k -> new ArrayList<>()).add(log);
        }
    }

    private void saveTasks() throws IOException {
        String json = objectMapper.writeValueAsString(new ArrayList<>(tasks.values()));
        Files.writeString(tasksFile, json);
    }

    private void saveExecutionHistory() throws IOException {
        List<ExecutionHistory> allHistory = executionHistory.values().stream()
            .flatMap(List::stream)
            .collect(Collectors.toList());
        String json = objectMapper.writeValueAsString(allHistory);
        Files.writeString(historyFile, json);
    }

    private void saveCallbackLogs() throws IOException {
        List<CallbackLog> allLogs = callbackLogs.values().stream()
            .flatMap(List::stream)
            .collect(Collectors.toList());
        String json = objectMapper.writeValueAsString(allLogs);
        Files.writeString(callbacksFile, json);
    }

    public Task saveTask(Task task) throws IOException {
        tasks.put(task.getId(), task);
        saveTasks();
        return task;
    }

    public Task findTaskById(String id) {
        return tasks.get(id);
    }

    public List<Task> findAllTasks() {
        return new ArrayList<>(tasks.values());
    }

    public void deleteTask(String id) throws IOException {
        tasks.remove(id);
        saveTasks();
    }

    public void addExecutionHistory(ExecutionHistory history) throws IOException {
        executionHistory.computeIfAbsent(history.getTaskId(), k -> new ArrayList<>()).add(history);
        saveExecutionHistory();
    }

    public List<ExecutionHistory> findExecutionHistoryByTaskId(String taskId) {
        List<ExecutionHistory> list = executionHistory.getOrDefault(taskId, new ArrayList<>());
        int start = Math.max(0, list.size() - 1000);
        return new ArrayList<>(list.subList(start, list.size()));
    }

    public void addCallbackLog(CallbackLog log) throws IOException {
        callbackLogs.computeIfAbsent(log.getTaskId(), k -> new ArrayList<>()).add(log);
        saveCallbackLogs();
    }

    public List<CallbackLog> findCallbackLogsByTaskId(String taskId) {
        List<CallbackLog> list = callbackLogs.getOrDefault(taskId, new ArrayList<>());
        int start = Math.max(0, list.size() - 1000);
        return new ArrayList<>(list.subList(start, list.size()));
    }
}

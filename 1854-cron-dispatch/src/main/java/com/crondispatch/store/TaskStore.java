package com.crondispatch.store;

import com.crondispatch.model.Task;
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
import java.util.Map;
import java.util.concurrent.ConcurrentHashMap;
import java.util.concurrent.locks.ReentrantReadWriteLock;

@Component
public class TaskStore {
    private static final String DATA_DIR = "data";
    private static final String TASKS_FILE = DATA_DIR + "/tasks.json";

    private final Map<String, Task> tasks = new ConcurrentHashMap<>();
    private final ReentrantReadWriteLock lock = new ReentrantReadWriteLock();
    private final ObjectMapper objectMapper;

    public TaskStore() {
        this.objectMapper = new ObjectMapper();
        this.objectMapper.registerModule(new JavaTimeModule());
        this.objectMapper.disable(SerializationFeature.WRITE_DATES_AS_TIMESTAMPS);
    }

    @PostConstruct
    public void init() {
        loadFromFile();
    }

    public void save(Task task) {
        lock.writeLock().lock();
        try {
            tasks.put(task.getId(), task);
            persistToFile();
        } finally {
            lock.writeLock().unlock();
        }
    }

    public Task get(String id) {
        lock.readLock().lock();
        try {
            return tasks.get(id);
        } finally {
            lock.readLock().unlock();
        }
    }

    public List<Task> getAll() {
        lock.readLock().lock();
        try {
            return new ArrayList<>(tasks.values());
        } finally {
            lock.readLock().unlock();
        }
    }

    public boolean delete(String id) {
        lock.writeLock().lock();
        try {
            Task removed = tasks.remove(id);
            if (removed != null) {
                persistToFile();
                return true;
            }
            return false;
        } finally {
            lock.writeLock().unlock();
        }
    }

    public boolean exists(String id) {
        lock.readLock().lock();
        try {
            return tasks.containsKey(id);
        } finally {
            lock.readLock().unlock();
        }
    }

    private void persistToFile() {
        try {
            File dir = new File(DATA_DIR);
            if (!dir.exists()) {
                dir.mkdirs();
            }
            objectMapper.writerWithDefaultPrettyPrinter().writeValue(new File(TASKS_FILE), tasks.values());
        } catch (IOException e) {
            throw new RuntimeException("保存任务失败", e);
        }
    }

    private void loadFromFile() {
        File file = new File(TASKS_FILE);
        if (!file.exists()) {
            return;
        }

        try {
            List<Task> loadedTasks = objectMapper.readValue(file, new TypeReference<List<Task>>() {});
            for (Task task : loadedTasks) {
                tasks.put(task.getId(), task);
            }
        } catch (IOException e) {
            throw new RuntimeException("加载任务失败", e);
        }
    }
}

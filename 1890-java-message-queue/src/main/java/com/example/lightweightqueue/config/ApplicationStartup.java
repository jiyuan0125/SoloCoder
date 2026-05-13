package com.example.lightweightqueue.config;

import com.example.lightweightqueue.service.PersistenceService;
import com.example.lightweightqueue.service.TopicManager;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.boot.CommandLineRunner;
import org.springframework.stereotype.Component;

@Slf4j
@Component
@RequiredArgsConstructor
public class ApplicationStartup implements CommandLineRunner {

    private final PersistenceService persistenceService;
    private final TopicManager topicManager;

    @Override
    public void run(String... args) {
        log.info("Starting application, restoring consumer progress...");
        persistenceService.init();
        persistenceService.restoreProgress(topicManager);
        log.info("Application startup complete");
    }
}

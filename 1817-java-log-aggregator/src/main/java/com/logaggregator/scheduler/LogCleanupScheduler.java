package com.logaggregator.scheduler;

import com.logaggregator.service.LogService;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.scheduling.annotation.Scheduled;
import org.springframework.stereotype.Component;

import java.time.LocalDate;
import java.util.Map;

@Slf4j
@Component
@RequiredArgsConstructor
public class LogCleanupScheduler {

    private final LogService logService;

    @Scheduled(cron = "0 0 2 * * ?")
    public void cleanupExpiredLogs() {
        log.info("Starting scheduled log cleanup task...");
        
        try {
            Map<LocalDate, Long> deletionStats = logService.cleanupExpiredLogs();
            
            if (deletionStats.isEmpty()) {
                log.info("No expired logs to clean up.");
            } else {
                log.info("Log cleanup completed. Daily statistics:");
                deletionStats.forEach((date, count) ->
                        log.info("  {}: {} logs deleted", date, count));
            }
        } catch (Exception e) {
            log.error("Error during scheduled log cleanup", e);
        }
    }
}

package com.health.service;

import com.health.model.ServiceRegistration;
import lombok.extern.slf4j.Slf4j;
import org.springframework.stereotype.Service;

import java.time.LocalDateTime;
import java.time.format.DateTimeFormatter;

@Slf4j
@Service
public class LoggingNotificationService implements NotificationService {

    private static final DateTimeFormatter FORMATTER = 
            DateTimeFormatter.ofPattern("yyyy-MM-dd HH:mm:ss");

    @Override
    public void notifyUnhealthy(ServiceRegistration service) {
        String timestamp = LocalDateTime.now().format(FORMATTER);
        log.error("=" .repeat(60));
        log.error("   ALERT: Service is UNHEALTHY!");
        log.error("=" .repeat(60));
        log.error("   Time:          {}", timestamp);
        log.error("   Service ID:    {}", service.getId());
        log.error("   Service Name:  {}", service.getName());
        log.error("   Check Type:    {}", service.getCheckType());
        log.error("   Endpoint:      {}:{}", service.getEndpoint(), 
                service.getPort() != null ? service.getPort() : "");
        log.error("   Path:          {}", service.getPath() != null ? service.getPath() : "-");
        log.error("   Status:        UNHEALTHY");
        log.error("   Message:       {}", service.getLastMessage());
        log.error("   Failures:      {}", service.getConsecutiveFailures());
        log.error("=" .repeat(60));
    }

    @Override
    public void notifyRecovered(ServiceRegistration service) {
        String timestamp = LocalDateTime.now().format(FORMATTER);
        log.info("=" .repeat(60));
        log.info("   NOTICE: Service has RECOVERED!");
        log.info("=" .repeat(60));
        log.info("   Time:          {}", timestamp);
        log.info("   Service ID:    {}", service.getId());
        log.info("   Service Name:  {}", service.getName());
        log.info("   Check Type:    {}", service.getCheckType());
        log.info("   Endpoint:      {}:{}", service.getEndpoint(), 
                service.getPort() != null ? service.getPort() : "");
        log.info("   Status:        HEALTHY (RECOVERED)");
        log.info("   Message:       {}", service.getLastMessage());
        log.info("=" .repeat(60));
    }
}
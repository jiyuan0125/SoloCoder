package com.example.healthchecker.service;

import com.example.healthchecker.model.HealthEvent;
import com.example.healthchecker.model.HealthStatus;
import com.example.healthchecker.repository.EventLogRepository;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;
import org.springframework.stereotype.Service;

import java.util.List;

@Service
public class EventLogService {

    private static final Logger logger = LoggerFactory.getLogger(EventLogService.class);

    private final EventLogRepository eventLogRepository;

    public EventLogService(EventLogRepository eventLogRepository) {
        this.eventLogRepository = eventLogRepository;
    }

    public void logStatusChange(String serviceName, HealthStatus previousStatus, HealthStatus newStatus, String reason) {
        if (previousStatus == newStatus) {
            return;
        }

        HealthEvent event = new HealthEvent();
        event.setServiceName(serviceName);
        event.setPreviousStatus(previousStatus);
        event.setNewStatus(newStatus);
        event.setReason(reason);

        eventLogRepository.save(event);

        logger.info("Health status changed: service={}, from={}, to={}, reason={}",
                serviceName, previousStatus, newStatus, reason);
    }

    public List<HealthEvent> getAllEvents() {
        return eventLogRepository.findAll();
    }

    public List<HealthEvent> getEventsByService(String serviceName) {
        return eventLogRepository.findByServiceName(serviceName);
    }
}

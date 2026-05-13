package com.example.healthchecker.repository;

import com.example.healthchecker.model.HealthEvent;
import org.springframework.stereotype.Repository;

import java.util.ArrayList;
import java.util.Collections;
import java.util.List;
import java.util.concurrent.CopyOnWriteArrayList;

@Repository
public class EventLogRepository {

    private final List<HealthEvent> events = new CopyOnWriteArrayList<>();

    public void save(HealthEvent event) {
        events.add(event);
    }

    public List<HealthEvent> findAll() {
        return Collections.unmodifiableList(new ArrayList<>(events));
    }

    public List<HealthEvent> findByServiceName(String serviceName) {
        List<HealthEvent> result = new ArrayList<>();
        for (HealthEvent event : events) {
            if (serviceName.equals(event.getServiceName())) {
                result.add(event);
            }
        }
        return Collections.unmodifiableList(result);
    }
}

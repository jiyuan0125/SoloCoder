package com.example.server.repository;

import com.example.common.dto.AlertDTO;
import com.example.common.enums.AlertLevel;
import org.springframework.stereotype.Repository;
import java.util.ArrayList;
import java.util.List;
import java.util.Map;
import java.util.Optional;
import java.util.concurrent.ConcurrentHashMap;
import java.util.stream.Collectors;

@Repository
public class AlertRepository {
    private final Map<String, AlertDTO> alerts = new ConcurrentHashMap<>();

    public AlertDTO save(AlertDTO alert) {
        alerts.put(alert.getId(), alert);
        return alert;
    }

    public Optional<AlertDTO> findById(String id) {
        return Optional.ofNullable(alerts.get(id));
    }

    public List<AlertDTO> findAll() {
        return new ArrayList<>(alerts.values());
    }

    public void deleteById(String id) {
        alerts.remove(id);
    }

    public boolean existsById(String id) {
        return alerts.containsKey(id);
    }

    public List<AlertDTO> findByResponsiblePerson(String responsiblePerson) {
        return alerts.values().stream()
                .filter(a -> responsiblePerson.equals(a.getResponsiblePerson()))
                .collect(Collectors.toList());
    }

    public List<AlertDTO> findByLevel(AlertLevel level) {
        return alerts.values().stream()
                .filter(a -> a.getLevel() == level)
                .collect(Collectors.toList());
    }

    public List<AlertDTO> findUnread() {
        return alerts.values().stream()
                .filter(a -> !a.isRead())
                .collect(Collectors.toList());
    }

    public long countUnread() {
        return alerts.values().stream()
                .filter(a -> !a.isRead())
                .count();
    }
}

package com.orgchart.server.repository;

import com.orgchart.server.entity.TransferHistory;
import org.springframework.stereotype.Repository;

import java.util.ArrayList;
import java.util.Comparator;
import java.util.List;
import java.util.Map;
import java.util.Optional;
import java.util.concurrent.ConcurrentHashMap;
import java.util.stream.Collectors;

@Repository
public class TransferHistoryRepository {

    private final Map<String, TransferHistory> histories = new ConcurrentHashMap<>();

    public TransferHistory save(TransferHistory history) {
        histories.put(history.getId(), history);
        return history;
    }

    public Optional<TransferHistory> findById(String id) {
        return Optional.ofNullable(histories.get(id));
    }

    public List<TransferHistory> findAll() {
        return new ArrayList<>(histories.values());
    }

    public List<TransferHistory> findByEmployeeId(String employeeId) {
        return histories.values().stream()
                .filter(h -> employeeId.equals(h.getEmployeeId()))
                .sorted(Comparator.comparing(TransferHistory::getTransferTime).reversed())
                .collect(Collectors.toList());
    }

    public List<TransferHistory> findByFromDepartmentId(String departmentId) {
        return histories.values().stream()
                .filter(h -> departmentId.equals(h.getFromDepartmentId()))
                .sorted(Comparator.comparing(TransferHistory::getTransferTime).reversed())
                .collect(Collectors.toList());
    }

    public List<TransferHistory> findByToDepartmentId(String departmentId) {
        return histories.values().stream()
                .filter(h -> departmentId.equals(h.getToDepartmentId()))
                .sorted(Comparator.comparing(TransferHistory::getTransferTime).reversed())
                .collect(Collectors.toList());
    }
}

package com.company.leave.repository;

import com.company.leave.entity.LeaveRecord;
import com.company.leave.enums.LeaveStatus;
import org.springframework.stereotype.Repository;

import java.time.LocalDate;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.Optional;
import java.util.stream.Collectors;

@Repository
public class LeaveRecordRepository {
    private final Map<Long, LeaveRecord> leaveRecords = new HashMap<>();
    private Long nextId = 1L;

    public LeaveRecord save(LeaveRecord record) {
        if (record.getId() == null) {
            record.setId(nextId++);
        }
        leaveRecords.put(record.getId(), record);
        return record;
    }

    public Optional<LeaveRecord> findById(Long id) {
        return Optional.ofNullable(leaveRecords.get(id));
    }

    public List<LeaveRecord> findAll() {
        return List.copyOf(leaveRecords.values());
    }

    public List<LeaveRecord> findByEmployeeId(Long employeeId) {
        return leaveRecords.values().stream()
                .filter(r -> employeeId.equals(r.getEmployeeId()))
                .collect(Collectors.toList());
    }

    public List<LeaveRecord> findByEmployeeIdAndStatus(Long employeeId, LeaveStatus status) {
        return leaveRecords.values().stream()
                .filter(r -> employeeId.equals(r.getEmployeeId()) && status.equals(r.getStatus()))
                .collect(Collectors.toList());
    }

    public List<LeaveRecord> findOverlappingByEmployeeId(Long employeeId, LocalDate startDate, LocalDate endDate) {
        return leaveRecords.values().stream()
                .filter(r -> employeeId.equals(r.getEmployeeId()))
                .filter(r -> LeaveStatus.REJECTED != r.getStatus())
                .filter(r -> isOverlapping(r.getStartDate(), r.getEndDate(), startDate, endDate))
                .collect(Collectors.toList());
    }

    public List<LeaveRecord> findByDateRange(LocalDate startDate, LocalDate endDate) {
        return leaveRecords.values().stream()
                .filter(r -> LeaveStatus.REJECTED != r.getStatus())
                .filter(r -> isOverlapping(r.getStartDate(), r.getEndDate(), startDate, endDate))
                .collect(Collectors.toList());
    }

    private boolean isOverlapping(LocalDate start1, LocalDate end1, LocalDate start2, LocalDate end2) {
        return !(start1.isAfter(end2) || end1.isBefore(start2));
    }
}

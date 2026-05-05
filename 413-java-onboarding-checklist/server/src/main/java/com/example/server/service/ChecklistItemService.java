package com.example.server.service;

import com.example.common.dto.ChecklistItemDTO;
import com.example.common.dto.EmployeeDTO;
import com.example.common.enums.ItemStatus;
import org.springframework.stereotype.Service;
import java.time.LocalDate;
import java.time.LocalDateTime;
import java.util.List;
import java.util.Map;
import java.util.Optional;
import java.util.stream.Collectors;

@Service
public class ChecklistItemService {

    public Optional<ChecklistItemDTO> findItemById(EmployeeDTO employee, String itemId) {
        return employee.getChecklistItems().stream()
                .filter(item -> itemId.equals(item.getId()))
                .findFirst();
    }

    public void markAsCompleted(ChecklistItemDTO item, boolean completed) {
        if (completed) {
            item.setStatus(ItemStatus.COMPLETED);
            item.setCompletedDate(LocalDate.now());
        } else {
            item.setStatus(ItemStatus.PENDING);
            item.setCompletedDate(null);
        }
        item.setUpdatedAt(LocalDateTime.now());
    }

    public void checkAndUpdateOverdueStatus(ChecklistItemDTO item, LocalDate today) {
        if (item.getStatus() == ItemStatus.COMPLETED || item.getStatus() == ItemStatus.ESCALATED) {
            return;
        }
        
        if (item.getDueDate() != null && item.getDueDate().isBefore(today)) {
            item.setStatus(ItemStatus.OVERDUE);
            item.setUpdatedAt(LocalDateTime.now());
        }
    }

    public void escalateToManager(ChecklistItemDTO item, String managerName) {
        item.setStatus(ItemStatus.ESCALATED);
        item.setEscalated(true);
        item.setEscalatedTo(managerName);
        item.setUpdatedAt(LocalDateTime.now());
    }

    public Map<String, Long> countOverdueItemsByResponsible(List<EmployeeDTO> employees) {
        return employees.stream()
                .flatMap(employee -> employee.getChecklistItems().stream())
                .filter(item -> item.getStatus() == ItemStatus.OVERDUE)
                .collect(Collectors.groupingBy(
                        ChecklistItemDTO::getResponsiblePerson,
                        Collectors.counting()
                ));
    }

    public long countItemsByStatus(List<EmployeeDTO> employees, ItemStatus status) {
        return employees.stream()
                .flatMap(employee -> employee.getChecklistItems().stream())
                .filter(item -> item.getStatus() == status)
                .count();
    }

    public List<ChecklistItemDTO> getItemsDueInDays(EmployeeDTO employee, int days) {
        LocalDate today = LocalDate.now();
        LocalDate targetDate = today.plusDays(days);
        return employee.getChecklistItems().stream()
                .filter(item -> item.getStatus() != ItemStatus.COMPLETED)
                .filter(item -> item.getDueDate() != null)
                .filter(item -> !item.getDueDate().isBefore(today) && 
                                !item.getDueDate().isAfter(targetDate))
                .collect(Collectors.toList());
    }

    public List<ChecklistItemDTO> getOnboardingDayRequiredItems(EmployeeDTO employee) {
        return employee.getChecklistItems().stream()
                .filter(ChecklistItemDTO::isOnboardingDayRequired)
                .collect(Collectors.toList());
    }

    public List<ChecklistItemDTO> getPreOnboardingItems(EmployeeDTO employee) {
        return employee.getChecklistItems().stream()
                .filter(ChecklistItemDTO::isPreOnboarding)
                .collect(Collectors.toList());
    }
}

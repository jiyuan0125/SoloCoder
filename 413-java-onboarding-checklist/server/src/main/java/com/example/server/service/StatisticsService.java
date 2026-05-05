package com.example.server.service;

import com.example.common.dto.EmployeeDTO;
import com.example.common.dto.OnboardingStatisticsDTO;
import com.example.common.enums.ItemStatus;
import com.example.common.enums.OnboardingStatus;
import com.example.server.repository.EmployeeRepository;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Service;
import java.util.LinkedHashMap;
import java.util.List;
import java.util.Map;

@Service
public class StatisticsService {

    @Autowired
    private EmployeeRepository employeeRepository;

    @Autowired
    private ChecklistItemService checklistItemService;

    public OnboardingStatisticsDTO getStatistics() {
        OnboardingStatisticsDTO statistics = new OnboardingStatisticsDTO();
        List<EmployeeDTO> employees = employeeRepository.findAll();

        statistics.setTotalEmployees(employees.size());
        statistics.setCompletedOnboarding(employeeRepository.countByStatus(OnboardingStatus.COMPLETE));
        statistics.setIncompleteOnboarding(employeeRepository.countByStatus(OnboardingStatus.INCOMPLETE));

        if (!employees.isEmpty()) {
            double completionRate = (double) statistics.getCompletedOnboarding() / employees.size() * 100;
            statistics.setCompletionRate(Math.round(completionRate * 100.0) / 100.0);
        }

        Map<String, Long> itemsByStatus = new LinkedHashMap<>();
        itemsByStatus.put("待处理", checklistItemService.countItemsByStatus(employees, ItemStatus.PENDING));
        itemsByStatus.put("进行中", checklistItemService.countItemsByStatus(employees, ItemStatus.IN_PROGRESS));
        itemsByStatus.put("已完成", checklistItemService.countItemsByStatus(employees, ItemStatus.COMPLETED));
        itemsByStatus.put("已逾期", checklistItemService.countItemsByStatus(employees, ItemStatus.OVERDUE));
        itemsByStatus.put("已升级", checklistItemService.countItemsByStatus(employees, ItemStatus.ESCALATED));
        statistics.setItemsByStatus(itemsByStatus);

        Map<String, Long> overdueItems = checklistItemService.countOverdueItemsByResponsible(employees);
        statistics.setOverdueItemsByResponsible(overdueItems);

        Map<String, Long> employeesByPosition = new LinkedHashMap<>();
        for (EmployeeDTO employee : employees) {
            String positionDesc = employee.getPositionType() != null 
                    ? employee.getPositionType().getDescription() 
                    : "未知";
            employeesByPosition.merge(positionDesc, 1L, Long::sum);
        }
        statistics.setEmployeesByPosition(employeesByPosition);

        long completedEmployees = statistics.getCompletedOnboarding();
        if (completedEmployees > 0) {
            double totalDays = employees.stream()
                    .filter(e -> e.getStatus() == OnboardingStatus.COMPLETE)
                    .mapToLong(e -> calculateCompletionDays(e))
                    .sum();
            double avgDays = totalDays / completedEmployees;
            statistics.setAverageCompletionTime(Math.round(avgDays * 100.0) / 100.0);
        }

        return statistics;
    }

    private long calculateCompletionDays(EmployeeDTO employee) {
        if (employee.getOnboardingDate() == null || employee.getChecklistItems() == null) {
            return 0;
        }

        return employee.getChecklistItems().stream()
                .filter(item -> item.getStatus() == ItemStatus.COMPLETED && item.getCompletedDate() != null)
                .mapToLong(item -> java.time.temporal.ChronoUnit.DAYS.between(
                        employee.getOnboardingDate(), item.getCompletedDate()))
                .max()
                .orElse(0L);
    }
}

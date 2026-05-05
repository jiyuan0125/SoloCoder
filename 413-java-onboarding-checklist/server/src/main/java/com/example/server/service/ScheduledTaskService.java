package com.example.server.service;

import com.example.common.dto.ChecklistItemDTO;
import com.example.common.dto.EmployeeDTO;
import com.example.common.enums.AlertLevel;
import com.example.common.enums.ItemStatus;
import com.example.common.enums.OnboardingStatus;
import com.example.server.repository.EmployeeRepository;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.scheduling.annotation.Scheduled;
import org.springframework.stereotype.Service;
import java.time.LocalDate;
import java.time.temporal.ChronoUnit;
import java.util.List;
import java.util.Map;

@Service
public class ScheduledTaskService {

    @Autowired
    private EmployeeRepository employeeRepository;

    @Autowired
    private ChecklistItemService checklistItemService;

    @Autowired
    private AlertService alertService;

    @Autowired
    private EmployeeService employeeService;

    @Scheduled(cron = "0 0 9 * * ?")
    public void runDailyTasks() {
        LocalDate today = LocalDate.now();
        
        checkPreOnboardingReminders(today);
        checkOverdueItems(today);
        checkOnboardingDayRequiredItems(today);
        checkEmployeeOnboardingStatus();
        checkResponsiblePersonOverdueThreshold();
    }

    private void checkPreOnboardingReminders(LocalDate today) {
        List<EmployeeDTO> employees = employeeRepository.findAll();
        for (EmployeeDTO employee : employees) {
            if (employee.getOnboardingDate() == null) continue;
            
            long daysUntilOnboarding = ChronoUnit.DAYS.between(today, employee.getOnboardingDate());
            if (daysUntilOnboarding == 3) {
                List<ChecklistItemDTO> preOnboardingItems = checklistItemService.getPreOnboardingItems(employee);
                for (ChecklistItemDTO item : preOnboardingItems) {
                    if (item.getStatus() != ItemStatus.COMPLETED) {
                        alertService.createAlert(
                                employee.getId(),
                                employee.getName(),
                                item.getResponsiblePerson(),
                                "入职前3天提醒：请检查并完成入职前准备事项 - " + item.getName(),
                                AlertLevel.INFO,
                                0
                        );
                    }
                }
            }
        }
    }

    private void checkOverdueItems(LocalDate today) {
        List<EmployeeDTO> employees = employeeRepository.findAll();
        for (EmployeeDTO employee : employees) {
            for (ChecklistItemDTO item : employee.getChecklistItems()) {
                checklistItemService.checkAndUpdateOverdueStatus(item, today);
            }
            employeeRepository.save(employee);
        }
    }

    private void checkOnboardingDayRequiredItems(LocalDate today) {
        List<EmployeeDTO> employees = employeeRepository.findAll();
        for (EmployeeDTO employee : employees) {
            if (employee.getOnboardingDate() == null) continue;
            
            long daysSinceOnboarding = ChronoUnit.DAYS.between(employee.getOnboardingDate(), today);
            if (daysSinceOnboarding == 1) {
                List<ChecklistItemDTO> onboardingDayItems = checklistItemService.getOnboardingDayRequiredItems(employee);
                for (ChecklistItemDTO item : onboardingDayItems) {
                    if (item.getStatus() != ItemStatus.COMPLETED && !item.isEscalated()) {
                        checklistItemService.escalateToManager(item, "部门经理");
                        
                        alertService.createAlert(
                                employee.getId(),
                                employee.getName(),
                                "部门经理",
                                "紧急！员工 " + employee.getName() + " 入职当天必须完成的事项未完成，已升级关注：" + item.getName(),
                                AlertLevel.CRITICAL,
                                1
                        );
                    }
                }
                employeeRepository.save(employee);
            }
        }
    }

    private void checkEmployeeOnboardingStatus() {
        List<EmployeeDTO> employees = employeeRepository.findAll();
        for (EmployeeDTO employee : employees) {
            employeeService.updateEmployeeStatus(employee);
        }
    }

    private void checkResponsiblePersonOverdueThreshold() {
        List<EmployeeDTO> employees = employeeRepository.findAll();
        Map<String, Long> overdueCountByResponsible = checklistItemService.countOverdueItemsByResponsible(employees);
        
        for (Map.Entry<String, Long> entry : overdueCountByResponsible.entrySet()) {
            String responsiblePerson = entry.getKey();
            long overdueCount = entry.getValue();
            
            if (overdueCount > 5) {
                alertService.createAlert(
                        null,
                        null,
                        responsiblePerson,
                        "负责人 " + responsiblePerson + " 名下逾期未完成事项已超过5个（当前：" + overdueCount + "个），请HR跟进处理",
                        AlertLevel.WARNING,
                        (int) overdueCount
                );
            }
        }
    }
}

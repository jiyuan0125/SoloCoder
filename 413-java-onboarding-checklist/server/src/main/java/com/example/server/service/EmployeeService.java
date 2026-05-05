package com.example.server.service;

import com.example.common.dto.ChecklistItemDTO;
import com.example.common.dto.EmployeeDTO;
import com.example.common.enums.ItemStatus;
import com.example.common.enums.OnboardingStatus;
import com.example.common.request.CreateEmployeeRequest;
import com.example.server.repository.EmployeeRepository;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Service;
import java.time.LocalDate;
import java.util.List;
import java.util.Optional;

@Service
public class EmployeeService {

    @Autowired
    private EmployeeRepository employeeRepository;

    @Autowired
    private TemplateService templateService;

    @Autowired
    private ChecklistItemService checklistItemService;

    public EmployeeDTO createEmployee(CreateEmployeeRequest request) {
        EmployeeDTO employee = new EmployeeDTO();
        employee.setName(request.getName());
        employee.setEmail(request.getEmail());
        employee.setPositionType(request.getPositionType());
        employee.setDepartment(request.getDepartment());
        employee.setOnboardingDate(request.getOnboardingDate());
        employee.setStatus(OnboardingStatus.IN_PROGRESS);

        List<ChecklistItemDTO> checklistItems = templateService.generateChecklistFromTemplate(
                request.getPositionType(), request.getOnboardingDate());
        employee.setChecklistItems(checklistItems);

        return employeeRepository.save(employee);
    }

    public Optional<EmployeeDTO> getEmployeeById(String id) {
        return employeeRepository.findById(id);
    }

    public List<EmployeeDTO> getAllEmployees() {
        return employeeRepository.findAll();
    }

    public boolean deleteEmployee(String id) {
        if (employeeRepository.existsById(id)) {
            employeeRepository.deleteById(id);
            return true;
        }
        return false;
    }

    public void updateEmployeeStatus(EmployeeDTO employee) {
        List<ChecklistItemDTO> items = employee.getChecklistItems();
        if (items.isEmpty()) {
            employee.setStatus(OnboardingStatus.COMPLETE);
            employeeRepository.save(employee);
            return;
        }

        long completedCount = items.stream()
                .filter(item -> item.getStatus() == ItemStatus.COMPLETED)
                .count();
        double completionRate = (double) completedCount / items.size();

        if (completionRate >= 0.8) {
            if (completionRate == 1.0) {
                employee.setStatus(OnboardingStatus.COMPLETE);
            } else {
                employee.setStatus(OnboardingStatus.IN_PROGRESS);
            }
        } else {
            employee.setStatus(OnboardingStatus.INCOMPLETE);
        }

        employeeRepository.save(employee);
    }

    public Optional<ChecklistItemDTO> addChecklistItem(String employeeId, ChecklistItemDTO item) {
        Optional<EmployeeDTO> employeeOpt = employeeRepository.findById(employeeId);
        if (employeeOpt.isPresent()) {
            EmployeeDTO employee = employeeOpt.get();
            employee.getChecklistItems().add(item);
            updateEmployeeStatus(employee);
            return Optional.of(item);
        }
        return Optional.empty();
    }

    public Optional<ChecklistItemDTO> updateChecklistItem(String employeeId, String itemId, 
            String name, String description, String responsiblePerson, 
            String department, LocalDate dueDate, Boolean isCompleted) {
        Optional<EmployeeDTO> employeeOpt = employeeRepository.findById(employeeId);
        if (employeeOpt.isPresent()) {
            EmployeeDTO employee = employeeOpt.get();
            Optional<ChecklistItemDTO> itemOpt = checklistItemService.findItemById(employee, itemId);
            if (itemOpt.isPresent()) {
                ChecklistItemDTO item = itemOpt.get();
                if (name != null) item.setName(name);
                if (description != null) item.setDescription(description);
                if (responsiblePerson != null) item.setResponsiblePerson(responsiblePerson);
                if (department != null) item.setDepartment(department);
                if (dueDate != null) item.setDueDate(dueDate);
                if (isCompleted != null) {
                    checklistItemService.markAsCompleted(item, isCompleted);
                }
                updateEmployeeStatus(employee);
                return Optional.of(item);
            }
        }
        return Optional.empty();
    }
}

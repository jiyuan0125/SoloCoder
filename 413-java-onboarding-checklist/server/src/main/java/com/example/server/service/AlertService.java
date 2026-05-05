package com.example.server.service;

import com.example.common.dto.AlertDTO;
import com.example.common.enums.AlertLevel;
import com.example.server.repository.AlertRepository;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Service;
import java.util.List;
import java.util.Optional;

@Service
public class AlertService {

    @Autowired
    private AlertRepository alertRepository;

    public AlertDTO createAlert(String employeeId, String employeeName, 
            String responsiblePerson, String message, AlertLevel level, int overdueCount) {
        AlertDTO alert = new AlertDTO();
        alert.setEmployeeId(employeeId);
        alert.setEmployeeName(employeeName);
        alert.setResponsiblePerson(responsiblePerson);
        alert.setMessage(message);
        alert.setLevel(level);
        alert.setOverdueCount(overdueCount);
        return alertRepository.save(alert);
    }

    public Optional<AlertDTO> getAlertById(String id) {
        return alertRepository.findById(id);
    }

    public List<AlertDTO> getAllAlerts() {
        return alertRepository.findAll();
    }

    public List<AlertDTO> getUnreadAlerts() {
        return alertRepository.findUnread();
    }

    public long getUnreadAlertCount() {
        return alertRepository.countUnread();
    }

    public boolean markAsRead(String id) {
        Optional<AlertDTO> alertOpt = alertRepository.findById(id);
        if (alertOpt.isPresent()) {
            AlertDTO alert = alertOpt.get();
            alert.setRead(true);
            alertRepository.save(alert);
            return true;
        }
        return false;
    }

    public boolean deleteAlert(String id) {
        if (alertRepository.existsById(id)) {
            alertRepository.deleteById(id);
            return true;
        }
        return false;
    }
}

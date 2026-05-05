package com.workshift.server.repository;

import com.workshift.server.entity.EmployeeEntity;
import com.workshift.server.entity.ObjectionEntity;
import com.workshift.server.entity.ShiftEntity;
import com.workshift.server.entity.ShiftSwapEntity;
import java.time.LocalDate;
import java.util.ArrayList;
import java.util.List;
import java.util.Map;
import java.util.UUID;
import java.util.concurrent.ConcurrentHashMap;
import java.util.stream.Collectors;
import org.springframework.stereotype.Component;

@Component
public class DataStore {
    private final Map<String, EmployeeEntity> employees = new ConcurrentHashMap<>();
    private final Map<String, ShiftEntity> shifts = new ConcurrentHashMap<>();
    private final Map<String, ShiftSwapEntity> swaps = new ConcurrentHashMap<>();
    private final Map<String, ObjectionEntity> objections = new ConcurrentHashMap<>();
    private final Map<LocalDate, String> holidays = new ConcurrentHashMap<>();

    public String generateId() {
        return UUID.randomUUID().toString();
    }

    public void saveEmployee(EmployeeEntity employee) {
        employees.put(employee.getId(), employee);
    }

    public EmployeeEntity getEmployee(String id) {
        return employees.get(id);
    }

    public List<EmployeeEntity> getAllEmployees() {
        return new ArrayList<>(employees.values());
    }

    public void removeEmployee(String id) {
        employees.remove(id);
    }

    public void saveShift(ShiftEntity shift) {
        shifts.put(shift.getId(), shift);
    }

    public ShiftEntity getShift(String id) {
        return shifts.get(id);
    }

    public List<ShiftEntity> getAllShifts() {
        return new ArrayList<>(shifts.values());
    }

    public List<ShiftEntity> getShiftsByEmployee(String employeeId) {
        return shifts.values().stream()
                .filter(s -> employeeId.equals(s.getEmployeeId()))
                .sorted((a, b) -> a.getDate().compareTo(b.getDate()))
                .collect(Collectors.toList());
    }

    public List<ShiftEntity> getShiftsByEmployeeAndDateRange(String employeeId, LocalDate startDate, LocalDate endDate) {
        return shifts.values().stream()
                .filter(s -> employeeId.equals(s.getEmployeeId()))
                .filter(s -> !s.getDate().isBefore(startDate) && !s.getDate().isAfter(endDate))
                .sorted((a, b) -> a.getDate().compareTo(b.getDate()))
                .collect(Collectors.toList());
    }

    public List<ShiftEntity> getShiftsByDateRange(LocalDate startDate, LocalDate endDate) {
        return shifts.values().stream()
                .filter(s -> !s.getDate().isBefore(startDate) && !s.getDate().isAfter(endDate))
                .sorted((a, b) -> a.getDate().compareTo(b.getDate()))
                .collect(Collectors.toList());
    }

    public ShiftEntity getShiftByEmployeeAndDate(String employeeId, LocalDate date) {
        return shifts.values().stream()
                .filter(s -> employeeId.equals(s.getEmployeeId()))
                .filter(s -> date.equals(s.getDate()))
                .findFirst()
                .orElse(null);
    }

    public void removeShift(String id) {
        shifts.remove(id);
    }

    public void saveSwap(ShiftSwapEntity swap) {
        swaps.put(swap.getId(), swap);
    }

    public ShiftSwapEntity getSwap(String id) {
        return swaps.get(id);
    }

    public List<ShiftSwapEntity> getAllSwaps() {
        return new ArrayList<>(swaps.values());
    }

    public List<ShiftSwapEntity> getSwapsByEmployee(String employeeId) {
        return swaps.values().stream()
                .filter(s -> employeeId.equals(s.getRequesterEmployeeId()) || 
                              employeeId.equals(s.getTargetEmployeeId()))
                .collect(Collectors.toList());
    }

    public void saveObjection(ObjectionEntity objection) {
        objections.put(objection.getId(), objection);
    }

    public ObjectionEntity getObjection(String id) {
        return objections.get(id);
    }

    public List<ObjectionEntity> getAllObjections() {
        return new ArrayList<>(objections.values());
    }

    public ObjectionEntity getObjectionByShiftAndEmployee(String shiftId, String employeeId) {
        return objections.values().stream()
                .filter(o -> shiftId.equals(o.getShiftId()))
                .filter(o -> employeeId.equals(o.getEmployeeId()))
                .findFirst()
                .orElse(null);
    }

    public void addHoliday(LocalDate date, String name) {
        holidays.put(date, name);
    }

    public boolean isHoliday(LocalDate date) {
        return holidays.containsKey(date);
    }

    public String getHolidayName(LocalDate date) {
        return holidays.get(date);
    }

    public void clearAll() {
        employees.clear();
        shifts.clear();
        swaps.clear();
        objections.clear();
        holidays.clear();
    }
}
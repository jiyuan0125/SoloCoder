package com.workshift.server.service;

import com.workshift.common.enums.ErrorCode;
import com.workshift.common.enums.ShiftType;
import com.workshift.server.entity.ShiftEntity;
import java.time.LocalDate;
import java.util.ArrayList;
import java.util.List;
import java.util.Optional;
import org.springframework.stereotype.Component;

@Component
public class ShiftRuleValidator {

    public Optional<ErrorCode> validateNewShift(String employeeId, LocalDate date, ShiftType newShiftType,
            List<ShiftEntity> existingShifts) {
        
        List<ShiftEntity> employeeShifts = filterByEmployee(existingShifts, employeeId);
        
        Optional<ErrorCode> result = checkShiftAlreadyExists(employeeShifts, date);
        if (result.isPresent()) {
            return result;
        }
        
        result = checkConsecutiveNightShifts(employeeShifts, date, newShiftType);
        if (result.isPresent()) {
            return result;
        }
        
        result = checkShiftInterval(employeeShifts, date, newShiftType);
        if (result.isPresent()) {
            return result;
        }
        
        result = checkNightToMorning(employeeShifts, date, newShiftType);
        if (result.isPresent()) {
            return result;
        }
        
        List<ShiftEntity> tempShifts = new ArrayList<>(employeeShifts);
        ShiftEntity tempShift = new ShiftEntity();
        tempShift.setEmployeeId(employeeId);
        tempShift.setDate(date);
        tempShift.setShiftType(newShiftType);
        tempShifts.add(tempShift);
        
        result = checkWeeklyRestDay(tempShifts, date);
        if (result.isPresent()) {
            return result;
        }
        
        result = checkConsecutiveWorkDays(tempShifts, date);
        if (result.isPresent()) {
            return result;
        }
        
        return Optional.empty();
    }

    public Optional<ErrorCode> validateSwapResult(
            String employeeId1, ShiftEntity shift1,
            String employeeId2, ShiftEntity shift2,
            List<ShiftEntity> allShifts) {
        
        List<ShiftEntity> emp1Shifts = filterByEmployee(allShifts, employeeId1);
        List<ShiftEntity> emp2Shifts = filterByEmployee(allShifts, employeeId2);
        
        List<ShiftEntity> emp1Simulated = simulateSwap(emp1Shifts, employeeId1, shift1, employeeId2, shift2);
        List<ShiftEntity> emp2Simulated = simulateSwap(emp2Shifts, employeeId2, shift2, employeeId1, shift1);
        
        Optional<ErrorCode> result = validateAfterSwap(emp1Simulated, employeeId1);
        if (result.isPresent()) {
            return result;
        }
        
        result = validateAfterSwap(emp2Simulated, employeeId2);
        if (result.isPresent()) {
            return result;
        }
        
        return Optional.empty();
    }

    private Optional<ErrorCode> validateAfterSwap(List<ShiftEntity> shifts, String employeeId) {
        Optional<ErrorCode> result = checkMaxConsecutiveNightShifts(shifts);
        if (result.isPresent()) {
            return result;
        }
        
        result = checkMaxConsecutiveWorkDays(shifts);
        if (result.isPresent()) {
            return result;
        }
        
        result = checkAllWeeklyRestDays(shifts);
        if (result.isPresent()) {
            return result;
        }
        
        result = checkAllShiftIntervals(shifts);
        if (result.isPresent()) {
            return result;
        }
        
        return Optional.empty();
    }

    private Optional<ErrorCode> checkShiftAlreadyExists(List<ShiftEntity> shifts, LocalDate date) {
        boolean exists = shifts.stream()
                .anyMatch(s -> s.getDate().equals(date));
        if (exists) {
            return Optional.of(ErrorCode.SHIFT_ALREADY_EXISTS);
        }
        return Optional.empty();
    }

    private Optional<ErrorCode> checkConsecutiveNightShifts(List<ShiftEntity> shifts, LocalDate newDate, ShiftType newShiftType) {
        if (newShiftType != ShiftType.NIGHT) {
            return Optional.empty();
        }
        
        int consecutiveNights = 0;
        LocalDate checkDate = newDate.minusDays(1);
        
        while (true) {
            LocalDate finalCheckDate = checkDate;
            Optional<ShiftEntity> shift = shifts.stream()
                    .filter(s -> s.getDate().equals(finalCheckDate))
                    .findFirst();
            
            if (shift.isPresent() && shift.get().getShiftType() == ShiftType.NIGHT) {
                consecutiveNights++;
                checkDate = checkDate.minusDays(1);
            } else {
                break;
            }
        }
        
        if (consecutiveNights >= 3) {
            return Optional.of(ErrorCode.CONSECUTIVE_NIGHT_SHIFT_VIOLATION);
        }
        return Optional.empty();
    }

    private Optional<ErrorCode> checkMaxConsecutiveNightShifts(List<ShiftEntity> shifts) {
        if (shifts.isEmpty()) {
            return Optional.empty();
        }
        
        List<ShiftEntity> sorted = new ArrayList<>(shifts);
        sorted.sort((a, b) -> a.getDate().compareTo(b.getDate()));
        
        int maxConsecutive = 0;
        int current = 0;
        
        for (ShiftEntity shift : sorted) {
            if (shift.getShiftType() == ShiftType.NIGHT) {
                current++;
                maxConsecutive = Math.max(maxConsecutive, current);
            } else {
                current = 0;
            }
        }
        
        if (maxConsecutive > 3) {
            return Optional.of(ErrorCode.CONSECUTIVE_NIGHT_SHIFT_VIOLATION);
        }
        return Optional.empty();
    }

    private Optional<ErrorCode> checkShiftInterval(List<ShiftEntity> shifts, LocalDate newDate, ShiftType newShiftType) {
        Optional<ErrorCode> prevCheck = checkPreviousDay(shifts, newDate, newShiftType);
        if (prevCheck.isPresent()) {
            return prevCheck;
        }
        
        Optional<ErrorCode> nextCheck = checkNextDay(shifts, newDate, newShiftType);
        if (nextCheck.isPresent()) {
            return nextCheck;
        }
        
        return Optional.empty();
    }

    private Optional<ErrorCode> checkPreviousDay(List<ShiftEntity> shifts, LocalDate newDate, ShiftType newShiftType) {
        LocalDate prevDate = newDate.minusDays(1);
        Optional<ShiftEntity> prevShift = shifts.stream()
                .filter(s -> s.getDate().equals(prevDate))
                .findFirst();
        
        if (prevShift.isPresent()) {
            int interval = calculateInterval(prevShift.get().getShiftType(), newShiftType);
            if (interval < 8) {
                return Optional.of(ErrorCode.SHIFT_INTERVAL_VIOLATION);
            }
        }
        return Optional.empty();
    }

    private Optional<ErrorCode> checkNextDay(List<ShiftEntity> shifts, LocalDate newDate, ShiftType newShiftType) {
        LocalDate nextDate = newDate.plusDays(1);
        Optional<ShiftEntity> nextShift = shifts.stream()
                .filter(s -> s.getDate().equals(nextDate))
                .findFirst();
        
        if (nextShift.isPresent()) {
            int interval = calculateInterval(newShiftType, nextShift.get().getShiftType());
            if (interval < 8) {
                return Optional.of(ErrorCode.SHIFT_INTERVAL_VIOLATION);
            }
        }
        return Optional.empty();
    }

    private Optional<ErrorCode> checkNightToMorning(List<ShiftEntity> shifts, LocalDate newDate, ShiftType newShiftType) {
        LocalDate prevDate = newDate.minusDays(1);
        Optional<ShiftEntity> prevShift = shifts.stream()
                .filter(s -> s.getDate().equals(prevDate))
                .findFirst();
        
        if (prevShift.isPresent() && 
            prevShift.get().getShiftType() == ShiftType.NIGHT && 
            newShiftType == ShiftType.MORNING) {
            return Optional.of(ErrorCode.NIGHT_TO_MORNING_VIOLATION);
        }
        return Optional.empty();
    }

    private Optional<ErrorCode> checkWeeklyRestDay(List<ShiftEntity> shifts, LocalDate newDate) {
        List<ShiftEntity> weekShifts = getShiftsForWeekContaining(shifts, newDate);
        
        int workDays = (int) weekShifts.stream()
                .filter(s -> s.getShiftType() != null)
                .count();
        
        int weekDays = 7;
        int restDays = weekDays - workDays;
        
        if (restDays < 1) {
            return Optional.of(ErrorCode.WEEKLY_REST_DAY_VIOLATION);
        }
        return Optional.empty();
    }

    private Optional<ErrorCode> checkAllWeeklyRestDays(List<ShiftEntity> shifts) {
        if (shifts.isEmpty()) {
            return Optional.empty();
        }
        
        List<ShiftEntity> sorted = new ArrayList<>(shifts);
        sorted.sort((a, b) -> a.getDate().compareTo(b.getDate()));
        
        LocalDate minDate = sorted.get(0).getDate();
        LocalDate maxDate = sorted.get(sorted.size() - 1).getDate();
        
        LocalDate weekStart = minDate.minusDays(minDate.getDayOfWeek().getValue() - 1);
        
        while (!weekStart.isAfter(maxDate)) {
            LocalDate currentWeekStart = weekStart;
            LocalDate weekEnd = weekStart.plusDays(6);
            
            List<ShiftEntity> weekShifts = sorted.stream()
                    .filter(s -> !s.getDate().isBefore(currentWeekStart) && !s.getDate().isAfter(weekEnd))
                    .toList();
            
            int workDays = weekShifts.size();
            if (workDays >= 7) {
                return Optional.of(ErrorCode.WEEKLY_REST_DAY_VIOLATION);
            }
            
            weekStart = weekStart.plusWeeks(1);
        }
        
        return Optional.empty();
    }

    private Optional<ErrorCode> checkConsecutiveWorkDays(List<ShiftEntity> shifts, LocalDate newDate) {
        List<ShiftEntity> sorted = new ArrayList<>(shifts);
        sorted.sort((a, b) -> a.getDate().compareTo(b.getDate()));
        
        int maxConsecutive = calculateMaxConsecutiveWorkDays(sorted);
        
        if (maxConsecutive > 6) {
            return Optional.of(ErrorCode.CONSECUTIVE_WORK_DAY_VIOLATION);
        }
        return Optional.empty();
    }

    private Optional<ErrorCode> checkMaxConsecutiveWorkDays(List<ShiftEntity> shifts) {
        if (shifts.isEmpty()) {
            return Optional.empty();
        }
        
        int maxConsecutive = calculateMaxConsecutiveWorkDays(shifts);
        
        if (maxConsecutive > 6) {
            return Optional.of(ErrorCode.CONSECUTIVE_WORK_DAY_VIOLATION);
        }
        return Optional.empty();
    }

    private Optional<ErrorCode> checkAllShiftIntervals(List<ShiftEntity> shifts) {
        if (shifts.size() < 2) {
            return Optional.empty();
        }
        
        List<ShiftEntity> sorted = new ArrayList<>(shifts);
        sorted.sort((a, b) -> a.getDate().compareTo(b.getDate()));
        
        for (int i = 0; i < sorted.size() - 1; i++) {
            ShiftEntity current = sorted.get(i);
            ShiftEntity next = sorted.get(i + 1);
            
            if (next.getDate().equals(current.getDate().plusDays(1))) {
                int interval = calculateInterval(current.getShiftType(), next.getShiftType());
                if (interval < 8) {
                    return Optional.of(ErrorCode.SHIFT_INTERVAL_VIOLATION);
                }
                
                if (current.getShiftType() == ShiftType.NIGHT && 
                    next.getShiftType() == ShiftType.MORNING) {
                    return Optional.of(ErrorCode.NIGHT_TO_MORNING_VIOLATION);
                }
            }
        }
        
        return Optional.empty();
    }

    private int calculateInterval(ShiftType previous, ShiftType next) {
        int prevEndHour = previous.getEndHour();
        int nextStartHour = next.getStartHour();
        
        return (24 - prevEndHour) + nextStartHour;
    }

    private List<ShiftEntity> getShiftsForWeekContaining(List<ShiftEntity> shifts, LocalDate date) {
        LocalDate weekStart = date.minusDays(date.getDayOfWeek().getValue() - 1);
        LocalDate weekEnd = weekStart.plusDays(6);
        
        return shifts.stream()
                .filter(s -> !s.getDate().isBefore(weekStart) && !s.getDate().isAfter(weekEnd))
                .toList();
    }

    private int calculateMaxConsecutiveWorkDays(List<ShiftEntity> shifts) {
        if (shifts.isEmpty()) {
            return 0;
        }
        
        List<ShiftEntity> sorted = new ArrayList<>(shifts);
        sorted.sort((a, b) -> a.getDate().compareTo(b.getDate()));
        
        int maxConsecutive = 0;
        int current = 1;
        
        for (int i = 1; i < sorted.size(); i++) {
            LocalDate currentDate = sorted.get(i).getDate();
            LocalDate prevDate = sorted.get(i - 1).getDate();
            
            if (currentDate.equals(prevDate.plusDays(1))) {
                current++;
                maxConsecutive = Math.max(maxConsecutive, current);
            } else {
                current = 1;
            }
        }
        
        return Math.max(maxConsecutive, current);
    }

    private List<ShiftEntity> filterByEmployee(List<ShiftEntity> shifts, String employeeId) {
        return shifts.stream()
                .filter(s -> employeeId.equals(s.getEmployeeId()))
                .toList();
    }

    private List<ShiftEntity> simulateSwap(List<ShiftEntity> originalShifts, 
            String currentEmployeeId, ShiftEntity currentShift,
            String otherEmployeeId, ShiftEntity otherShift) {
        
        List<ShiftEntity> result = new ArrayList<>();
        
        for (ShiftEntity shift : originalShifts) {
            if (shift.getDate().equals(currentShift.getDate())) {
                ShiftEntity newShift = new ShiftEntity();
                newShift.setId(shift.getId());
                newShift.setEmployeeId(currentEmployeeId);
                newShift.setDate(currentShift.getDate());
                newShift.setShiftType(otherShift.getShiftType());
                newShift.setPublished(shift.isPublished());
                newShift.setPublishTime(shift.getPublishTime());
                newShift.setHoliday(shift.isHoliday());
                newShift.setWorkHours(shift.getWorkHours());
                result.add(newShift);
            } else {
                result.add(shift);
            }
        }
        
        return result;
    }
}
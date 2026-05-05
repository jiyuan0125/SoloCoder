package com.workshift.server.service;

import com.workshift.common.dto.MonthlyStatisticsDTO;
import com.workshift.common.enums.ErrorCode;
import com.workshift.common.enums.ShiftType;
import com.workshift.server.entity.EmployeeEntity;
import com.workshift.server.entity.ShiftEntity;
import com.workshift.server.repository.DataStore;
import java.math.BigDecimal;
import java.math.RoundingMode;
import java.time.LocalDate;
import java.time.YearMonth;
import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.Optional;
import org.springframework.stereotype.Service;

@Service
public class StatisticsService {

    private static final int STANDARD_WORK_HOURS = 8;
    private static final BigDecimal HOLIDAY_OVERTIME_MULTIPLIER = BigDecimal.valueOf(2);

    private final DataStore dataStore;
    private final SwapService swapService;
    private final ObjectionService objectionService;

    public StatisticsService(DataStore dataStore, SwapService swapService, ObjectionService objectionService) {
        this.dataStore = dataStore;
        this.swapService = swapService;
        this.objectionService = objectionService;
    }

    public MonthlyStatisticsDTO generateMonthlyStatistics(String employeeId, YearMonth yearMonth) {
        EmployeeEntity employee = dataStore.getEmployee(employeeId);
        if (employee == null) {
            return null;
        }

        LocalDate startDate = yearMonth.atDay(1);
        LocalDate endDate = yearMonth.atEndOfMonth();

        List<ShiftEntity> shifts = dataStore.getShiftsByEmployeeAndDateRange(employeeId, startDate, endDate);

        MonthlyStatisticsDTO stats = new MonthlyStatisticsDTO();
        stats.setEmployeeId(employeeId);
        stats.setYearMonth(yearMonth);

        int totalDays = endDate.getDayOfMonth();
        int workingDays = 0;
        int restDays = 0;
        int morningCount = 0;
        int afternoonCount = 0;
        int nightCount = 0;
        int holidayWorkingDays = 0;
        double normalHours = 0.0;
        double holidayHours = 0.0;

        for (ShiftEntity shift : shifts) {
            workingDays++;
            if (shift.getShiftType() == ShiftType.MORNING) {
                morningCount++;
            } else if (shift.getShiftType() == ShiftType.AFTERNOON) {
                afternoonCount++;
            } else if (shift.getShiftType() == ShiftType.NIGHT) {
                nightCount++;
            }

            if (shift.isHoliday()) {
                holidayWorkingDays++;
                holidayHours += shift.getWorkHours();
            } else {
                normalHours += shift.getWorkHours();
            }
        }

        restDays = totalDays - workingDays;

        stats.setTotalWorkingDays(workingDays);
        stats.setTotalRestDays(restDays);
        stats.setMorningShiftCount(morningCount);
        stats.setAfternoonShiftCount(afternoonCount);
        stats.setNightShiftCount(nightCount);
        stats.setHolidayWorkingDays(holidayWorkingDays);
        stats.setNormalWorkHours(normalHours);
        stats.setHolidayWorkHours(holidayHours);
        stats.setTotalWorkHours(normalHours + holidayHours);

        int maxConsecutiveNights = calculateMaxConsecutiveNightShifts(shifts);
        int maxConsecutiveWorkDays = calculateMaxConsecutiveWorkDays(shifts);
        stats.setConsecutiveNightShiftMax(maxConsecutiveNights);
        stats.setConsecutiveWorkDaysMax(maxConsecutiveWorkDays);

        BigDecimal normalWage = BigDecimal.valueOf(normalHours)
                .multiply(employee.getHourlyWage())
                .setScale(2, RoundingMode.HALF_UP);

        BigDecimal holidayOvertimeWage = BigDecimal.valueOf(holidayHours)
                .multiply(employee.getHourlyWage())
                .multiply(HOLIDAY_OVERTIME_MULTIPLIER)
                .setScale(2, RoundingMode.HALF_UP);

        stats.setNormalWage(normalWage);
        stats.setHolidayOvertimeWage(holidayOvertimeWage);
        stats.setTotalWage(normalWage.add(holidayOvertimeWage));

        int swapCount = swapService.getSwapCountByEmployeeAndMonth(employeeId, 
                yearMonth.getYear(), yearMonth.getMonthValue());
        stats.setSwapCount(swapCount);

        int objectionCount = objectionService.getObjectionCountByEmployeeAndMonth(employeeId,
                yearMonth.getYear(), yearMonth.getMonthValue());
        stats.setObjectionCount(objectionCount);

        Map<String, Object> violations = new HashMap<>();
        boolean compliant = true;

        if (maxConsecutiveNights > 3) {
            violations.put("consecutiveNightShifts", 
                    "连续夜班超过3天: " + maxConsecutiveNights + "天");
            compliant = false;
        }

        if (maxConsecutiveWorkDays > 6) {
            violations.put("consecutiveWorkDays", 
                    "连续工作超过6天: " + maxConsecutiveWorkDays + "天");
            compliant = false;
        }

        List<String> shiftIntervalViolations = checkShiftIntervalViolations(shifts);
        if (!shiftIntervalViolations.isEmpty()) {
            violations.put("shiftInterval", shiftIntervalViolations);
            compliant = false;
        }

        stats.setRuleCompliant(compliant);
        stats.setViolations(violations);

        return stats;
    }

    public List<MonthlyStatisticsDTO> generateAllMonthlyStatistics(YearMonth yearMonth) {
        List<EmployeeEntity> employees = dataStore.getAllEmployees();
        List<MonthlyStatisticsDTO> statsList = new ArrayList<>();

        for (EmployeeEntity employee : employees) {
            MonthlyStatisticsDTO stats = generateMonthlyStatistics(employee.getId(), yearMonth);
            if (stats != null) {
                statsList.add(stats);
            }
        }

        return statsList;
    }

    public Optional<ErrorCode> addHoliday(LocalDate date, String name) {
        dataStore.addHoliday(date, name);
        return Optional.empty();
    }

    public boolean isHoliday(LocalDate date) {
        return dataStore.isHoliday(date);
    }

    private int calculateMaxConsecutiveNightShifts(List<ShiftEntity> shifts) {
        if (shifts.isEmpty()) {
            return 0;
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

        return maxConsecutive;
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

    private List<String> checkShiftIntervalViolations(List<ShiftEntity> shifts) {
        List<String> violations = new ArrayList<>();

        if (shifts.size() < 2) {
            return violations;
        }

        List<ShiftEntity> sorted = new ArrayList<>(shifts);
        sorted.sort((a, b) -> a.getDate().compareTo(b.getDate()));

        for (int i = 0; i < sorted.size() - 1; i++) {
            ShiftEntity current = sorted.get(i);
            ShiftEntity next = sorted.get(i + 1);

            if (next.getDate().equals(current.getDate().plusDays(1))) {
                int interval = calculateInterval(current.getShiftType(), next.getShiftType());
                if (interval < 8) {
                    violations.add(String.format("班次间隔不足8小时: %s(%s) -> %s(%s), 间隔%d小时",
                            current.getDate(), current.getShiftType().getDescription(),
                            next.getDate(), next.getShiftType().getDescription(), interval));
                }

                if (current.getShiftType() == ShiftType.NIGHT && 
                    next.getShiftType() == ShiftType.MORNING) {
                    violations.add(String.format("夜班接早班违规: %s(夜班) -> %s(早班)",
                            current.getDate(), next.getDate()));
                }
            }
        }

        return violations;
    }

    private int calculateInterval(ShiftType previous, ShiftType next) {
        int prevEndHour = previous.getEndHour();
        int nextStartHour = next.getStartHour();

        if (prevEndHour == 24) {
            prevEndHour = 0;
        }

        if (prevEndHour < nextStartHour) {
            return nextStartHour - prevEndHour;
        } else {
            return (24 - prevEndHour) + nextStartHour;
        }
    }
}
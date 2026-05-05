package com.workshift.common.dto;

import java.math.BigDecimal;
import java.time.YearMonth;
import java.util.Map;

public class MonthlyStatisticsDTO {
    private String employeeId;
    private YearMonth yearMonth;
    private int totalWorkingDays;
    private int totalRestDays;
    private int morningShiftCount;
    private int afternoonShiftCount;
    private int nightShiftCount;
    private int consecutiveNightShiftMax;
    private int consecutiveWorkDaysMax;
    private int holidayWorkingDays;
    private double normalWorkHours;
    private double holidayWorkHours;
    private double totalWorkHours;
    private BigDecimal normalWage;
    private BigDecimal holidayOvertimeWage;
    private BigDecimal totalWage;
    private int swapCount;
    private int objectionCount;
    private boolean ruleCompliant;
    private Map<String, Object> violations;

    public MonthlyStatisticsDTO() {
    }

    public String getEmployeeId() {
        return employeeId;
    }

    public void setEmployeeId(String employeeId) {
        this.employeeId = employeeId;
    }

    public YearMonth getYearMonth() {
        return yearMonth;
    }

    public void setYearMonth(YearMonth yearMonth) {
        this.yearMonth = yearMonth;
    }

    public int getTotalWorkingDays() {
        return totalWorkingDays;
    }

    public void setTotalWorkingDays(int totalWorkingDays) {
        this.totalWorkingDays = totalWorkingDays;
    }

    public int getTotalRestDays() {
        return totalRestDays;
    }

    public void setTotalRestDays(int totalRestDays) {
        this.totalRestDays = totalRestDays;
    }

    public int getMorningShiftCount() {
        return morningShiftCount;
    }

    public void setMorningShiftCount(int morningShiftCount) {
        this.morningShiftCount = morningShiftCount;
    }

    public int getAfternoonShiftCount() {
        return afternoonShiftCount;
    }

    public void setAfternoonShiftCount(int afternoonShiftCount) {
        this.afternoonShiftCount = afternoonShiftCount;
    }

    public int getNightShiftCount() {
        return nightShiftCount;
    }

    public void setNightShiftCount(int nightShiftCount) {
        this.nightShiftCount = nightShiftCount;
    }

    public int getConsecutiveNightShiftMax() {
        return consecutiveNightShiftMax;
    }

    public void setConsecutiveNightShiftMax(int consecutiveNightShiftMax) {
        this.consecutiveNightShiftMax = consecutiveNightShiftMax;
    }

    public int getConsecutiveWorkDaysMax() {
        return consecutiveWorkDaysMax;
    }

    public void setConsecutiveWorkDaysMax(int consecutiveWorkDaysMax) {
        this.consecutiveWorkDaysMax = consecutiveWorkDaysMax;
    }

    public int getHolidayWorkingDays() {
        return holidayWorkingDays;
    }

    public void setHolidayWorkingDays(int holidayWorkingDays) {
        this.holidayWorkingDays = holidayWorkingDays;
    }

    public double getNormalWorkHours() {
        return normalWorkHours;
    }

    public void setNormalWorkHours(double normalWorkHours) {
        this.normalWorkHours = normalWorkHours;
    }

    public double getHolidayWorkHours() {
        return holidayWorkHours;
    }

    public void setHolidayWorkHours(double holidayWorkHours) {
        this.holidayWorkHours = holidayWorkHours;
    }

    public double getTotalWorkHours() {
        return totalWorkHours;
    }

    public void setTotalWorkHours(double totalWorkHours) {
        this.totalWorkHours = totalWorkHours;
    }

    public BigDecimal getNormalWage() {
        return normalWage;
    }

    public void setNormalWage(BigDecimal normalWage) {
        this.normalWage = normalWage;
    }

    public BigDecimal getHolidayOvertimeWage() {
        return holidayOvertimeWage;
    }

    public void setHolidayOvertimeWage(BigDecimal holidayOvertimeWage) {
        this.holidayOvertimeWage = holidayOvertimeWage;
    }

    public BigDecimal getTotalWage() {
        return totalWage;
    }

    public void setTotalWage(BigDecimal totalWage) {
        this.totalWage = totalWage;
    }

    public int getSwapCount() {
        return swapCount;
    }

    public void setSwapCount(int swapCount) {
        this.swapCount = swapCount;
    }

    public int getObjectionCount() {
        return objectionCount;
    }

    public void setObjectionCount(int objectionCount) {
        this.objectionCount = objectionCount;
    }

    public boolean isRuleCompliant() {
        return ruleCompliant;
    }

    public void setRuleCompliant(boolean ruleCompliant) {
        this.ruleCompliant = ruleCompliant;
    }

    public Map<String, Object> getViolations() {
        return violations;
    }

    public void setViolations(Map<String, Object> violations) {
        this.violations = violations;
    }
}
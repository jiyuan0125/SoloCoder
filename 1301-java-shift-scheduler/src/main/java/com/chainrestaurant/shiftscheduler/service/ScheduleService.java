package com.chainrestaurant.shiftscheduler.service;

import com.chainrestaurant.shiftscheduler.dto.ComplianceViolation;
import com.chainrestaurant.shiftscheduler.dto.CreateScheduleRequest;
import com.chainrestaurant.shiftscheduler.dto.ScheduleViewDTO;
import com.chainrestaurant.shiftscheduler.entity.Employee;
import com.chainrestaurant.shiftscheduler.entity.Schedule;
import com.chainrestaurant.shiftscheduler.entity.Shift;
import com.chainrestaurant.shiftscheduler.exception.BusinessException;
import com.chainrestaurant.shiftscheduler.exception.ResourceNotFoundException;
import com.chainrestaurant.shiftscheduler.repository.EmployeeRepository;
import com.chainrestaurant.shiftscheduler.repository.ScheduleRepository;
import com.chainrestaurant.shiftscheduler.repository.ShiftRepository;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.time.LocalDate;
import java.util.ArrayList;
import java.util.Comparator;
import java.util.List;
import java.util.stream.Collectors;

@Service
@Transactional
public class ScheduleService {
    
    private static final int MAX_CONSECUTIVE_WORK_HOURS = 12;
    
    @Autowired
    private ScheduleRepository scheduleRepository;
    
    @Autowired
    private EmployeeRepository employeeRepository;
    
    @Autowired
    private ShiftRepository shiftRepository;
    
    public Schedule createSchedule(CreateScheduleRequest request) {
        validateCreateRequest(request);
        
        Employee employee = employeeRepository.findById(request.getEmployeeId())
                .orElseThrow(() -> new ResourceNotFoundException("员工不存在: " + request.getEmployeeId()));
        
        if (scheduleRepository.existsByEmployeeIdAndScheduleDate(request.getEmployeeId(), request.getScheduleDate())) {
            throw new BusinessException("员工 " + employee.getName() + " 在 " + request.getScheduleDate() + " 已有排班记录");
        }
        
        Schedule schedule = new Schedule();
        schedule.setEmployee(employee);
        schedule.setScheduleDate(request.getScheduleDate());
        schedule.setIsDayOff(request.getIsDayOff() != null ? request.getIsDayOff() : false);
        
        if (!schedule.getIsDayOff() && request.getShiftId() != null) {
            Shift shift = shiftRepository.findById(request.getShiftId())
                    .orElseThrow(() -> new ResourceNotFoundException("班次不存在: " + request.getShiftId()));
            schedule.setShift(shift);
        }
        
        Schedule savedSchedule = scheduleRepository.save(schedule);
        
        checkComplianceAfterSchedule(savedSchedule);
        
        return savedSchedule;
    }
    
    public Schedule updateSchedule(Long id, CreateScheduleRequest request) {
        Schedule schedule = scheduleRepository.findById(id)
                .orElseThrow(() -> new ResourceNotFoundException("排班记录不存在: " + id));
        
        validateCreateRequest(request);
        
        if (!schedule.getEmployee().getId().equals(request.getEmployeeId())) {
            Employee employee = employeeRepository.findById(request.getEmployeeId())
                    .orElseThrow(() -> new ResourceNotFoundException("员工不存在: " + request.getEmployeeId()));
            schedule.setEmployee(employee);
        }
        
        schedule.setScheduleDate(request.getScheduleDate());
        schedule.setIsDayOff(request.getIsDayOff() != null ? request.getIsDayOff() : false);
        
        if (schedule.getIsDayOff()) {
            schedule.setShift(null);
        } else if (request.getShiftId() != null) {
            Shift shift = shiftRepository.findById(request.getShiftId())
                    .orElseThrow(() -> new ResourceNotFoundException("班次不存在: " + request.getShiftId()));
            schedule.setShift(shift);
        }
        
        Schedule savedSchedule = scheduleRepository.save(schedule);
        
        checkComplianceAfterSchedule(savedSchedule);
        
        return savedSchedule;
    }
    
    public void deleteSchedule(Long id) {
        if (!scheduleRepository.existsById(id)) {
            throw new ResourceNotFoundException("排班记录不存在: " + id);
        }
        scheduleRepository.deleteById(id);
    }
    
    public Schedule getScheduleById(Long id) {
        return scheduleRepository.findById(id)
                .orElseThrow(() -> new ResourceNotFoundException("排班记录不存在: " + id));
    }
    
    public List<ScheduleViewDTO> getSchedulesByEmployee(Long employeeId, LocalDate startDate, LocalDate endDate) {
        List<Schedule> schedules = scheduleRepository.findByEmployeeIdAndDateRangeWithShift(employeeId, startDate, endDate);
        return schedules.stream()
                .map(this::toViewDTO)
                .sorted(Comparator.comparing(ScheduleViewDTO::getScheduleDate))
                .collect(Collectors.toList());
    }
    
    public List<ScheduleViewDTO> getSchedulesByStore(String storeName, LocalDate startDate, LocalDate endDate) {
        List<Schedule> schedules = scheduleRepository.findByStoreNameAndDateRange(storeName, startDate, endDate);
        return schedules.stream()
                .map(this::toViewDTO)
                .sorted(Comparator.comparing(ScheduleViewDTO::getScheduleDate)
                        .thenComparing(ScheduleViewDTO::getEmployeeName))
                .collect(Collectors.toList());
    }
    
    public List<ScheduleViewDTO> getSchedulesByDate(LocalDate date) {
        List<Schedule> schedules = scheduleRepository.findByScheduleDateBetweenOrderByScheduleDateAsc(date, date);
        return schedules.stream()
                .map(this::toViewDTO)
                .collect(Collectors.toList());
    }
    
    public List<ComplianceViolation> checkCompliance(Long employeeId, LocalDate startDate, LocalDate endDate) {
        if (!employeeRepository.existsById(employeeId)) {
            throw new ResourceNotFoundException("员工不存在: " + employeeId);
        }
        
        List<Schedule> schedules = scheduleRepository.findByEmployeeIdAndDateRangeWithShift(
                employeeId, startDate.minusDays(1), endDate.plusDays(1));
        
        return detectConsecutiveWorkViolations(schedules, startDate, endDate);
    }
    
    private void validateCreateRequest(CreateScheduleRequest request) {
        if (request.getEmployeeId() == null) {
            throw new BusinessException("员工ID不能为空");
        }
        if (request.getScheduleDate() == null) {
            throw new BusinessException("排班日期不能为空");
        }
        
        Boolean isDayOff = request.getIsDayOff() != null ? request.getIsDayOff() : false;
        if (!isDayOff && request.getShiftId() == null) {
            throw new BusinessException("非休息日必须指定班次");
        }
    }
    
    private void checkComplianceAfterSchedule(Schedule schedule) {
        LocalDate checkStart = schedule.getScheduleDate().minusDays(2);
        LocalDate checkEnd = schedule.getScheduleDate().plusDays(2);
        
        List<ComplianceViolation> violations = checkCompliance(
                schedule.getEmployee().getId(), checkStart, checkEnd);
        
        if (!violations.isEmpty()) {
            StringBuilder sb = new StringBuilder();
            sb.append("检测到合规问题：\n");
            for (ComplianceViolation v : violations) {
                sb.append(String.format("- %s (累计%d小时，超出%d小时限制)\n",
                        v.getMessage(), v.getTotalHours(), v.getMaxAllowedHours()));
            }
            throw new BusinessException(sb.toString().trim());
        }
    }
    
    private List<ComplianceViolation> detectConsecutiveWorkViolations(
            List<Schedule> schedules, LocalDate queryStart, LocalDate queryEnd) {
        
        List<ComplianceViolation> violations = new ArrayList<>();
        
        if (schedules.size() < 2) {
            return violations;
        }
        
        schedules.sort(Comparator.comparing(Schedule::getScheduleDate));
        
        Schedule prev = null;
        for (Schedule curr : schedules) {
            if (prev != null && !prev.getIsDayOff() && !curr.getIsDayOff()) {
                LocalDate expectedNextDate = prev.getScheduleDate().plusDays(1);
                if (curr.getScheduleDate().equals(expectedNextDate)) {
                    int prevHours = prev.getShift() != null ? prev.getShift().getDurationHours() : 0;
                    int currHours = curr.getShift() != null ? curr.getShift().getDurationHours() : 0;
                    int totalHours = prevHours + currHours;
                    
                    if (totalHours > MAX_CONSECUTIVE_WORK_HOURS) {
                        LocalDate violationStart = prev.getScheduleDate();
                        LocalDate violationEnd = curr.getScheduleDate();
                        
                        if ((violationStart.isEqual(queryStart) || violationStart.isAfter(queryStart))
                                && (violationEnd.isEqual(queryEnd) || violationEnd.isBefore(queryEnd))) {
                            
                            String prevShiftName = prev.getShift() != null ? prev.getShift().getName() : "未排班";
                            String currShiftName = curr.getShift() != null ? curr.getShift().getName() : "未排班";
                            
                            ComplianceViolation violation = new ComplianceViolation();
                            violation.setStartDate(violationStart);
                            violation.setEndDate(violationEnd);
                            violation.setMessage(String.format(
                                    "%s 和 %s 连续排班，%s(%d小时) + %s(%d小时) = %d小时",
                                    violationStart, violationEnd,
                                    prevShiftName, prevHours, currShiftName, currHours, totalHours));
                            violation.setTotalHours(totalHours);
                            violation.setMaxAllowedHours(MAX_CONSECUTIVE_WORK_HOURS);
                            
                            violations.add(violation);
                        }
                    }
                }
            }
            
            if (curr.getIsDayOff()) {
                prev = null;
            } else {
                prev = curr;
            }
        }
        
        return violations;
    }
    
    private ScheduleViewDTO toViewDTO(Schedule schedule) {
        ScheduleViewDTO dto = new ScheduleViewDTO();
        dto.setId(schedule.getId());
        dto.setEmployeeId(schedule.getEmployee().getId());
        dto.setEmployeeName(schedule.getEmployee().getName());
        dto.setEmployeeNo(schedule.getEmployee().getEmployeeNo());
        dto.setStoreName(schedule.getEmployee().getStoreName());
        dto.setScheduleDate(schedule.getScheduleDate());
        dto.setIsDayOff(schedule.getIsDayOff());
        
        if (schedule.getShift() != null) {
            dto.setShiftName(schedule.getShift().getName());
            dto.setShiftStartTime(schedule.getShift().getStartTime());
            dto.setShiftEndTime(schedule.getShift().getEndTime());
            dto.setShiftDurationHours(schedule.getShift().getDurationHours());
            dto.setIsShiftCrossDay(schedule.getShift().getIsCrossDay());
        }
        
        return dto;
    }
}

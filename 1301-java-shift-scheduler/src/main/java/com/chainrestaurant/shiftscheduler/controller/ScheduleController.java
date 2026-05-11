package com.chainrestaurant.shiftscheduler.controller;

import com.chainrestaurant.shiftscheduler.dto.ComplianceViolation;
import com.chainrestaurant.shiftscheduler.dto.CreateScheduleRequest;
import com.chainrestaurant.shiftscheduler.dto.ScheduleViewDTO;
import com.chainrestaurant.shiftscheduler.entity.Schedule;
import com.chainrestaurant.shiftscheduler.service.ScheduleService;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.format.annotation.DateTimeFormat;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import java.time.LocalDate;
import java.util.List;

@RestController
@RequestMapping("/api/schedules")
public class ScheduleController {
    
    @Autowired
    private ScheduleService scheduleService;
    
    @PostMapping
    public ResponseEntity<Schedule> createSchedule(@RequestBody CreateScheduleRequest request) {
        Schedule created = scheduleService.createSchedule(request);
        return ResponseEntity.ok(created);
    }
    
    @PutMapping("/{id}")
    public ResponseEntity<Schedule> updateSchedule(@PathVariable Long id, @RequestBody CreateScheduleRequest request) {
        Schedule updated = scheduleService.updateSchedule(id, request);
        return ResponseEntity.ok(updated);
    }
    
    @DeleteMapping("/{id}")
    public ResponseEntity<Void> deleteSchedule(@PathVariable Long id) {
        scheduleService.deleteSchedule(id);
        return ResponseEntity.noContent().build();
    }
    
    @GetMapping("/{id}")
    public ResponseEntity<Schedule> getScheduleById(@PathVariable Long id) {
        Schedule schedule = scheduleService.getScheduleById(id);
        return ResponseEntity.ok(schedule);
    }
    
    @GetMapping("/by-employee/{employeeId}")
    public ResponseEntity<List<ScheduleViewDTO>> getSchedulesByEmployee(
            @PathVariable Long employeeId,
            @RequestParam @DateTimeFormat(iso = DateTimeFormat.ISO.DATE) LocalDate startDate,
            @RequestParam @DateTimeFormat(iso = DateTimeFormat.ISO.DATE) LocalDate endDate) {
        List<ScheduleViewDTO> schedules = scheduleService.getSchedulesByEmployee(employeeId, startDate, endDate);
        return ResponseEntity.ok(schedules);
    }
    
    @GetMapping("/by-store/{storeName}")
    public ResponseEntity<List<ScheduleViewDTO>> getSchedulesByStore(
            @PathVariable String storeName,
            @RequestParam @DateTimeFormat(iso = DateTimeFormat.ISO.DATE) LocalDate startDate,
            @RequestParam @DateTimeFormat(iso = DateTimeFormat.ISO.DATE) LocalDate endDate) {
        List<ScheduleViewDTO> schedules = scheduleService.getSchedulesByStore(storeName, startDate, endDate);
        return ResponseEntity.ok(schedules);
    }
    
    @GetMapping("/by-date")
    public ResponseEntity<List<ScheduleViewDTO>> getSchedulesByDate(
            @RequestParam @DateTimeFormat(iso = DateTimeFormat.ISO.DATE) LocalDate date) {
        List<ScheduleViewDTO> schedules = scheduleService.getSchedulesByDate(date);
        return ResponseEntity.ok(schedules);
    }
    
    @GetMapping("/compliance/{employeeId}")
    public ResponseEntity<List<ComplianceViolation>> checkCompliance(
            @PathVariable Long employeeId,
            @RequestParam @DateTimeFormat(iso = DateTimeFormat.ISO.DATE) LocalDate startDate,
            @RequestParam @DateTimeFormat(iso = DateTimeFormat.ISO.DATE) LocalDate endDate) {
        List<ComplianceViolation> violations = scheduleService.checkCompliance(employeeId, startDate, endDate);
        return ResponseEntity.ok(violations);
    }
}

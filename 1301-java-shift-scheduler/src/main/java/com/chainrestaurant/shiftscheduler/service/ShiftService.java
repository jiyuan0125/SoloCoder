package com.chainrestaurant.shiftscheduler.service;

import com.chainrestaurant.shiftscheduler.entity.Shift;
import com.chainrestaurant.shiftscheduler.exception.BusinessException;
import com.chainrestaurant.shiftscheduler.exception.ResourceNotFoundException;
import com.chainrestaurant.shiftscheduler.repository.ShiftRepository;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.time.Duration;
import java.time.LocalTime;
import java.util.List;

@Service
@Transactional
public class ShiftService {
    
    @Autowired
    private ShiftRepository shiftRepository;
    
    public Shift createShift(Shift shift) {
        validateShift(shift);
        
        if (shiftRepository.existsByName(shift.getName())) {
            throw new BusinessException("班次名称已存在: " + shift.getName());
        }
        
        shift.setDurationHours(calculateDurationHours(shift.getStartTime(), shift.getEndTime()));
        shift.setIsCrossDay(shift.getEndTime().isBefore(shift.getStartTime()));
        
        return shiftRepository.save(shift);
    }
    
    public Shift updateShift(Long id, Shift shiftDetails) {
        Shift shift = getShiftById(id);
        
        validateShift(shiftDetails);
        
        if (!shift.getName().equals(shiftDetails.getName())) {
            if (shiftRepository.existsByName(shiftDetails.getName())) {
                throw new BusinessException("班次名称已存在: " + shiftDetails.getName());
            }
        }
        
        shift.setName(shiftDetails.getName());
        shift.setStartTime(shiftDetails.getStartTime());
        shift.setEndTime(shiftDetails.getEndTime());
        shift.setDurationHours(calculateDurationHours(shiftDetails.getStartTime(), shiftDetails.getEndTime()));
        shift.setIsCrossDay(shiftDetails.getEndTime().isBefore(shiftDetails.getStartTime()));
        
        return shiftRepository.save(shift);
    }
    
    public void deleteShift(Long id) {
        if (!shiftRepository.existsById(id)) {
            throw new ResourceNotFoundException("班次不存在: " + id);
        }
        shiftRepository.deleteById(id);
    }
    
    public Shift getShiftById(Long id) {
        return shiftRepository.findById(id)
                .orElseThrow(() -> new ResourceNotFoundException("班次不存在: " + id));
    }
    
    public List<Shift> getAllShifts() {
        return shiftRepository.findAll();
    }
    
    public Shift getShiftByName(String name) {
        return shiftRepository.findByName(name)
                .orElseThrow(() -> new ResourceNotFoundException("班次不存在: " + name));
    }
    
    private void validateShift(Shift shift) {
        if (shift.getStartTime() == null || shift.getEndTime() == null) {
            throw new BusinessException("班次开始时间和结束时间不能为空");
        }
    }
    
    private int calculateDurationHours(LocalTime startTime, LocalTime endTime) {
        if (endTime.isAfter(startTime)) {
            return (int) Duration.between(startTime, endTime).toHours();
        } else {
            return (int) Duration.between(startTime, endTime.plusHours(24)).toHours();
        }
    }
}

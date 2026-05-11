package com.chainrestaurant.shiftscheduler.controller;

import com.chainrestaurant.shiftscheduler.entity.Shift;
import com.chainrestaurant.shiftscheduler.service.ShiftService;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import java.util.List;

@RestController
@RequestMapping("/api/shifts")
public class ShiftController {
    
    @Autowired
    private ShiftService shiftService;
    
    @PostMapping
    public ResponseEntity<Shift> createShift(@RequestBody Shift shift) {
        Shift created = shiftService.createShift(shift);
        return ResponseEntity.ok(created);
    }
    
    @PutMapping("/{id}")
    public ResponseEntity<Shift> updateShift(@PathVariable Long id, @RequestBody Shift shift) {
        Shift updated = shiftService.updateShift(id, shift);
        return ResponseEntity.ok(updated);
    }
    
    @DeleteMapping("/{id}")
    public ResponseEntity<Void> deleteShift(@PathVariable Long id) {
        shiftService.deleteShift(id);
        return ResponseEntity.noContent().build();
    }
    
    @GetMapping("/{id}")
    public ResponseEntity<Shift> getShiftById(@PathVariable Long id) {
        Shift shift = shiftService.getShiftById(id);
        return ResponseEntity.ok(shift);
    }
    
    @GetMapping
    public ResponseEntity<List<Shift>> getAllShifts() {
        return ResponseEntity.ok(shiftService.getAllShifts());
    }
    
    @GetMapping("/by-name/{name}")
    public ResponseEntity<Shift> getShiftByName(@PathVariable String name) {
        Shift shift = shiftService.getShiftByName(name);
        return ResponseEntity.ok(shift);
    }
}

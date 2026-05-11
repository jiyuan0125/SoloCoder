package com.hotelbooking.controller;

import com.hotelbooking.exception.ResourceNotFoundException;
import com.hotelbooking.model.entity.Holiday;
import com.hotelbooking.model.enums.HolidayType;
import com.hotelbooking.repository.HolidayRepository;
import lombok.RequiredArgsConstructor;
import org.springframework.format.annotation.DateTimeFormat;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import java.time.LocalDate;
import java.util.List;

@RestController
@RequestMapping("/api/holidays")
@RequiredArgsConstructor
public class HolidayController {

    private final HolidayRepository holidayRepository;

    @GetMapping
    public ResponseEntity<List<Holiday>> getAllHolidays() {
        return ResponseEntity.ok(holidayRepository.findAll());
    }

    @GetMapping("/date/{date}")
    public ResponseEntity<Holiday> getHolidayByDate(
            @PathVariable @DateTimeFormat(iso = DateTimeFormat.ISO.DATE) LocalDate date) {
        Holiday holiday = holidayRepository.findByHolidayDate(date)
                .orElseThrow(() -> new ResourceNotFoundException("该日期不是节假日"));
        return ResponseEntity.ok(holiday);
    }

    @GetMapping("/range")
    public ResponseEntity<List<Holiday>> getHolidaysInRange(
            @RequestParam @DateTimeFormat(iso = DateTimeFormat.ISO.DATE) LocalDate start,
            @RequestParam @DateTimeFormat(iso = DateTimeFormat.ISO.DATE) LocalDate end) {
        return ResponseEntity.ok(holidayRepository.findHolidaysInRange(start, end));
    }

    @PostMapping
    public ResponseEntity<Holiday> createHoliday(@RequestBody Holiday holiday) {
        return ResponseEntity.ok(holidayRepository.save(holiday));
    }

    @PostMapping("/batch")
    public ResponseEntity<List<Holiday>> createHolidays(@RequestBody List<Holiday> holidays) {
        return ResponseEntity.ok(holidayRepository.saveAll(holidays));
    }

    @PutMapping("/{id}")
    public ResponseEntity<Holiday> updateHoliday(@PathVariable Long id, @RequestBody Holiday holidayDetails) {
        Holiday holiday = holidayRepository.findById(id)
                .orElseThrow(() -> new ResourceNotFoundException("节假日不存在"));
        
        if (holidayDetails.getName() != null) holiday.setName(holidayDetails.getName());
        if (holidayDetails.getHolidayDate() != null) holiday.setHolidayDate(holidayDetails.getHolidayDate());
        if (holidayDetails.getHolidayType() != null) holiday.setHolidayType(holidayDetails.getHolidayType());
        if (holidayDetails.getDescription() != null) holiday.setDescription(holidayDetails.getDescription());
        
        return ResponseEntity.ok(holidayRepository.save(holiday));
    }

    @DeleteMapping("/{id}")
    public ResponseEntity<Void> deleteHoliday(@PathVariable Long id) {
        if (!holidayRepository.existsById(id)) {
            throw new ResourceNotFoundException("节假日不存在");
        }
        holidayRepository.deleteById(id);
        return ResponseEntity.ok().build();
    }

    @GetMapping("/types")
    public ResponseEntity<HolidayType[]> getHolidayTypes() {
        return ResponseEntity.ok(HolidayType.values());
    }
}

package com.hospital.controller;

import com.hospital.dto.AppointmentDTO;
import com.hospital.dto.BookingRequestDTO;
import com.hospital.dto.SlotAvailabilityDTO;
import com.hospital.enums.TimeSlot;
import com.hospital.service.AppointmentService;
import org.springframework.format.annotation.DateTimeFormat;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import java.time.LocalDate;
import java.util.List;

@RestController
@RequestMapping("/api/appointments")
public class AppointmentController {

    private final AppointmentService appointmentService;

    public AppointmentController(AppointmentService appointmentService) {
        this.appointmentService = appointmentService;
    }

    @PostMapping("/book")
    public ResponseEntity<AppointmentDTO> bookAppointment(@RequestBody BookingRequestDTO request) {
        AppointmentDTO booked = appointmentService.bookAppointment(request);
        return ResponseEntity.status(HttpStatus.CREATED).body(booked);
    }

    @PostMapping("/{id}/cancel")
    public ResponseEntity<AppointmentDTO> cancelAppointment(@PathVariable Long id) {
        AppointmentDTO cancelled = appointmentService.cancelAppointment(id);
        return ResponseEntity.ok(cancelled);
    }

    @GetMapping("/availability")
    public ResponseEntity<SlotAvailabilityDTO> checkAvailability(
            @RequestParam Long doctorId,
            @RequestParam @DateTimeFormat(iso = DateTimeFormat.ISO.DATE) LocalDate date,
            @RequestParam TimeSlot timeSlot) {
        return ResponseEntity.ok(appointmentService.checkSlotAvailability(doctorId, date, timeSlot));
    }

    @GetMapping("/patient")
    public ResponseEntity<List<AppointmentDTO>> getPatientAppointments(
            @RequestParam String phone,
            @RequestParam(required = false) String name) {
        List<AppointmentDTO> appointments;
        if (name != null && !name.trim().isEmpty()) {
            appointments = appointmentService.getPatientAppointmentsByNameAndPhone(name, phone);
        } else {
            appointments = appointmentService.getPatientAppointments(phone);
        }
        return ResponseEntity.ok(appointments);
    }

    @GetMapping("/doctor/{doctorId}/date/{date}")
    public ResponseEntity<List<AppointmentDTO>> getDoctorDailyAppointments(
            @PathVariable Long doctorId,
            @PathVariable @DateTimeFormat(iso = DateTimeFormat.ISO.DATE) LocalDate date) {
        return ResponseEntity.ok(appointmentService.getDoctorDailyAppointments(doctorId, date));
    }

    @GetMapping("/{id}")
    public ResponseEntity<AppointmentDTO> getAppointmentById(@PathVariable Long id) {
        return ResponseEntity.ok(appointmentService.getAppointmentById(id));
    }
}

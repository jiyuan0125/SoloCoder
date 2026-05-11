package com.company.meetingroom.controller;

import com.company.meetingroom.dto.AvailabilityRequest;
import com.company.meetingroom.dto.ReservationRequest;
import com.company.meetingroom.model.Reservation;
import com.company.meetingroom.model.Room;
import com.company.meetingroom.service.ReservationService;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import java.time.LocalDate;
import java.util.List;

@RestController
@RequestMapping("/api/reservations")
@CrossOrigin(origins = "*")
public class ReservationController {
    
    private final ReservationService reservationService;
    
    @Autowired
    public ReservationController(ReservationService reservationService) {
        this.reservationService = reservationService;
    }
    
    @PostMapping
    public ResponseEntity<?> createReservation(@RequestBody ReservationRequest request) {
        try {
            Reservation reservation = reservationService.createReservation(
                    request.getRoomId(),
                    request.getStartTime(),
                    request.getEndTime(),
                    request.getBooker(),
                    request.getTopic()
            );
            return ResponseEntity.ok(reservation);
        } catch (Exception e) {
            return ResponseEntity.badRequest().body(e.getMessage());
        }
    }
    
    @GetMapping("/{id}")
    public ResponseEntity<Reservation> getReservationById(@PathVariable Long id) {
        return reservationService.getReservationById(id)
                .map(ResponseEntity::ok)
                .orElse(ResponseEntity.notFound().build());
    }
    
    @PostMapping("/{id}/checkin")
    public ResponseEntity<?> checkIn(@PathVariable Long id) {
        try {
            return ResponseEntity.ok(reservationService.checkIn(id));
        } catch (Exception e) {
            return ResponseEntity.badRequest().body(e.getMessage());
        }
    }
    
    @PostMapping("/{id}/cancel")
    public ResponseEntity<?> cancelReservation(@PathVariable Long id) {
        try {
            return ResponseEntity.ok(reservationService.cancelReservation(id));
        } catch (Exception e) {
            return ResponseEntity.badRequest().body(e.getMessage());
        }
    }
    
    @GetMapping("/room/{roomId}/date/{date}")
    public List<Reservation> getReservationsByRoomAndDate(
            @PathVariable Long roomId,
            @PathVariable String date) {
        LocalDate localDate = LocalDate.parse(date);
        return reservationService.getReservationsByRoomAndDate(roomId, localDate);
    }
    
    @GetMapping("/booker/{booker}")
    public List<Reservation> getReservationsByBooker(@PathVariable String booker) {
        return reservationService.getReservationsByBooker(booker);
    }
    
    @PostMapping("/available")
    public List<Room> getAvailableRooms(@RequestBody AvailabilityRequest request) {
        return reservationService.getAvailableRooms(request.getStartTime(), request.getEndTime());
    }
}

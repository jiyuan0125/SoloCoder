package com.hotelbooking.controller;

import com.hotelbooking.dto.AvailabilityResponse;
import com.hotelbooking.dto.BookingCancellationResult;
import com.hotelbooking.dto.BookingRequest;
import com.hotelbooking.dto.PriceCalculationResult;
import com.hotelbooking.exception.ResourceNotFoundException;
import com.hotelbooking.model.entity.Booking;
import com.hotelbooking.model.entity.Customer;
import com.hotelbooking.model.enums.BookingStatus;
import com.hotelbooking.service.BookingService;
import com.hotelbooking.service.InventoryService;
import com.hotelbooking.service.PriceService;
import jakarta.validation.Valid;
import lombok.RequiredArgsConstructor;
import org.springframework.format.annotation.DateTimeFormat;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import java.time.LocalDate;
import java.util.List;

@RestController
@RequestMapping("/api/bookings")
@RequiredArgsConstructor
public class BookingController {

    private final BookingService bookingService;
    private final InventoryService inventoryService;
    private final PriceService priceService;

    @GetMapping("/availability")
    public ResponseEntity<List<AvailabilityResponse>> checkAvailability(
            @RequestParam @DateTimeFormat(iso = DateTimeFormat.ISO.DATE) LocalDate checkIn,
            @RequestParam @DateTimeFormat(iso = DateTimeFormat.ISO.DATE) LocalDate checkOut) {
        List<AvailabilityResponse> availability = inventoryService.checkAllAvailability(checkIn, checkOut);
        return ResponseEntity.ok(availability);
    }

    @GetMapping("/price")
    public ResponseEntity<PriceCalculationResult> calculatePrice(
            @RequestParam com.hotelbooking.model.enums.RoomType roomType,
            @RequestParam @DateTimeFormat(iso = DateTimeFormat.ISO.DATE) LocalDate checkIn,
            @RequestParam @DateTimeFormat(iso = DateTimeFormat.ISO.DATE) LocalDate checkOut,
            @RequestParam(required = false) Long customerId) {
        
        Customer customer = null;
        if (customerId != null) {
            customer = new Customer();
            customer.setId(customerId);
        }
        
        PriceCalculationResult price = priceService.calculatePrice(roomType, checkIn, checkOut, customer);
        return ResponseEntity.ok(price);
    }

    @PostMapping
    public ResponseEntity<Booking> createBooking(@Valid @RequestBody BookingRequest request) {
        Booking booking = bookingService.createBooking(request);
        return ResponseEntity.ok(booking);
    }

    @GetMapping("/{bookingNumber}")
    public ResponseEntity<Booking> getBooking(@PathVariable String bookingNumber) {
        Booking booking = bookingService.findByBookingNumber(bookingNumber)
                .orElseThrow(() -> new ResourceNotFoundException("预订不存在"));
        return ResponseEntity.ok(booking);
    }

    @PutMapping("/{id}/confirm")
    public ResponseEntity<Booking> confirmBooking(@PathVariable Long id) {
        Booking booking = bookingService.confirmBooking(id);
        return ResponseEntity.ok(booking);
    }

    @PutMapping("/{id}/check-in")
    public ResponseEntity<Booking> checkIn(@PathVariable Long id) {
        Booking booking = bookingService.checkIn(id);
        return ResponseEntity.ok(booking);
    }

    @PutMapping("/{id}/check-out")
    public ResponseEntity<Booking> checkOut(
            @PathVariable Long id,
            @RequestParam(required = false) @DateTimeFormat(iso = DateTimeFormat.ISO.DATE) LocalDate actualCheckOut) {
        Booking booking = bookingService.checkOut(id, actualCheckOut);
        return ResponseEntity.ok(booking);
    }

    @PutMapping("/{id}/cancel")
    public ResponseEntity<BookingCancellationResult> cancelBooking(@PathVariable Long id) {
        BookingCancellationResult result = bookingService.cancelBooking(id);
        return ResponseEntity.ok(result);
    }

    @PutMapping("/{id}/upgrade")
    public ResponseEntity<Booking> upgradeRoom(
            @PathVariable Long id,
            @RequestParam com.hotelbooking.model.enums.RoomType newRoomType,
            @RequestParam(defaultValue = "false") boolean hotelCaused) {
        Booking booking = bookingService.upgradeRoom(id, newRoomType, hotelCaused);
        return ResponseEntity.ok(booking);
    }

    @GetMapping("/customer/{customerId}")
    public ResponseEntity<List<Booking>> getCustomerBookings(@PathVariable Long customerId) {
        List<Booking> bookings = bookingService.findByCustomerId(customerId);
        return ResponseEntity.ok(bookings);
    }

    @GetMapping("/status/{status}")
    public ResponseEntity<List<Booking>> getBookingsByStatus(@PathVariable BookingStatus status) {
        List<Booking> bookings = bookingService.findByStatus(status);
        return ResponseEntity.ok(bookings);
    }
}

package com.hotelbooking.repository;

import com.hotelbooking.model.entity.Booking;
import com.hotelbooking.model.enums.BookingStatus;
import com.hotelbooking.model.enums.RoomType;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.data.jpa.repository.Query;
import org.springframework.data.repository.query.Param;
import org.springframework.stereotype.Repository;

import java.time.LocalDate;
import java.util.List;
import java.util.Optional;

@Repository
public interface BookingRepository extends JpaRepository<Booking, Long> {
    Optional<Booking> findByBookingNumber(String bookingNumber);
    
    List<Booking> findByCustomerId(Long customerId);
    
    List<Booking> findByStatus(BookingStatus status);
    
    List<Booking> findByRoomIdAndStatusIn(Long roomId, List<BookingStatus> statuses);
    
    @Query("SELECT b FROM Booking b WHERE b.roomType = :roomType " +
           "AND b.status IN ('PENDING_CONFIRMATION', 'CONFIRMED', 'CHECKED_IN') " +
           "AND :checkInDate < b.checkOutDate AND :checkOutDate > b.checkInDate")
    List<Booking> findActiveBookingsInDateRange(
            @Param("roomType") RoomType roomType,
            @Param("checkInDate") LocalDate checkInDate,
            @Param("checkOutDate") LocalDate checkOutDate);
    
    @Query("SELECT COUNT(DISTINCT b) FROM Booking b WHERE b.roomType = :roomType " +
           "AND b.status IN ('PENDING_CONFIRMATION', 'CONFIRMED', 'CHECKED_IN') " +
           "AND :date >= b.checkInDate AND :date < b.checkOutDate")
    Long countBookedRoomsForDate(
            @Param("roomType") RoomType roomType,
            @Param("date") LocalDate date);
    
    @Query("SELECT b FROM Booking b WHERE b.customer.id = :customerId " +
           "AND b.status IN ('CHECKED_IN', 'CHECKED_OUT', 'CONFIRMED') " +
           "AND b.checkOutDate < :date " +
           "ORDER BY b.checkOutDate DESC")
    List<Booking> findRecentBookingsByCustomer(
            @Param("customerId") Long customerId,
            @Param("date") LocalDate date);
}

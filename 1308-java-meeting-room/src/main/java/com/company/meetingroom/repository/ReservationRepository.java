package com.company.meetingroom.repository;

import com.company.meetingroom.model.Reservation;
import com.company.meetingroom.model.ReservationStatus;
import com.company.meetingroom.model.Room;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.data.jpa.repository.Query;
import org.springframework.data.repository.query.Param;
import org.springframework.stereotype.Repository;

import java.time.LocalDateTime;
import java.util.List;

@Repository
public interface ReservationRepository extends JpaRepository<Reservation, Long> {
    
    @Query("SELECT r FROM Reservation r WHERE r.room = :room AND r.status IN :statuses " +
           "AND r.startTime < :endTime AND r.endTime > :startTime")
    List<Reservation> findConflictingReservations(
            @Param("room") Room room,
            @Param("startTime") LocalDateTime startTime,
            @Param("endTime") LocalDateTime endTime,
            @Param("statuses") List<ReservationStatus> statuses);
    
    List<Reservation> findByRoomAndStartTimeBetweenOrderByStartTimeAsc(
            Room room, LocalDateTime start, LocalDateTime end);
    
    List<Reservation> findByBookerOrderByStartTimeAsc(String booker);
    
    @Query("SELECT r FROM Reservation r WHERE r.status = :status AND r.startTime < :threshold")
    List<Reservation> findBookedReservationsBeforeThreshold(
            @Param("status") ReservationStatus status,
            @Param("threshold") LocalDateTime threshold);
}

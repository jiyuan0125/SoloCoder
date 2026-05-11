package com.hotelbooking.repository;

import com.hotelbooking.model.entity.Room;
import com.hotelbooking.model.enums.RoomStatus;
import com.hotelbooking.model.enums.RoomType;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.data.jpa.repository.Query;
import org.springframework.data.repository.query.Param;
import org.springframework.stereotype.Repository;

import java.util.List;

@Repository
public interface RoomRepository extends JpaRepository<Room, Long> {
    List<Room> findByRoomType(RoomType roomType);
    List<Room> findByStatus(RoomStatus status);
    List<Room> findByRoomTypeAndStatus(RoomType roomType, RoomStatus status);
    Long countByRoomTypeAndStatus(RoomType roomType, RoomStatus status);
    Long countByRoomType(RoomType roomType);
    
    @Query("SELECT r FROM Room r WHERE r.roomType = :roomType AND r.status = 'AVAILABLE' " +
           "AND (r.cleaningAvailableAt IS NULL OR r.cleaningAvailableAt <= CURRENT_TIMESTAMP)")
    List<Room> findAvailableRooms(@Param("roomType") RoomType roomType);
    
    @Query("SELECT COUNT(r) FROM Room r WHERE r.roomType = :roomType AND r.status = 'AVAILABLE' " +
           "AND (r.cleaningAvailableAt IS NULL OR r.cleaningAvailableAt <= CURRENT_TIMESTAMP)")
    Long countAvailableRooms(@Param("roomType") RoomType roomType);
}

package com.hotelbooking.controller;

import com.hotelbooking.exception.ResourceNotFoundException;
import com.hotelbooking.model.entity.Room;
import com.hotelbooking.model.entity.RoomTypeConfig;
import com.hotelbooking.model.enums.RoomStatus;
import com.hotelbooking.model.enums.RoomType;
import com.hotelbooking.repository.RoomRepository;
import com.hotelbooking.repository.RoomTypeConfigRepository;
import lombok.RequiredArgsConstructor;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import java.util.List;

@RestController
@RequestMapping("/api/rooms")
@RequiredArgsConstructor
public class RoomController {

    private final RoomRepository roomRepository;
    private final RoomTypeConfigRepository roomTypeConfigRepository;

    @GetMapping
    public ResponseEntity<List<Room>> getAllRooms() {
        return ResponseEntity.ok(roomRepository.findAll());
    }

    @GetMapping("/type/{roomType}")
    public ResponseEntity<List<Room>> getRoomsByType(@PathVariable RoomType roomType) {
        return ResponseEntity.ok(roomRepository.findByRoomType(roomType));
    }

    @GetMapping("/status/{status}")
    public ResponseEntity<List<Room>> getRoomsByStatus(@PathVariable RoomStatus status) {
        return ResponseEntity.ok(roomRepository.findByStatus(status));
    }

    @GetMapping("/{id}")
    public ResponseEntity<Room> getRoom(@PathVariable Long id) {
        Room room = roomRepository.findById(id)
                .orElseThrow(() -> new ResourceNotFoundException("房间不存在"));
        return ResponseEntity.ok(room);
    }

    @PostMapping
    public ResponseEntity<Room> createRoom(@RequestBody Room room) {
        return ResponseEntity.ok(roomRepository.save(room));
    }

    @PutMapping("/{id}")
    public ResponseEntity<Room> updateRoom(@PathVariable Long id, @RequestBody Room roomDetails) {
        Room room = roomRepository.findById(id)
                .orElseThrow(() -> new ResourceNotFoundException("房间不存在"));
        
        if (roomDetails.getRoomNumber() != null) room.setRoomNumber(roomDetails.getRoomNumber());
        if (roomDetails.getRoomType() != null) room.setRoomType(roomDetails.getRoomType());
        if (roomDetails.getFloor() != null) room.setFloor(roomDetails.getFloor());
        if (roomDetails.getStatus() != null) room.setStatus(roomDetails.getStatus());
        if (roomDetails.getNotes() != null) room.setNotes(roomDetails.getNotes());
        
        return ResponseEntity.ok(roomRepository.save(room));
    }

    @GetMapping("/configs")
    public ResponseEntity<List<RoomTypeConfig>> getAllConfigs() {
        return ResponseEntity.ok(roomTypeConfigRepository.findAll());
    }

    @GetMapping("/configs/{roomType}")
    public ResponseEntity<RoomTypeConfig> getConfig(@PathVariable RoomType roomType) {
        RoomTypeConfig config = roomTypeConfigRepository.findByRoomType(roomType)
                .orElseThrow(() -> new ResourceNotFoundException("房型配置不存在"));
        return ResponseEntity.ok(config);
    }

    @PutMapping("/configs/{roomType}")
    public ResponseEntity<RoomTypeConfig> updateConfig(
            @PathVariable RoomType roomType,
            @RequestBody RoomTypeConfig configDetails) {
        RoomTypeConfig config = roomTypeConfigRepository.findByRoomType(roomType)
                .orElseThrow(() -> new ResourceNotFoundException("房型配置不存在"));
        
        if (configDetails.getRoomCount() != null) config.setRoomCount(configDetails.getRoomCount());
        if (configDetails.getBasePrice() != null) config.setBasePrice(configDetails.getBasePrice());
        if (configDetails.getHolidayEveSurcharge() != null) config.setHolidayEveSurcharge(configDetails.getHolidayEveSurcharge());
        if (configDetails.getHolidaySurcharge() != null) config.setHolidaySurcharge(configDetails.getHolidaySurcharge());
        if (configDetails.getHolidayAfterSurcharge() != null) config.setHolidayAfterSurcharge(configDetails.getHolidayAfterSurcharge());
        if (configDetails.getMaxGuests() != null) config.setMaxGuests(configDetails.getMaxGuests());
        if (configDetails.getDescription() != null) config.setDescription(configDetails.getDescription());
        
        return ResponseEntity.ok(roomTypeConfigRepository.save(config));
    }
}

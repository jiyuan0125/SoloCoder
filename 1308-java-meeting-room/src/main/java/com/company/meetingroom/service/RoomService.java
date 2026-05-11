package com.company.meetingroom.service;

import com.company.meetingroom.model.Room;
import com.company.meetingroom.repository.RoomRepository;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.util.List;
import java.util.Optional;

@Service
public class RoomService {
    
    private final RoomRepository roomRepository;
    
    @Autowired
    public RoomService(RoomRepository roomRepository) {
        this.roomRepository = roomRepository;
    }
    
    public List<Room> getAllRooms() {
        return roomRepository.findAll();
    }
    
    public Optional<Room> getRoomById(Long id) {
        return roomRepository.findById(id);
    }
    
    @Transactional
    public Room createRoom(Room room) {
        if (roomRepository.existsByCode(room.getCode())) {
            throw new RuntimeException("会议室编号已存在: " + room.getCode());
        }
        return roomRepository.save(room);
    }
    
    @Transactional
    public Room updateRoom(Long id, Room roomDetails) {
        return roomRepository.findById(id).map(room -> {
            if (!room.getCode().equals(roomDetails.getCode()) 
                && roomRepository.existsByCode(roomDetails.getCode())) {
                throw new RuntimeException("会议室编号已存在: " + roomDetails.getCode());
            }
            room.setCode(roomDetails.getCode());
            room.setName(roomDetails.getName());
            room.setLocation(roomDetails.getLocation());
            room.setCapacity(roomDetails.getCapacity());
            return roomRepository.save(room);
        }).orElseThrow(() -> new RuntimeException("会议室不存在: " + id));
    }
    
    @Transactional
    public void deleteRoom(Long id) {
        if (!roomRepository.existsById(id)) {
            throw new RuntimeException("会议室不存在: " + id);
        }
        roomRepository.deleteById(id);
    }
}

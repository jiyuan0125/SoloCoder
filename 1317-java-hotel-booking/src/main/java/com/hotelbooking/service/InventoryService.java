package com.hotelbooking.service;

import com.hotelbooking.dto.AvailabilityResponse;
import com.hotelbooking.exception.BookingException;
import com.hotelbooking.model.entity.Booking;
import com.hotelbooking.model.entity.Room;
import com.hotelbooking.model.entity.RoomTypeConfig;
import com.hotelbooking.model.enums.BookingStatus;
import com.hotelbooking.model.enums.RoomStatus;
import com.hotelbooking.model.enums.RoomType;
import com.hotelbooking.repository.BookingRepository;
import com.hotelbooking.repository.RoomRepository;
import com.hotelbooking.repository.RoomTypeConfigRepository;
import lombok.RequiredArgsConstructor;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.time.LocalDate;
import java.time.temporal.ChronoUnit;
import java.util.ArrayList;
import java.util.Arrays;
import java.util.List;

@Service
@RequiredArgsConstructor
public class InventoryService {

    private final RoomRepository roomRepository;
    private final RoomTypeConfigRepository roomTypeConfigRepository;
    private final BookingRepository bookingRepository;

    public List<AvailabilityResponse> checkAllAvailability(LocalDate checkInDate, LocalDate checkOutDate) {
        validateDates(checkInDate, checkOutDate);
        
        List<AvailabilityResponse> results = new ArrayList<>();
        for (RoomType roomType : RoomType.values()) {
            results.add(checkAvailability(roomType, checkInDate, checkOutDate));
        }
        return results;
    }

    public AvailabilityResponse checkAvailability(RoomType roomType, LocalDate checkInDate, LocalDate checkOutDate) {
        validateDates(checkInDate, checkOutDate);

        RoomTypeConfig config = roomTypeConfigRepository.findByRoomType(roomType)
                .orElseThrow(() -> new BookingException("房型配置不存在: " + roomType.getDisplayName()));

        int totalRooms = config.getRoomCount();
        int availableRooms = calculateAvailableRooms(roomType, checkInDate, checkOutDate, totalRooms);

        return AvailabilityResponse.builder()
                .roomType(roomType)
                .roomTypeName(roomType.getDisplayName())
                .totalRooms(totalRooms)
                .availableRooms(availableRooms)
                .basePrice(config.getBasePrice())
                .estimatedPricePerNight(config.getBasePrice())
                .available(availableRooms > 0)
                .build();
    }

    @Transactional(readOnly = true)
    public int calculateAvailableRooms(RoomType roomType, LocalDate checkInDate, LocalDate checkOutDate, int totalRooms) {
        int maxBooked = 0;
        LocalDate current = checkInDate;
        
        while (current.isBefore(checkOutDate)) {
            Long bookedCount = bookingRepository.countBookedRoomsForDate(roomType, current);
            maxBooked = Math.max(maxBooked, bookedCount != null ? bookedCount.intValue() : 0);
            current = current.plusDays(1);
        }
        
        return Math.max(0, totalRooms - maxBooked);
    }

    @Transactional(readOnly = true)
    public int calculateAvailableRoomsWithCleaningBuffer(RoomType roomType, LocalDate checkInDate, LocalDate checkOutDate, int totalRooms) {
        int maxBooked = 0;
        LocalDate current = checkInDate;
        
        while (current.isBefore(checkOutDate)) {
            Long bookedCount = bookingRepository.countBookedRoomsForDate(roomType, current);
            maxBooked = Math.max(maxBooked, bookedCount != null ? bookedCount.intValue() : 0);
            current = current.plusDays(1);
        }
        
        return Math.max(0, totalRooms - maxBooked);
    }

    @Transactional(readOnly = true)
    public void validateAvailability(Booking booking) {
        RoomTypeConfig config = roomTypeConfigRepository.findByRoomType(booking.getRoomType())
                .orElseThrow(() -> new BookingException("房型配置不存在"));

        int available = calculateAvailableRooms(
                booking.getRoomType(),
                booking.getCheckInDate(),
                booking.getCheckOutDate(),
                config.getRoomCount()
        );

        if (available <= 0) {
            throw new BookingException("所选房型在指定日期没有可用房间");
        }
    }

    @Transactional(readOnly = true)
    public List<Room> findAvailableRoomsForCheckIn(RoomType roomType, LocalDate checkInDate) {
        List<Room> rooms = roomRepository.findByRoomTypeAndStatus(roomType, RoomStatus.AVAILABLE);
        List<Room> availableRooms = new ArrayList<>();
        
        for (Room room : rooms) {
            if (isRoomAvailable(room, checkInDate)) {
                availableRooms.add(room);
            }
        }
        
        return availableRooms;
    }

    private boolean isRoomAvailable(Room room, LocalDate checkInDate) {
        if (room.getCleaningAvailableAt() != null) {
            LocalDate cleaningAvailableDate = room.getCleaningAvailableAt().toLocalDate();
            if (checkInDate.isBefore(cleaningAvailableDate)) {
                return false;
            }
        }
        
        List<BookingStatus> activeStatuses = Arrays.asList(
                BookingStatus.PENDING_CONFIRMATION,
                BookingStatus.CONFIRMED,
                BookingStatus.CHECKED_IN
        );
        
        List<Booking> roomBookings = bookingRepository.findByRoomIdAndStatusIn(room.getId(), activeStatuses);
        for (Booking booking : roomBookings) {
            if (!checkInDate.isBefore(booking.getCheckInDate()) && checkInDate.isBefore(booking.getCheckOutDate())) {
                return false;
            }
        }
        
        return true;
    }

    @Transactional
    public Room assignRoomToBooking(Booking booking) {
        List<Room> availableRooms = findAvailableRoomsForCheckIn(booking.getRoomType(), booking.getCheckInDate());
        
        if (availableRooms.isEmpty()) {
            throw new BookingException("没有可用房间可以分配");
        }
        
        Room room = availableRooms.get(0);
        room.setStatus(RoomStatus.RESERVED);
        roomRepository.save(room);
        
        return room;
    }

    private void validateDates(LocalDate checkInDate, LocalDate checkOutDate) {
        if (checkInDate == null || checkOutDate == null) {
            throw new BookingException("入住和退房日期不能为空");
        }
        
        if (checkOutDate.isBefore(checkInDate) || checkOutDate.isEqual(checkInDate)) {
            throw new BookingException("退房日期必须晚于入住日期");
        }
        
        long nights = ChronoUnit.DAYS.between(checkInDate, checkOutDate);
        if (nights > 365) {
            throw new BookingException("单次预订不能超过365天");
        }
    }
}

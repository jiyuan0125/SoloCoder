package com.company.meetingroom.service;

import com.company.meetingroom.model.Reservation;
import com.company.meetingroom.model.ReservationStatus;
import com.company.meetingroom.model.Room;
import com.company.meetingroom.repository.ReservationRepository;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.time.Duration;
import java.time.LocalDate;
import java.time.LocalDateTime;
import java.time.LocalTime;
import java.util.Arrays;
import java.util.List;
import java.util.Optional;
import java.util.stream.Collectors;

@Service
public class ReservationService {
    
    private static final long MIN_DURATION_MINUTES = 15;
    private static final long MAX_DURATION_MINUTES = 4 * 60;
    private static final long CHECKIN_WINDOW_MINUTES = 10;
    
    private final ReservationRepository reservationRepository;
    private final RoomService roomService;
    
    @Autowired
    public ReservationService(ReservationRepository reservationRepository, RoomService roomService) {
        this.reservationRepository = reservationRepository;
        this.roomService = roomService;
    }
    
    @Transactional
    public Reservation createReservation(Long roomId, LocalDateTime startTime, 
                                         LocalDateTime endTime, String booker, String topic) {
        validateTimeRange(startTime, endTime);
        
        Room room = roomService.getRoomById(roomId)
                .orElseThrow(() -> new RuntimeException("会议室不存在: " + roomId));
        
        checkTimeConflict(room, startTime, endTime);
        
        Reservation reservation = new Reservation();
        reservation.setRoom(room);
        reservation.setStartTime(startTime);
        reservation.setEndTime(endTime);
        reservation.setBooker(booker);
        reservation.setTopic(topic);
        reservation.setStatus(ReservationStatus.BOOKED);
        
        return reservationRepository.save(reservation);
    }
    
    @Transactional
    public Reservation checkIn(Long reservationId) {
        Reservation reservation = reservationRepository.findById(reservationId)
                .orElseThrow(() -> new RuntimeException("预约不存在: " + reservationId));
        
        LocalDateTime now = LocalDateTime.now();
        LocalDateTime checkInStart = reservation.getStartTime().minusMinutes(CHECKIN_WINDOW_MINUTES);
        LocalDateTime checkInEnd = reservation.getStartTime().plusMinutes(CHECKIN_WINDOW_MINUTES);
        
        if (reservation.getStatus() != ReservationStatus.BOOKED) {
            throw new RuntimeException("只有已预约状态的会议才能签到");
        }
        
        if (now.isBefore(checkInStart)) {
            throw new RuntimeException("签到尚未开始，请在会议开始前10分钟内签到");
        }
        
        if (now.isAfter(checkInEnd)) {
            throw new RuntimeException("签到时间已过，无法签到");
        }
        
        reservation.setStatus(ReservationStatus.CHECKED_IN);
        reservation.setCheckInTime(now);
        return reservationRepository.save(reservation);
    }
    
    @Transactional
    public Reservation cancelReservation(Long reservationId) {
        Reservation reservation = reservationRepository.findById(reservationId)
                .orElseThrow(() -> new RuntimeException("预约不存在: " + reservationId));
        
        LocalDateTime now = LocalDateTime.now();
        LocalDateTime checkInStart = reservation.getStartTime().minusMinutes(CHECKIN_WINDOW_MINUTES);
        
        if (reservation.getStatus() != ReservationStatus.BOOKED) {
            throw new RuntimeException("只有已预约状态的会议才能取消");
        }
        
        if (now.isAfter(reservation.getStartTime())) {
            throw new RuntimeException("会议已经开始，无法取消");
        }
        
        if (!now.isBefore(checkInStart)) {
            throw new RuntimeException("签到时间已开始，无法取消预约");
        }
        
        reservation.setStatus(ReservationStatus.CANCELLED);
        return reservationRepository.save(reservation);
    }
    
    @Transactional
    public void autoReleaseExpiredReservations() {
        LocalDateTime threshold = LocalDateTime.now().minusMinutes(CHECKIN_WINDOW_MINUTES);
        List<Reservation> expired = reservationRepository.findBookedReservationsBeforeThreshold(
                ReservationStatus.BOOKED, threshold);
        
        for (Reservation r : expired) {
            if (r.getCheckInTime() == null) {
                r.setStatus(ReservationStatus.RELEASED);
                reservationRepository.save(r);
            }
        }
    }
    
    public List<Reservation> getReservationsByRoomAndDate(Long roomId, LocalDate date) {
        autoReleaseExpiredReservations();
        Room room = roomService.getRoomById(roomId)
                .orElseThrow(() -> new RuntimeException("会议室不存在: " + roomId));
        
        LocalDateTime start = date.atStartOfDay();
        LocalDateTime end = date.atTime(LocalTime.MAX);
        return reservationRepository.findByRoomAndStartTimeBetweenOrderByStartTimeAsc(room, start, end);
    }
    
    public List<Reservation> getReservationsByBooker(String booker) {
        autoReleaseExpiredReservations();
        return reservationRepository.findByBookerOrderByStartTimeAsc(booker);
    }
    
    public List<Room> getAvailableRooms(LocalDateTime startTime, LocalDateTime endTime) {
        validateTimeRange(startTime, endTime);
        autoReleaseExpiredReservations();
        
        List<ReservationStatus> activeStatuses = Arrays.asList(
                ReservationStatus.BOOKED, 
                ReservationStatus.CHECKED_IN
        );
        
        List<Room> allRooms = roomService.getAllRooms();
        
        return allRooms.stream()
                .filter(room -> {
                    List<Reservation> conflicts = reservationRepository.findConflictingReservations(
                            room, startTime, endTime, activeStatuses);
                    return conflicts.isEmpty();
                })
                .collect(Collectors.toList());
    }
    
    public Optional<Reservation> getReservationById(Long id) {
        autoReleaseExpiredReservations();
        return reservationRepository.findById(id);
    }
    
    private void validateTimeRange(LocalDateTime startTime, LocalDateTime endTime) {
        if (startTime.isAfter(endTime) || startTime.isEqual(endTime)) {
            throw new RuntimeException("开始时间必须早于结束时间");
        }
        
        long minutes = Duration.between(startTime, endTime).toMinutes();
        if (minutes < MIN_DURATION_MINUTES) {
            throw new RuntimeException("预约时长至少需要15分钟");
        }
        if (minutes > MAX_DURATION_MINUTES) {
            throw new RuntimeException("预约时长不能超过4小时");
        }
    }
    
    private void checkTimeConflict(Room room, LocalDateTime startTime, LocalDateTime endTime) {
        List<ReservationStatus> activeStatuses = Arrays.asList(
                ReservationStatus.BOOKED, 
                ReservationStatus.CHECKED_IN
        );
        
        List<Reservation> conflicts = reservationRepository.findConflictingReservations(
                room, startTime, endTime, activeStatuses);
        
        if (!conflicts.isEmpty()) {
            Reservation conflict = conflicts.get(0);
            throw new RuntimeException(String.format(
                    "时间冲突！该会议室在 %s - %s 已被预约：%s",
                    conflict.getStartTime(), conflict.getEndTime(), conflict.getTopic()));
        }
    }
}

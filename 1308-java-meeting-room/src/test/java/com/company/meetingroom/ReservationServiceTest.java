package com.company.meetingroom;

import com.company.meetingroom.model.Reservation;
import com.company.meetingroom.model.ReservationStatus;
import com.company.meetingroom.model.Room;
import com.company.meetingroom.repository.ReservationRepository;
import com.company.meetingroom.repository.RoomRepository;
import com.company.meetingroom.service.ReservationService;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.boot.test.context.SpringBootTest;
import org.springframework.test.annotation.DirtiesContext;

import java.time.LocalDateTime;
import java.util.List;

import static org.junit.jupiter.api.Assertions.*;

@SpringBootTest
@DirtiesContext(classMode = DirtiesContext.ClassMode.AFTER_EACH_TEST_METHOD)
public class ReservationServiceTest {

    @Autowired
    private ReservationService reservationService;

    @Autowired
    private RoomRepository roomRepository;

    @Autowired
    private ReservationRepository reservationRepository;

    private Room testRoom;

    @BeforeEach
    void setUp() {
        reservationRepository.deleteAll();
        roomRepository.deleteAll();
        
        Room room = new Room();
        room.setCode("TEST101");
        room.setName("测试会议室");
        room.setLocation("1楼101号");
        room.setCapacity(10);
        testRoom = roomRepository.save(room);
    }

    @Test
    void testCreateReservation_Success() {
        LocalDateTime start = LocalDateTime.now().plusHours(1);
        LocalDateTime end = start.plusHours(1);
        
        Reservation r = reservationService.createReservation(
                testRoom.getId(), start, end, "张三", "项目周会");
        
        assertNotNull(r);
        assertEquals(ReservationStatus.BOOKED, r.getStatus());
        assertEquals("张三", r.getBooker());
        assertEquals("项目周会", r.getTopic());
    }

    @Test
    void testCreateReservation_TooShort() {
        LocalDateTime start = LocalDateTime.now().plusHours(1);
        LocalDateTime end = start.plusMinutes(10);
        
        Exception ex = assertThrows(RuntimeException.class, () -> 
            reservationService.createReservation(testRoom.getId(), start, end, "张三", "短会"));
        assertTrue(ex.getMessage().contains("至少需要15分钟"));
    }

    @Test
    void testCreateReservation_TooLong() {
        LocalDateTime start = LocalDateTime.now().plusHours(1);
        LocalDateTime end = start.plusHours(5);
        
        Exception ex = assertThrows(RuntimeException.class, () -> 
            reservationService.createReservation(testRoom.getId(), start, end, "张三", "长会"));
        assertTrue(ex.getMessage().contains("不能超过4小时"));
    }

    @Test
    void testCreateReservation_TimeConflict_PartialOverlap() {
        LocalDateTime base = LocalDateTime.now().plusDays(1).withHour(14).withMinute(0);
        reservationService.createReservation(testRoom.getId(), 
                base, base.plusMinutes(90), "张三", "会议1");
        
        Exception ex = assertThrows(RuntimeException.class, () -> 
            reservationService.createReservation(testRoom.getId(),
                    base.plusMinutes(60), base.plusMinutes(120), "李四", "会议2"));
        assertTrue(ex.getMessage().contains("冲突"));
    }

    @Test
    void testCreateReservation_TimeConflict_ExactEnd() {
        LocalDateTime base = LocalDateTime.now().plusDays(1)
                .withHour(14).withMinute(0).withSecond(0).withNano(0);
        reservationService.createReservation(testRoom.getId(), 
                base, base.plusMinutes(60), "张三", "会议1");
        
        assertDoesNotThrow(() -> 
            reservationService.createReservation(testRoom.getId(),
                    base.plusMinutes(60), base.plusMinutes(120), "李四", "会议2"));
    }

    @Test
    void testCheckIn_Success() {
        LocalDateTime start = LocalDateTime.now().plusMinutes(5);
        Reservation r = reservationService.createReservation(
                testRoom.getId(), start, start.plusHours(1), "张三", "测试签到");
        
        Reservation checkedIn = reservationService.checkIn(r.getId());
        assertEquals(ReservationStatus.CHECKED_IN, checkedIn.getStatus());
        assertNotNull(checkedIn.getCheckInTime());
    }

    @Test
    void testCheckIn_TooEarly() {
        LocalDateTime start = LocalDateTime.now().plusHours(1);
        Reservation r = reservationService.createReservation(
                testRoom.getId(), start, start.plusHours(1), "张三", "测试");
        
        Exception ex = assertThrows(RuntimeException.class, () -> reservationService.checkIn(r.getId()));
        assertTrue(ex.getMessage().contains("尚未开始"));
    }

    @Test
    void testCancelReservation_Success() {
        LocalDateTime start = LocalDateTime.now().plusHours(2);
        Reservation r = reservationService.createReservation(
                testRoom.getId(), start, start.plusHours(1), "张三", "测试");
        
        Reservation cancelled = reservationService.cancelReservation(r.getId());
        assertEquals(ReservationStatus.CANCELLED, cancelled.getStatus());
    }

    @Test
    void testCancelReservation_AlreadyStarted() {
        LocalDateTime start = LocalDateTime.now().minusMinutes(30);
        Reservation r = reservationService.createReservation(
                testRoom.getId(), start, start.plusHours(1), "张三", "测试");
        
        Exception ex = assertThrows(RuntimeException.class, () -> reservationService.cancelReservation(r.getId()));
        assertTrue(ex.getMessage().contains("已经开始"));
    }

    @Test
    void testGetAvailableRooms() {
        LocalDateTime start = LocalDateTime.now().plusDays(1).withHour(10).withMinute(0);
        reservationService.createReservation(testRoom.getId(), 
                start, start.plusHours(2), "张三", "占用");
        
        List<Room> available = reservationService.getAvailableRooms(
                start.plusHours(3), start.plusHours(4));
        assertTrue(available.stream().anyMatch(r -> r.getId().equals(testRoom.getId())));
        
        List<Room> notAvailable = reservationService.getAvailableRooms(
                start.plusMinutes(30), start.plusMinutes(60));
        assertFalse(notAvailable.stream().anyMatch(r -> r.getId().equals(testRoom.getId())));
    }
}

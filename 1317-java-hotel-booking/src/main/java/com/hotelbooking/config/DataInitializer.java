package com.hotelbooking.config;

import com.hotelbooking.model.entity.*;
import com.hotelbooking.model.enums.*;
import com.hotelbooking.repository.*;
import lombok.RequiredArgsConstructor;
import org.springframework.boot.CommandLineRunner;
import org.springframework.security.crypto.password.PasswordEncoder;
import org.springframework.stereotype.Component;
import org.springframework.transaction.annotation.Transactional;

import java.math.BigDecimal;
import java.time.LocalDate;
import java.time.LocalDateTime;

@Component
@RequiredArgsConstructor
public class DataInitializer implements CommandLineRunner {

    private final RoomTypeConfigRepository roomTypeConfigRepository;
    private final RoomRepository roomRepository;
    private final ContinuousStayDiscountRepository continuousStayDiscountRepository;
    private final HolidayRepository holidayRepository;
    private final UserRepository userRepository;
    private final CustomerRepository customerRepository;
    private final PasswordEncoder passwordEncoder;

    @Override
    @Transactional
    public void run(String... args) {
        initializeRoomTypeConfigs();
        initializeRooms();
        initializeContinuousStayDiscounts();
        initializeHolidays();
        initializeUsers();
        initializeCustomers();
    }

    private void initializeRoomTypeConfigs() {
        if (roomTypeConfigRepository.count() > 0) return;

        roomTypeConfigRepository.save(RoomTypeConfig.builder()
                .roomType(RoomType.STANDARD_SINGLE)
                .roomCount(20)
                .basePrice(new BigDecimal("198.00"))
                .holidayEveSurcharge(new BigDecimal("0.15"))
                .holidaySurcharge(new BigDecimal("0.50"))
                .holidayAfterSurcharge(new BigDecimal("0.10"))
                .maxGuests(1)
                .description("标准单人间，适合单人入住")
                .build());

        roomTypeConfigRepository.save(RoomTypeConfig.builder()
                .roomType(RoomType.STANDARD_DOUBLE)
                .roomCount(30)
                .basePrice(new BigDecimal("258.00"))
                .holidayEveSurcharge(new BigDecimal("0.15"))
                .holidaySurcharge(new BigDecimal("0.50"))
                .holidayAfterSurcharge(new BigDecimal("0.10"))
                .maxGuests(2)
                .description("标准双人间，两张单人床")
                .build());

        roomTypeConfigRepository.save(RoomTypeConfig.builder()
                .roomType(RoomType.KING_BED)
                .roomCount(25)
                .basePrice(new BigDecimal("328.00"))
                .holidayEveSurcharge(new BigDecimal("0.20"))
                .holidaySurcharge(new BigDecimal("0.60"))
                .holidayAfterSurcharge(new BigDecimal("0.10"))
                .maxGuests(2)
                .description("大床房，一张特大床")
                .build());

        roomTypeConfigRepository.save(RoomTypeConfig.builder()
                .roomType(RoomType.FAMILY_ROOM)
                .roomCount(15)
                .basePrice(new BigDecimal("398.00"))
                .holidayEveSurcharge(new BigDecimal("0.20"))
                .holidaySurcharge(new BigDecimal("0.60"))
                .holidayAfterSurcharge(new BigDecimal("0.15"))
                .maxGuests(4)
                .description("家庭房，一张大床+两张单人床")
                .build());

        roomTypeConfigRepository.save(RoomTypeConfig.builder()
                .roomType(RoomType.SUITE)
                .roomCount(10)
                .basePrice(new BigDecimal("598.00"))
                .holidayEveSurcharge(new BigDecimal("0.25"))
                .holidaySurcharge(new BigDecimal("0.70"))
                .holidayAfterSurcharge(new BigDecimal("0.15"))
                .maxGuests(2)
                .description("豪华套房，独立客厅+卧室")
                .build());
    }

    private void initializeRooms() {
        if (roomRepository.count() > 0) return;

        int roomNumber = 101;
        
        for (int i = 0; i < 20; i++) {
            roomRepository.save(Room.builder()
                    .roomNumber(String.valueOf(roomNumber++))
                    .roomType(RoomType.STANDARD_SINGLE)
                    .floor(1)
                    .status(RoomStatus.AVAILABLE)
                    .build());
        }
        
        for (int i = 0; i < 30; i++) {
            roomRepository.save(Room.builder()
                    .roomNumber(String.valueOf(roomNumber++))
                    .roomType(RoomType.STANDARD_DOUBLE)
                    .floor(2)
                    .status(RoomStatus.AVAILABLE)
                    .build());
        }
        
        for (int i = 0; i < 25; i++) {
            roomRepository.save(Room.builder()
                    .roomNumber(String.valueOf(roomNumber++))
                    .roomType(RoomType.KING_BED)
                    .floor(3)
                    .status(RoomStatus.AVAILABLE)
                    .build());
        }
        
        for (int i = 0; i < 15; i++) {
            roomRepository.save(Room.builder()
                    .roomNumber(String.valueOf(roomNumber++))
                    .roomType(RoomType.FAMILY_ROOM)
                    .floor(4)
                    .status(RoomStatus.AVAILABLE)
                    .build());
        }
        
        for (int i = 0; i < 10; i++) {
            roomRepository.save(Room.builder()
                    .roomNumber(String.valueOf(roomNumber++))
                    .roomType(RoomType.SUITE)
                    .floor(5)
                    .status(RoomStatus.AVAILABLE)
                    .build());
        }
    }

    private void initializeContinuousStayDiscounts() {
        if (continuousStayDiscountRepository.count() > 0) return;

        continuousStayDiscountRepository.save(ContinuousStayDiscount.builder()
                .minDays(3)
                .maxDays(6)
                .discountRate(new BigDecimal("0.05"))
                .description("住3-6天打95折")
                .isActive(true)
                .build());

        continuousStayDiscountRepository.save(ContinuousStayDiscount.builder()
                .minDays(7)
                .maxDays(13)
                .discountRate(new BigDecimal("0.10"))
                .description("住7-13天打9折")
                .isActive(true)
                .build());

        continuousStayDiscountRepository.save(ContinuousStayDiscount.builder()
                .minDays(14)
                .maxDays(365)
                .discountRate(new BigDecimal("0.15"))
                .description("住14天及以上打85折")
                .isActive(true)
                .build());
    }

    private void initializeHolidays() {
        if (holidayRepository.count() > 0) return;

        LocalDate now = LocalDate.now();
        int year = now.getYear();

        holidayRepository.save(Holiday.builder()
                .name("元旦")
                .holidayDate(LocalDate.of(year, 1, 1))
                .holidayType(HolidayType.HOLIDAY)
                .description("元旦节")
                .build());

        holidayRepository.save(Holiday.builder()
                .name("春节")
                .holidayDate(LocalDate.of(year, 2, 10))
                .holidayType(HolidayType.HOLIDAY)
                .description("春节第一天")
                .build());

        holidayRepository.save(Holiday.builder()
                .name("春节")
                .holidayDate(LocalDate.of(year, 2, 9))
                .holidayType(HolidayType.HOLIDAY_EVE)
                .description("春节前一天")
                .build());

        holidayRepository.save(Holiday.builder()
                .name("劳动节")
                .holidayDate(LocalDate.of(year, 5, 1))
                .holidayType(HolidayType.HOLIDAY)
                .description("五一劳动节")
                .build());

        holidayRepository.save(Holiday.builder()
                .name("劳动节")
                .holidayDate(LocalDate.of(year, 4, 30))
                .holidayType(HolidayType.HOLIDAY_EVE)
                .description("劳动节前一天")
                .build());

        holidayRepository.save(Holiday.builder()
                .name("国庆节")
                .holidayDate(LocalDate.of(year, 10, 1))
                .holidayType(HolidayType.HOLIDAY)
                .description("国庆节")
                .build());

        holidayRepository.save(Holiday.builder()
                .name("国庆节")
                .holidayDate(LocalDate.of(year, 9, 30))
                .holidayType(HolidayType.HOLIDAY_EVE)
                .description("国庆节前一天")
                .build());

        holidayRepository.save(Holiday.builder()
                .name("中秋节")
                .holidayDate(LocalDate.of(year, 10, 6))
                .holidayType(HolidayType.HOLIDAY)
                .description("中秋节")
                .build());
    }

    private void initializeUsers() {
        if (userRepository.count() > 0) return;

        userRepository.save(User.builder()
                .username("admin")
                .password(passwordEncoder.encode("admin123"))
                .name("系统管理员")
                .phone("13800000001")
                .email("admin@hotel.com")
                .role(Role.ADMIN)
                .active(true)
                .build());

        userRepository.save(User.builder()
                .username("reception")
                .password(passwordEncoder.encode("reception123"))
                .name("前台小张")
                .phone("13800000002")
                .email("reception@hotel.com")
                .role(Role.RECEPTION)
                .active(true)
                .build());

        userRepository.save(User.builder()
                .username("housekeeping")
                .password(passwordEncoder.encode("housekeeping123"))
                .name("清洁员小李")
                .phone("13800000003")
                .email("housekeeping@hotel.com")
                .role(Role.HOUSEKEEPING)
                .active(true)
                .build());
    }

    private void initializeCustomers() {
        if (customerRepository.count() > 0) return;

        customerRepository.save(Customer.builder()
                .name("张三")
                .phone("13900000001")
                .email("zhangsan@example.com")
                .idCard("110101199001010001")
                .memberLevel(MemberLevel.GOLD)
                .totalStays(15)
                .totalSpent(new BigDecimal("15680.00"))
                .lastStayDate(LocalDate.now().minusDays(30))
                .build());

        customerRepository.save(Customer.builder()
                .name("李四")
                .phone("13900000002")
                .email("lisi@example.com")
                .idCard("110101199002020002")
                .memberLevel(MemberLevel.SILVER)
                .totalStays(8)
                .totalSpent(new BigDecimal("8920.00"))
                .lastStayDate(LocalDate.now().minusDays(15))
                .build());

        customerRepository.save(Customer.builder()
                .name("王五")
                .phone("13900000003")
                .email("wangwu@example.com")
                .idCard("110101199003030003")
                .memberLevel(MemberLevel.NONE)
                .totalStays(2)
                .totalSpent(new BigDecimal("1280.00"))
                .lastStayDate(LocalDate.now().minusDays(60))
                .build());
    }
}

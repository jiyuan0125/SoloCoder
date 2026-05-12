package com.hotelbooking.service;

import com.hotelbooking.dto.BookingCancellationResult;
import com.hotelbooking.dto.BookingRequest;
import com.hotelbooking.dto.BookingResponse;
import com.hotelbooking.dto.DailyPriceDetail;
import com.hotelbooking.dto.PriceCalculationResult;
import com.hotelbooking.exception.BookingException;
import com.hotelbooking.exception.ResourceNotFoundException;
import com.hotelbooking.model.entity.*;
import com.hotelbooking.model.enums.BookingStatus;
import com.hotelbooking.model.enums.RoomStatus;
import com.hotelbooking.repository.BookingDailyDetailRepository;
import com.hotelbooking.repository.BookingRepository;
import com.hotelbooking.repository.CustomerRepository;
import com.hotelbooking.repository.RoomRepository;
import lombok.RequiredArgsConstructor;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.math.BigDecimal;
import java.math.RoundingMode;
import java.time.LocalDateTime;
import java.time.temporal.ChronoUnit;
import java.util.List;
import java.util.Optional;
import java.util.stream.Collectors;

@Service
@RequiredArgsConstructor
public class BookingService {

    private final BookingRepository bookingRepository;
    private final BookingDailyDetailRepository bookingDailyDetailRepository;
    private final CustomerRepository customerRepository;
    private final RoomRepository roomRepository;
    private final InventoryService inventoryService;
    private final PriceService priceService;

    @Transactional
    public BookingResponse createBooking(BookingRequest request) {
        if (request.getCheckOutDate().isBefore(request.getCheckInDate()) ||
            request.getCheckOutDate().isEqual(request.getCheckInDate())) {
            throw new BookingException("退房日期必须晚于入住日期");
        }

        Customer customer = findOrCreateCustomer(request);
        PriceCalculationResult priceResult = priceService.calculatePrice(
                request.getRoomType(),
                request.getCheckInDate(),
                request.getCheckOutDate(),
                customer
        );

        Booking booking = Booking.builder()
                .customer(customer)
                .roomType(request.getRoomType())
                .checkInDate(request.getCheckInDate())
                .checkOutDate(request.getCheckOutDate())
                .numberOfGuests(request.getNumberOfGuests())
                .status(BookingStatus.PENDING_CONFIRMATION)
                .baseTotal(priceResult.getBaseTotal())
                .holidaySurchargeTotal(priceResult.getHolidaySurchargeTotal())
                .continuousStayDiscount(priceResult.getContinuousStayDiscount())
                .memberDiscount(priceResult.getMemberDiscount())
                .totalAmount(priceResult.getTotalAmount())
                .paidAmount(BigDecimal.ZERO)
                .cancellationFee(BigDecimal.ZERO)
                .earlyCheckOutFee(BigDecimal.ZERO)
                .specialRequests(request.getSpecialRequests())
                .isHotelCausedUpgrade(false)
                .build();

        inventoryService.validateAvailability(booking);
        
        booking = bookingRepository.save(booking);
        saveDailyDetails(booking, priceResult.getDailyDetails());

        return toResponse(booking);
    }

    @Transactional
    public BookingResponse confirmBooking(Long bookingId) {
        Booking booking = bookingRepository.findById(bookingId)
                .orElseThrow(() -> new ResourceNotFoundException("预订不存在"));

        if (booking.getStatus() != BookingStatus.PENDING_CONFIRMATION) {
            throw new BookingException("只能确认待确认状态的预订");
        }

        inventoryService.validateAvailability(booking);

        booking.setStatus(BookingStatus.CONFIRMED);
        booking = bookingRepository.save(booking);
        return toResponse(booking);
    }

    @Transactional
    public BookingResponse checkIn(Long bookingId) {
        Booking booking = bookingRepository.findById(bookingId)
                .orElseThrow(() -> new ResourceNotFoundException("预订不存在"));

        if (booking.getStatus() != BookingStatus.CONFIRMED) {
            throw new BookingException("只能为已确认的预订办理入住");
        }

        Room room = inventoryService.assignRoomToBooking(booking);
        booking.setRoom(room);
        booking.setStatus(BookingStatus.CHECKED_IN);
        booking.setCheckInTime(LocalDateTime.now());
        room.setStatus(RoomStatus.OCCUPIED);
        roomRepository.save(room);

        booking = bookingRepository.save(booking);
        return toResponse(booking);
    }

    @Transactional
    public BookingResponse checkOut(Long bookingId) {
        return checkOut(bookingId, null);
    }

    @Transactional
    public BookingResponse checkOut(Long bookingId, java.time.LocalDate actualCheckOutDate) {
        Booking booking = bookingRepository.findById(bookingId)
                .orElseThrow(() -> new ResourceNotFoundException("预订不存在"));

        if (booking.getStatus() != BookingStatus.CHECKED_IN) {
            throw new BookingException("只能为已入住的预订办理退房");
        }

        java.time.LocalDate checkOutDate = actualCheckOutDate != null ? actualCheckOutDate : booking.getCheckOutDate();
        BigDecimal earlyCheckOutFee = BigDecimal.ZERO;

        if (!checkOutDate.isEqual(booking.getCheckOutDate()) && checkOutDate.isBefore(booking.getCheckOutDate())) {
            booking.setOriginalCheckOutDate(booking.getCheckOutDate());
            BigDecimal avgDailyRate = booking.getBaseTotal()
                    .divide(BigDecimal.valueOf(
                            ChronoUnit.DAYS.between(booking.getCheckInDate(), booking.getCheckOutDate())),
                            2, RoundingMode.HALF_UP);
            earlyCheckOutFee = priceService.calculateEarlyCheckOutFee(
                    booking.getCheckOutDate(),
                    checkOutDate,
                    avgDailyRate
            );
            booking.setEarlyCheckOutFee(earlyCheckOutFee);
            booking.setCheckOutDate(checkOutDate);
        }

        if (booking.getRoom() != null) {
            Room room = booking.getRoom();
            room.setStatus(RoomStatus.CLEANING);
            room.setCleaningAvailableAt(LocalDateTime.now().plusHours(3));
            roomRepository.save(room);
        }

        booking.setStatus(BookingStatus.CHECKED_OUT);
        booking.setCheckOutTime(LocalDateTime.now());

        Customer customer = booking.getCustomer();
        customer.setTotalStays(customer.getTotalStays() + 1);
        customer.setTotalSpent(customer.getTotalSpent().add(booking.getTotalAmount()));
        customer.setLastStayDate(checkOutDate);
        customerRepository.save(customer);

        booking = bookingRepository.save(booking);
        return toResponse(booking);
    }

    @Transactional
    public BookingCancellationResult cancelBooking(Long bookingId) {
        Booking booking = bookingRepository.findById(bookingId)
                .orElseThrow(() -> new ResourceNotFoundException("预订不存在"));

        if (booking.getStatus() == BookingStatus.CANCELLED ||
            booking.getStatus() == BookingStatus.CHECKED_OUT) {
            throw new BookingException("该预订无法取消");
        }

        List<BookingDailyDetail> dailyDetails = bookingDailyDetailRepository.findByBookingId(bookingId);
        BigDecimal firstNightPrice = dailyDetails.isEmpty() ?
                booking.getTotalAmount() : dailyDetails.get(0).getDailyPrice();

        BigDecimal cancellationFee = priceService.calculateCancellationFee(
                booking.getCheckInDate(),
                firstNightPrice,
                LocalDateTime.now()
        );

        BigDecimal paidAmount = booking.getPaidAmount() != null ? booking.getPaidAmount() : BigDecimal.ZERO;
        BigDecimal refundAmount = paidAmount.subtract(cancellationFee);
        if (refundAmount.compareTo(BigDecimal.ZERO) < 0) {
            refundAmount = BigDecimal.ZERO;
        }

        if (booking.getRoom() != null) {
            Room room = booking.getRoom();
            if (room.getStatus() == RoomStatus.RESERVED) {
                room.setStatus(RoomStatus.AVAILABLE);
                roomRepository.save(room);
            }
        }

        booking.setStatus(BookingStatus.CANCELLED);
        booking.setCancellationFee(cancellationFee);
        booking.setCancelledAt(LocalDateTime.now());
        bookingRepository.save(booking);

        String message = cancellationFee.compareTo(BigDecimal.ZERO) == 0 ?
                "已免费取消预订" :
                "取消预订，收取违约金: " + cancellationFee;

        return BookingCancellationResult.builder()
                .success(true)
                .message(message)
                .cancellationFee(cancellationFee)
                .refundAmount(refundAmount)
                .build();
    }

    @Transactional
    public BookingResponse upgradeRoom(Long bookingId, com.hotelbooking.model.enums.RoomType newRoomType, boolean hotelCaused) {
        Booking booking = bookingRepository.findById(bookingId)
                .orElseThrow(() -> new ResourceNotFoundException("预订不存在"));

        if (booking.getStatus() == BookingStatus.CANCELLED ||
            booking.getStatus() == BookingStatus.CHECKED_OUT) {
            throw new BookingException("该预订无法升级房型");
        }

        if (booking.getRoomType() == newRoomType) {
            throw new BookingException("新房型与原房型相同");
        }

        booking.setUpgradedFromRoomType(booking.getRoomType());
        booking.setRoomType(newRoomType);
        booking.setIsHotelCausedUpgrade(hotelCaused);

        booking = bookingRepository.save(booking);
        return toResponse(booking);
    }

    public Optional<BookingResponse> findById(Long id) {
        return bookingRepository.findById(id).map(this::toResponse);
    }

    public Optional<BookingResponse> findByBookingNumber(String bookingNumber) {
        return bookingRepository.findByBookingNumber(bookingNumber).map(this::toResponse);
    }

    public List<BookingResponse> findByCustomerId(Long customerId) {
        return bookingRepository.findByCustomerId(customerId).stream()
                .map(this::toResponse)
                .collect(Collectors.toList());
    }

    public List<BookingResponse> findByStatus(BookingStatus status) {
        return bookingRepository.findByStatus(status).stream()
                .map(this::toResponse)
                .collect(Collectors.toList());
    }

    private BookingResponse toResponse(Booking booking) {
        BookingResponse.BookingResponseBuilder builder = BookingResponse.builder()
                .id(booking.getId())
                .bookingNumber(booking.getBookingNumber())
                .roomType(booking.getRoomType())
                .roomTypeName(booking.getRoomType().getDisplayName())
                .checkInDate(booking.getCheckInDate())
                .checkOutDate(booking.getCheckOutDate())
                .numberOfGuests(booking.getNumberOfGuests())
                .status(booking.getStatus())
                .statusName(booking.getStatus().getDisplayName())
                .originalCheckOutDate(booking.getOriginalCheckOutDate())
                .baseTotal(booking.getBaseTotal())
                .continuousStayDiscount(booking.getContinuousStayDiscount())
                .memberDiscount(booking.getMemberDiscount())
                .holidaySurchargeTotal(booking.getHolidaySurchargeTotal())
                .totalAmount(booking.getTotalAmount())
                .paidAmount(booking.getPaidAmount() != null ? booking.getPaidAmount() : BigDecimal.ZERO)
                .cancellationFee(booking.getCancellationFee() != null ? booking.getCancellationFee() : BigDecimal.ZERO)
                .earlyCheckOutFee(booking.getEarlyCheckOutFee() != null ? booking.getEarlyCheckOutFee() : BigDecimal.ZERO)
                .specialRequests(booking.getSpecialRequests())
                .upgradedFromRoomType(booking.getUpgradedFromRoomType())
                .isHotelCausedUpgrade(booking.getIsHotelCausedUpgrade() != null ? booking.getIsHotelCausedUpgrade() : false)
                .checkInTime(booking.getCheckInTime())
                .checkOutTime(booking.getCheckOutTime())
                .cancelledAt(booking.getCancelledAt())
                .createdAt(booking.getCreatedAt())
                .updatedAt(booking.getUpdatedAt());

        if (booking.getCustomer() != null) {
            builder.customerId(booking.getCustomer().getId())
                   .customerName(booking.getCustomer().getName())
                   .customerPhone(booking.getCustomer().getPhone());
        }

        if (booking.getRoom() != null) {
            builder.roomId(booking.getRoom().getId())
                   .roomNumber(booking.getRoom().getRoomNumber());
        }

        return builder.build();
    }

    private Customer findOrCreateCustomer(BookingRequest request) {
        Optional<Customer> existing = customerRepository.findByPhone(request.getCustomerPhone());
        if (existing.isPresent()) {
            Customer customer = existing.get();
            if (request.getCustomerName() != null && !request.getCustomerName().isEmpty()) {
                customer.setName(request.getCustomerName());
            }
            if (request.getCustomerEmail() != null && !request.getCustomerEmail().isEmpty()) {
                customer.setEmail(request.getCustomerEmail());
            }
            if (request.getCustomerIdCard() != null && !request.getCustomerIdCard().isEmpty()) {
                customer.setIdCard(request.getCustomerIdCard());
            }
            return customerRepository.save(customer);
        }

        return customerRepository.save(Customer.builder()
                .name(request.getCustomerName() != null ? request.getCustomerName() : "未命名客户")
                .phone(request.getCustomerPhone())
                .email(request.getCustomerEmail())
                .idCard(request.getCustomerIdCard())
                .memberLevel(com.hotelbooking.model.enums.MemberLevel.NONE)
                .totalStays(0)
                .totalSpent(java.math.BigDecimal.ZERO)
                .build());
    }

    private void saveDailyDetails(Booking booking, List<DailyPriceDetail> dailyDetails) {
        for (DailyPriceDetail detail : dailyDetails) {
            bookingDailyDetailRepository.save(BookingDailyDetail.builder()
                    .booking(booking)
                    .stayDate(detail.getStayDate())
                    .basePrice(detail.getBasePrice())
                    .holidayType(detail.getHolidayType())
                    .surchargeRate(detail.getSurchargeRate())
                    .surchargeAmount(detail.getSurchargeAmount())
                    .dailyPrice(detail.getDailyPrice())
                    .build());
        }
    }
}

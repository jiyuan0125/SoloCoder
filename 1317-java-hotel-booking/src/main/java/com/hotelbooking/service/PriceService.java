package com.hotelbooking.service;

import com.hotelbooking.dto.DailyPriceDetail;
import com.hotelbooking.dto.PriceCalculationResult;
import com.hotelbooking.exception.BookingException;
import com.hotelbooking.model.entity.ContinuousStayDiscount;
import com.hotelbooking.model.entity.Customer;
import com.hotelbooking.model.entity.Holiday;
import com.hotelbooking.model.entity.RoomTypeConfig;
import com.hotelbooking.model.enums.HolidayType;
import com.hotelbooking.model.enums.RoomType;
import com.hotelbooking.repository.ContinuousStayDiscountRepository;
import com.hotelbooking.repository.HolidayRepository;
import com.hotelbooking.repository.RoomTypeConfigRepository;
import lombok.RequiredArgsConstructor;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.math.BigDecimal;
import java.math.RoundingMode;
import java.time.LocalDate;
import java.time.LocalDateTime;
import java.time.temporal.ChronoUnit;
import java.util.ArrayList;
import java.util.List;
import java.util.Map;
import java.util.stream.Collectors;

@Service
@RequiredArgsConstructor
public class PriceService {

    private final RoomTypeConfigRepository roomTypeConfigRepository;
    private final HolidayRepository holidayRepository;
    private final ContinuousStayDiscountRepository continuousStayDiscountRepository;

    @Transactional(readOnly = true)
    public PriceCalculationResult calculatePrice(
            RoomType roomType,
            LocalDate checkInDate,
            LocalDate checkOutDate,
            Customer customer
    ) {
        validateDates(checkInDate, checkOutDate);

        RoomTypeConfig config = roomTypeConfigRepository.findByRoomType(roomType)
                .orElseThrow(() -> new BookingException("房型配置不存在"));

        int numberOfNights = (int) ChronoUnit.DAYS.between(checkInDate, checkOutDate);
        
        Map<LocalDate, Holiday> holidayMap = holidayRepository.findHolidaysInRange(checkInDate, checkOutDate.minusDays(1))
                .stream()
                .collect(Collectors.toMap(Holiday::getHolidayDate, h -> h));

        List<DailyPriceDetail> dailyDetails = new ArrayList<>();
        BigDecimal baseTotal = BigDecimal.ZERO;
        BigDecimal holidaySurchargeTotal = BigDecimal.ZERO;

        LocalDate current = checkInDate;
        for (int i = 0; i < numberOfNights; i++) {
            Holiday holiday = holidayMap.get(current);
            HolidayType holidayType = holiday != null ? holiday.getHolidayType() : HolidayType.NORMAL;
            BigDecimal surchargeRate = getSurchargeRate(config, holidayType);
            
            BigDecimal basePrice = config.getBasePrice();
            BigDecimal surchargeAmount = basePrice.multiply(surchargeRate).setScale(2, RoundingMode.HALF_UP);
            BigDecimal dailyPrice = basePrice.add(surchargeAmount);

            dailyDetails.add(DailyPriceDetail.builder()
                    .stayDate(current)
                    .holidayType(holidayType)
                    .basePrice(basePrice)
                    .surchargeRate(surchargeRate)
                    .surchargeAmount(surchargeAmount)
                    .dailyPrice(dailyPrice)
                    .build());

            baseTotal = baseTotal.add(basePrice);
            holidaySurchargeTotal = holidaySurchargeTotal.add(surchargeAmount);
            current = current.plusDays(1);
        }

        BigDecimal priceBeforeDiscounts = baseTotal.add(holidaySurchargeTotal);
        
        BigDecimal continuousStayDiscountRate = getContinuousStayDiscountRate(numberOfNights);
        BigDecimal continuousStayDiscount = priceBeforeDiscounts.multiply(continuousStayDiscountRate)
                .setScale(2, RoundingMode.HALF_UP);
        BigDecimal afterContinuousDiscount = priceBeforeDiscounts.subtract(continuousStayDiscount);

        BigDecimal memberDiscountRate = customer != null ? customer.getMemberLevel().getDiscountRate() : BigDecimal.ONE;
        memberDiscountRate = BigDecimal.ONE.subtract(memberDiscountRate);
        BigDecimal memberDiscount = afterContinuousDiscount.multiply(memberDiscountRate)
                .setScale(2, RoundingMode.HALF_UP);

        BigDecimal totalAmount = afterContinuousDiscount.subtract(memberDiscount);

        return PriceCalculationResult.builder()
                .baseTotal(baseTotal)
                .holidaySurchargeTotal(holidaySurchargeTotal)
                .continuousStayDiscount(continuousStayDiscount)
                .memberDiscount(memberDiscount)
                .totalAmount(totalAmount)
                .numberOfNights(numberOfNights)
                .dailyDetails(dailyDetails)
                .build();
    }

    private BigDecimal getSurchargeRate(RoomTypeConfig config, HolidayType holidayType) {
        return switch (holidayType) {
            case HOLIDAY -> config.getHolidaySurcharge();
            case HOLIDAY_EVE -> config.getHolidayEveSurcharge();
            case HOLIDAY_AFTER -> config.getHolidayAfterSurcharge();
            default -> BigDecimal.ZERO;
        };
    }

    private BigDecimal getContinuousStayDiscountRate(int numberOfNights) {
        return continuousStayDiscountRepository.findByDays(numberOfNights)
                .map(ContinuousStayDiscount::getDiscountRate)
                .orElse(BigDecimal.ZERO);
    }

    public BigDecimal calculateCancellationFee(
            LocalDate checkInDate,
            BigDecimal firstNightPrice,
            LocalDateTime cancellationTime
    ) {
        long hoursBeforeCheckIn = ChronoUnit.HOURS.between(cancellationTime, checkInDate.atStartOfDay());

        if (hoursBeforeCheckIn >= 48) {
            return BigDecimal.ZERO;
        } else if (hoursBeforeCheckIn >= 24) {
            return firstNightPrice.multiply(new BigDecimal("0.50"))
                    .setScale(2, RoundingMode.HALF_UP);
        } else {
            return firstNightPrice.setScale(2, RoundingMode.HALF_UP);
        }
    }

    public BigDecimal calculateEarlyCheckOutFee(
            LocalDate originalCheckOutDate,
            LocalDate actualCheckOutDate,
            BigDecimal averageDailyRate
    ) {
        if (actualCheckOutDate.isEqual(originalCheckOutDate) || actualCheckOutDate.isAfter(originalCheckOutDate)) {
            return BigDecimal.ZERO;
        }

        int unStayedNights = (int) ChronoUnit.DAYS.between(actualCheckOutDate, originalCheckOutDate);
        BigDecimal totalUnStayedPrice = averageDailyRate.multiply(BigDecimal.valueOf(unStayedNights));
        
        return totalUnStayedPrice.multiply(new BigDecimal("0.50"))
                .setScale(2, RoundingMode.HALF_UP);
    }

    private void validateDates(LocalDate checkInDate, LocalDate checkOutDate) {
        if (checkInDate == null || checkOutDate == null) {
            throw new BookingException("入住和退房日期不能为空");
        }
        
        if (checkOutDate.isBefore(checkInDate) || checkOutDate.isEqual(checkInDate)) {
            throw new BookingException("退房日期必须晚于入住日期");
        }
    }
}

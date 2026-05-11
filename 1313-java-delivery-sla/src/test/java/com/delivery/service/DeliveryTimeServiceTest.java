package com.delivery.service;

import com.delivery.entity.Holiday;
import com.delivery.enums.DeliveryType;
import com.delivery.repository.HolidayRepository;
import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.InjectMocks;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;
import org.mockito.junit.jupiter.MockitoSettings;
import org.mockito.quality.Strictness;

import java.time.LocalDate;
import java.time.LocalDateTime;
import java.util.Optional;

import static org.junit.jupiter.api.Assertions.*;
import static org.mockito.ArgumentMatchers.any;
import static org.mockito.Mockito.when;

@ExtendWith(MockitoExtension.class)
@MockitoSettings(strictness = Strictness.LENIENT)
class DeliveryTimeServiceTest {

    @Mock
    private HolidayRepository holidayRepository;

    @InjectMocks
    private DeliveryTimeService deliveryTimeService;

    @Test
    @DisplayName("截单时间 - 节假日前一天特殊截单时间")
    void testGetCutOffHour_HolidayEve() {
        LocalDateTime orderTime = LocalDateTime.of(2026, 9, 30, 10, 0);
        LocalDate nextDay = LocalDate.of(2026, 10, 1);
        
        Holiday holiday = Holiday.builder()
            .holidayDate(nextDay)
            .holidayName("国庆节")
            .specialCutOffHour(14)
            .build();
            
        when(holidayRepository.findByHolidayDate(any())).thenReturn(Optional.empty());
        when(holidayRepository.findByHolidayDate(nextDay)).thenReturn(Optional.of(holiday));
        
        int cutOffHour = deliveryTimeService.getCutOffHour(DeliveryType.NEXT_DAY, orderTime);
        
        assertEquals(14, cutOffHour);
    }

    @Test
    @DisplayName("截单时间 - 普通日期默认截单时间")
    void testGetCutOffHour_NormalDay() {
        LocalDateTime orderTime = LocalDateTime.of(2026, 5, 12, 10, 0);
        
        when(holidayRepository.findByHolidayDate(any())).thenReturn(Optional.empty());
        
        int cutOffHour = deliveryTimeService.getCutOffHour(DeliveryType.NEXT_DAY, orderTime);
        
        assertEquals(16, cutOffHour);
    }

    @Test
    @DisplayName("是否在截单时间之前")
    void testIsBeforeCutOff_True() {
        LocalDateTime orderTime = LocalDateTime.of(2026, 5, 12, 14, 0);
        
        when(holidayRepository.findByHolidayDate(any())).thenReturn(Optional.empty());
        
        boolean result = deliveryTimeService.isBeforeCutOff(DeliveryType.NEXT_DAY, orderTime);
        
        assertTrue(result);
    }

    @Test
    @DisplayName("是否在截单时间之后")
    void testIsBeforeCutOff_False() {
        LocalDateTime orderTime = LocalDateTime.of(2026, 5, 12, 17, 0);
        
        when(holidayRepository.findByHolidayDate(any())).thenReturn(Optional.empty());
        
        boolean result = deliveryTimeService.isBeforeCutOff(DeliveryType.NEXT_DAY, orderTime);
        
        assertFalse(result);
    }

    @Test
    @DisplayName("有效下单时间 - 超过截单时间顺延到下一个工作日")
    void testCalculateEffectiveOrderTime_AfterCutOff() {
        LocalDateTime orderTime = LocalDateTime.of(2026, 5, 12, 17, 0);
        
        when(holidayRepository.findByHolidayDate(any())).thenReturn(Optional.empty());
        
        LocalDateTime effectiveTime = deliveryTimeService.calculateEffectiveOrderTime(
            DeliveryType.NEXT_DAY, orderTime);
        
        LocalDateTime expected = LocalDateTime.of(2026, 5, 13, 0, 0);
        assertEquals(expected, effectiveTime);
    }

    @Test
    @DisplayName("有效下单时间 - 周五下午4点后下单顺延到下周一")
    void testCalculateEffectiveOrderTime_FridayAfterCutOff() {
        LocalDateTime orderTime = LocalDateTime.of(2026, 5, 15, 17, 0);
        
        when(holidayRepository.findByHolidayDate(any())).thenReturn(Optional.empty());
        
        LocalDateTime effectiveTime = deliveryTimeService.calculateEffectiveOrderTime(
            DeliveryType.NEXT_DAY, orderTime);
        
        LocalDateTime expected = LocalDateTime.of(2026, 5, 18, 0, 0);
        assertEquals(expected, effectiveTime);
    }

    @Test
    @DisplayName("承诺送达时间 - 次日达")
    void testCalculatePromisedDeliveryTime_NextDay() {
        LocalDateTime effectiveTime = LocalDateTime.of(2026, 5, 12, 10, 0);
        
        when(holidayRepository.findByHolidayDate(any())).thenReturn(Optional.empty());
        
        LocalDateTime promisedTime = deliveryTimeService.calculatePromisedDeliveryTime(
            DeliveryType.NEXT_DAY, effectiveTime);
        
        LocalDateTime expected = LocalDateTime.of(2026, 5, 13, 21, 0);
        assertEquals(expected, promisedTime);
    }

    @Test
    @DisplayName("承诺送达时间 - 隔日达")
    void testCalculatePromisedDeliveryTime_TwoDay() {
        LocalDateTime effectiveTime = LocalDateTime.of(2026, 5, 12, 10, 0);
        
        when(holidayRepository.findByHolidayDate(any())).thenReturn(Optional.empty());
        
        LocalDateTime promisedTime = deliveryTimeService.calculatePromisedDeliveryTime(
            DeliveryType.TWO_DAY, effectiveTime);
        
        LocalDateTime expected = LocalDateTime.of(2026, 5, 14, 21, 0);
        assertEquals(expected, promisedTime);
    }

    @Test
    @DisplayName("承诺送达时间 - 普通配送")
    void testCalculatePromisedDeliveryTime_Standard() {
        LocalDateTime effectiveTime = LocalDateTime.of(2026, 5, 12, 10, 0);
        
        when(holidayRepository.findByHolidayDate(any())).thenReturn(Optional.empty());
        
        LocalDateTime promisedTime = deliveryTimeService.calculatePromisedDeliveryTime(
            DeliveryType.STANDARD, effectiveTime);
        
        LocalDateTime expected = LocalDateTime.of(2026, 5, 15, 21, 0);
        assertEquals(expected, promisedTime);
    }

    @Test
    @DisplayName("承诺送达时间 - 周末顺延")
    void testCalculatePromisedDeliveryTime_WeekendSkip() {
        LocalDateTime effectiveTime = LocalDateTime.of(2026, 5, 15, 10, 0);
        
        when(holidayRepository.findByHolidayDate(any())).thenReturn(Optional.empty());
        
        LocalDateTime promisedTime = deliveryTimeService.calculatePromisedDeliveryTime(
            DeliveryType.NEXT_DAY, effectiveTime);
        
        LocalDateTime expected = LocalDateTime.of(2026, 5, 18, 21, 0);
        assertEquals(expected, promisedTime);
    }

    @Test
    @DisplayName("工作日判断 - 周一到周五是工作日")
    void testIsWorkingDay_Weekday() {
        LocalDate monday = LocalDate.of(2026, 5, 11);
        LocalDate friday = LocalDate.of(2026, 5, 15);
        
        when(holidayRepository.findByHolidayDate(monday)).thenReturn(Optional.empty());
        when(holidayRepository.findByHolidayDate(friday)).thenReturn(Optional.empty());
        
        assertTrue(deliveryTimeService.isWorkingDay(monday));
        assertTrue(deliveryTimeService.isWorkingDay(friday));
    }

    @Test
    @DisplayName("工作日判断 - 周末不是工作日")
    void testIsWorkingDay_Weekend() {
        LocalDate saturday = LocalDate.of(2026, 5, 16);
        LocalDate sunday = LocalDate.of(2026, 5, 17);
        
        assertFalse(deliveryTimeService.isWorkingDay(saturday));
        assertFalse(deliveryTimeService.isWorkingDay(sunday));
    }

    @Test
    @DisplayName("工作日判断 - 法定节假日不是工作日")
    void testIsWorkingDay_Holiday() {
        LocalDate holidayDate = LocalDate.of(2026, 10, 1);
        
        Holiday holiday = Holiday.builder()
            .holidayDate(holidayDate)
            .holidayName("国庆节")
            .build();
        
        when(holidayRepository.findByHolidayDate(holidayDate)).thenReturn(Optional.of(holiday));
        
        assertFalse(deliveryTimeService.isWorkingDay(holidayDate));
    }
}

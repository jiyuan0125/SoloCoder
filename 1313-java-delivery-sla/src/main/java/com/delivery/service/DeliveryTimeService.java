package com.delivery.service;

import com.delivery.entity.Holiday;
import com.delivery.enums.DeliveryType;
import com.delivery.repository.HolidayRepository;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.stereotype.Service;
import java.time.DayOfWeek;
import java.time.LocalDate;
import java.time.LocalDateTime;
import java.time.LocalTime;
import java.util.Optional;

@Service
@RequiredArgsConstructor
@Slf4j
public class DeliveryTimeService {

    private final HolidayRepository holidayRepository;

    public int getCutOffHour(DeliveryType deliveryType, LocalDateTime orderTime) {
        LocalDate orderDate = orderTime.toLocalDate();
        LocalDate nextDay = orderDate.plusDays(1);
        
        Optional<Holiday> nextDayHoliday = holidayRepository.findByHolidayDate(nextDay);
        if (nextDayHoliday.isPresent() && nextDayHoliday.get().getSpecialCutOffHour() != null) {
            log.debug("节假日前一天使用特殊截单时间: {}点", nextDayHoliday.get().getSpecialCutOffHour());
            return nextDayHoliday.get().getSpecialCutOffHour();
        }
        
        return deliveryType.getDefaultCutOffHour();
    }

    public boolean isBeforeCutOff(DeliveryType deliveryType, LocalDateTime orderTime) {
        int cutOffHour = getCutOffHour(deliveryType, orderTime);
        LocalTime cutOffTime = LocalTime.of(cutOffHour, 0);
        LocalTime orderTimeOfDay = orderTime.toLocalTime();
        
        return orderTimeOfDay.isBefore(cutOffTime);
    }

    public LocalDateTime calculateEffectiveOrderTime(DeliveryType deliveryType, LocalDateTime orderTime) {
        LocalDateTime effectiveTime = orderTime;
        
        while (true) {
            if (!isWorkingDay(effectiveTime.toLocalDate())) {
                effectiveTime = effectiveTime.plusDays(1);
                effectiveTime = LocalDateTime.of(effectiveTime.toLocalDate(), LocalTime.of(0, 0));
                continue;
            }
            
            if (isBeforeCutOff(deliveryType, effectiveTime)) {
                break;
            }
            
            effectiveTime = effectiveTime.plusDays(1);
            effectiveTime = LocalDateTime.of(effectiveTime.toLocalDate(), LocalTime.of(0, 0));
        }
        
        log.debug("下单时间: {}, 有效下单时间: {}", orderTime, effectiveTime);
        return effectiveTime;
    }

    public LocalDateTime calculatePromisedDeliveryTime(DeliveryType deliveryType, 
                                                        LocalDateTime effectiveOrderTime) {
        LocalDate deliveryDate = effectiveOrderTime.toLocalDate();
        int deliveryDays = 0;
        int daysToAdd = getDeliveryDays(deliveryType);
        
        while (deliveryDays < daysToAdd) {
            deliveryDate = deliveryDate.plusDays(1);
            if (isWorkingDay(deliveryDate)) {
                deliveryDays++;
            }
        }
        
        LocalTime endTime = LocalTime.of(21, 0);
        LocalDateTime promisedTime = LocalDateTime.of(deliveryDate, endTime);
        
        log.debug("有效下单时间: {}, 承诺送达时间: {}", effectiveOrderTime, promisedTime);
        return promisedTime;
    }

    private int getDeliveryDays(DeliveryType deliveryType) {
        return switch (deliveryType) {
            case NEXT_DAY -> 1;
            case TWO_DAY -> 2;
            case STANDARD -> 3;
        };
    }

    public boolean isWorkingDay(LocalDate date) {
        DayOfWeek dayOfWeek = date.getDayOfWeek();
        
        if (dayOfWeek == DayOfWeek.SATURDAY || dayOfWeek == DayOfWeek.SUNDAY) {
            return false;
        }
        
        Optional<Holiday> holiday = holidayRepository.findByHolidayDate(date);
        return holiday.isEmpty();
    }

    public boolean isHoliday(LocalDate date) {
        Optional<Holiday> holiday = holidayRepository.findByHolidayDate(date);
        return holiday.isPresent();
    }
}

package com.delivery.config;

import lombok.Data;
import org.springframework.boot.context.properties.ConfigurationProperties;
import org.springframework.stereotype.Component;

@Component
@ConfigurationProperties(prefix = "delivery")
@Data
public class DeliveryProperties {
    
    private FeeConfig fee = new FeeConfig();
    private TimeSlotConfig timeSlot = new TimeSlotConfig();
    private CompensationConfig compensation = new CompensationConfig();
    private DiscountConfig discount = new DiscountConfig();
    
    @Data
    public static class FeeConfig {
        private int baseDistance = 3;
        private double baseFee = 15.0;
        private double perKmExtra = 2.0;
        private double baseWeight = 5.0;
        private double perKgExtra = 1.0;
        private double minimumFee = 10.0;
        private OversizedConfig oversized = new OversizedConfig();
    }
    
    @Data
    public static class OversizedConfig {
        private double maxLength = 1.5;
        private double maxWeight = 30.0;
        private double extraFee = 50.0;
    }
    
    @Data
    public static class TimeSlotConfig {
        private int morningStart = 9;
        private int morningEnd = 12;
        private int afternoonStart = 13;
        private int afternoonEnd = 18;
        private int eveningStart = 18;
        private int eveningEnd = 21;
    }
    
    @Data
    public static class CompensationConfig {
        private double within2Hours = 0.3;
        private double within12Hours = 0.6;
        private double over12Hours = 1.0;
    }
    
    @Data
    public static class DiscountConfig {
        private int tier1Threshold = 10;
        private double tier1Discount = 0.9;
        private int tier2Threshold = 30;
        private double tier2Discount = 0.8;
    }
}

package com.delivery.service;

import com.delivery.config.DeliveryProperties;
import com.delivery.entity.DeliveryOrder;
import com.delivery.entity.Customer;
import com.delivery.enums.DeliveryType;
import com.delivery.repository.DeliveryOrderRepository;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.stereotype.Service;
import java.math.BigDecimal;
import java.math.RoundingMode;
import java.time.LocalDateTime;
import java.time.YearMonth;

@Service
@RequiredArgsConstructor
@Slf4j
public class FeeCalculationService {

    private final DeliveryProperties deliveryProperties;
    private final DeliveryOrderRepository orderRepository;

    public static class FeeBreakdown {
        private BigDecimal baseFee;
        private BigDecimal distanceSurcharge;
        private BigDecimal weightSurcharge;
        private BigDecimal deliveryTypeSurcharge;
        private BigDecimal oversizedFee;
        private BigDecimal discount;
        private BigDecimal totalFee;
        private int monthlyOrderCount;

        public FeeBreakdown() {
            this.baseFee = BigDecimal.ZERO;
            this.distanceSurcharge = BigDecimal.ZERO;
            this.weightSurcharge = BigDecimal.ZERO;
            this.deliveryTypeSurcharge = BigDecimal.ZERO;
            this.oversizedFee = BigDecimal.ZERO;
            this.discount = BigDecimal.ZERO;
            this.totalFee = BigDecimal.ZERO;
            this.monthlyOrderCount = 0;
        }

        public BigDecimal getBaseFee() { return baseFee; }
        public void setBaseFee(BigDecimal baseFee) { this.baseFee = baseFee; }
        public BigDecimal getDistanceSurcharge() { return distanceSurcharge; }
        public void setDistanceSurcharge(BigDecimal distanceSurcharge) { this.distanceSurcharge = distanceSurcharge; }
        public BigDecimal getWeightSurcharge() { return weightSurcharge; }
        public void setWeightSurcharge(BigDecimal weightSurcharge) { this.weightSurcharge = weightSurcharge; }
        public BigDecimal getDeliveryTypeSurcharge() { return deliveryTypeSurcharge; }
        public void setDeliveryTypeSurcharge(BigDecimal deliveryTypeSurcharge) { this.deliveryTypeSurcharge = deliveryTypeSurcharge; }
        public BigDecimal getOversizedFee() { return oversizedFee; }
        public void setOversizedFee(BigDecimal oversizedFee) { this.oversizedFee = oversizedFee; }
        public BigDecimal getDiscount() { return discount; }
        public void setDiscount(BigDecimal discount) { this.discount = discount; }
        public BigDecimal getTotalFee() { return totalFee; }
        public void setTotalFee(BigDecimal totalFee) { this.totalFee = totalFee; }
        public int getMonthlyOrderCount() { return monthlyOrderCount; }
        public void setMonthlyOrderCount(int monthlyOrderCount) { this.monthlyOrderCount = monthlyOrderCount; }
    }

    public FeeBreakdown calculateFee(Customer customer,
                                      DeliveryType deliveryType,
                                      BigDecimal weight,
                                      BigDecimal distance,
                                      Double length,
                                      Double width,
                                      Double height) {
        FeeBreakdown breakdown = new FeeBreakdown();
        DeliveryProperties.FeeConfig feeConfig = deliveryProperties.getFee();
        
        BigDecimal baseFee = calculateBaseFee(distance, feeConfig);
        breakdown.setBaseFee(baseFee);
        
        BigDecimal deliveryTypeSurcharge = calculateDeliveryTypeSurcharge(baseFee, deliveryType);
        breakdown.setDeliveryTypeSurcharge(deliveryTypeSurcharge);
        
        BigDecimal weightSurcharge = calculateWeightSurcharge(weight, feeConfig);
        breakdown.setWeightSurcharge(weightSurcharge);
        
        BigDecimal oversizedFee = calculateOversizedFee(weight, length, width, height, feeConfig);
        breakdown.setOversizedFee(oversizedFee);
        
        BigDecimal subtotal = baseFee.add(deliveryTypeSurcharge)
                                     .add(weightSurcharge)
                                     .add(oversizedFee);
        
        int monthlyOrderCount = getMonthlyOrderCount(customer.getId());
        breakdown.setMonthlyOrderCount(monthlyOrderCount);
        
        BigDecimal discount = calculateDiscount(subtotal, monthlyOrderCount + 1);
        breakdown.setDiscount(discount);
        
        BigDecimal totalBeforeMinimum = subtotal.subtract(discount);
        BigDecimal minimumFee = BigDecimal.valueOf(feeConfig.getMinimumFee());
        BigDecimal totalFee = totalBeforeMinimum.compareTo(minimumFee) > 0 
                            ? totalBeforeMinimum : minimumFee;
        
        breakdown.setTotalFee(totalFee.setScale(2, RoundingMode.HALF_UP));
        
        log.info("费用计算完成 - 客户: {}, 总运费: {}", customer.getCustomerCode(), breakdown.getTotalFee());
        return breakdown;
    }

    private BigDecimal calculateBaseFee(BigDecimal distance, DeliveryProperties.FeeConfig config) {
        BigDecimal baseDistance = BigDecimal.valueOf(config.getBaseDistance());
        BigDecimal baseFee = BigDecimal.valueOf(config.getBaseFee());
        BigDecimal perKmExtra = BigDecimal.valueOf(config.getPerKmExtra());
        
        if (distance.compareTo(baseDistance) <= 0) {
            return baseFee;
        }
        
        BigDecimal extraDistance = distance.subtract(baseDistance);
        BigDecimal distanceSurcharge = extraDistance.multiply(perKmExtra);
        
        return baseFee.add(distanceSurcharge).setScale(2, RoundingMode.HALF_UP);
    }

    private BigDecimal calculateDeliveryTypeSurcharge(BigDecimal baseFee, DeliveryType deliveryType) {
        BigDecimal surchargeRate = BigDecimal.valueOf(deliveryType.getSurchargeRate());
        return baseFee.multiply(surchargeRate).setScale(2, RoundingMode.HALF_UP);
    }

    private BigDecimal calculateWeightSurcharge(BigDecimal weight, DeliveryProperties.FeeConfig config) {
        BigDecimal baseWeight = BigDecimal.valueOf(config.getBaseWeight());
        
        if (weight.compareTo(baseWeight) <= 0) {
            return BigDecimal.ZERO;
        }
        
        BigDecimal extraWeight = weight.subtract(baseWeight);
        BigDecimal perKgExtra = BigDecimal.valueOf(config.getPerKgExtra());
        
        return extraWeight.multiply(perKgExtra).setScale(2, RoundingMode.HALF_UP);
    }

    private BigDecimal calculateOversizedFee(BigDecimal weight, 
                                              Double length, 
                                              Double width, 
                                              Double height,
                                              DeliveryProperties.FeeConfig config) {
        DeliveryProperties.OversizedConfig oversizedConfig = config.getOversized();
        boolean isOversized = false;
        
        if (weight.compareTo(BigDecimal.valueOf(oversizedConfig.getMaxWeight())) > 0) {
            isOversized = true;
        }
        
        if (length != null && length > oversizedConfig.getMaxLength()) {
            isOversized = true;
        }
        if (width != null && width > oversizedConfig.getMaxLength()) {
            isOversized = true;
        }
        if (height != null && height > oversizedConfig.getMaxLength()) {
            isOversized = true;
        }
        
        return isOversized ? BigDecimal.valueOf(oversizedConfig.getExtraFee()) : BigDecimal.ZERO;
    }

    public int getMonthlyOrderCount(Long customerId) {
        YearMonth currentMonth = YearMonth.now();
        LocalDateTime startOfMonth = currentMonth.atDay(1).atStartOfDay();
        LocalDateTime endOfMonth = currentMonth.plusMonths(1).atDay(1).atStartOfDay();
        
        return orderRepository.countMonthlyOrders(customerId, startOfMonth, endOfMonth);
    }

    private BigDecimal calculateDiscount(BigDecimal subtotal, int currentOrderCount) {
        DeliveryProperties.DiscountConfig discountConfig = deliveryProperties.getDiscount();
        
        if (currentOrderCount > discountConfig.getTier2Threshold()) {
            BigDecimal discountRate = BigDecimal.valueOf(1.0 - discountConfig.getTier2Discount());
            return subtotal.multiply(discountRate).setScale(2, RoundingMode.HALF_UP);
        }
        
        if (currentOrderCount > discountConfig.getTier1Threshold()) {
            BigDecimal discountRate = BigDecimal.valueOf(1.0 - discountConfig.getTier1Discount());
            return subtotal.multiply(discountRate).setScale(2, RoundingMode.HALF_UP);
        }
        
        return BigDecimal.ZERO;
    }

    public BigDecimal calculateCancellationFee(DeliveryOrder order) {
        BigDecimal totalFee = order.getTotalFee();
        
        return switch (order.getStatus()) {
            case PENDING, CONFIRMED -> BigDecimal.ZERO;
            case DISPATCHED, IN_TRANSIT -> totalFee.multiply(BigDecimal.valueOf(0.5))
                                                    .setScale(2, RoundingMode.HALF_UP);
            case DELIVERED -> totalFee;
            default -> BigDecimal.ZERO;
        };
    }

    public BigDecimal calculateDiscountRate(int monthlyOrderCount) {
        DeliveryProperties.DiscountConfig discountConfig = deliveryProperties.getDiscount();
        int nextOrderCount = monthlyOrderCount + 1;
        
        if (nextOrderCount > discountConfig.getTier2Threshold()) {
            return BigDecimal.valueOf(discountConfig.getTier2Discount());
        }
        
        if (nextOrderCount > discountConfig.getTier1Threshold()) {
            return BigDecimal.valueOf(discountConfig.getTier1Discount());
        }
        
        return BigDecimal.ONE;
    }
}

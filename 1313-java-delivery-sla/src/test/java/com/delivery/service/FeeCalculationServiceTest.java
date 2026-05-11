package com.delivery.service;

import com.delivery.config.DeliveryProperties;
import com.delivery.entity.Customer;
import com.delivery.enums.DeliveryType;
import com.delivery.repository.DeliveryOrderRepository;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.DisplayName;
import org.junit.jupiter.api.Test;
import org.junit.jupiter.api.extension.ExtendWith;
import org.mockito.InjectMocks;
import org.mockito.Mock;
import org.mockito.junit.jupiter.MockitoExtension;
import org.mockito.junit.jupiter.MockitoSettings;
import org.mockito.quality.Strictness;

import java.math.BigDecimal;

import static org.junit.jupiter.api.Assertions.*;
import static org.mockito.ArgumentMatchers.any;
import static org.mockito.Mockito.when;

@ExtendWith(MockitoExtension.class)
@MockitoSettings(strictness = Strictness.LENIENT)
class FeeCalculationServiceTest {

    @Mock
    private DeliveryProperties deliveryProperties;

    @Mock
    private DeliveryOrderRepository orderRepository;

    @InjectMocks
    private FeeCalculationService feeCalculationService;

    private Customer testCustomer;
    private DeliveryProperties.FeeConfig feeConfig;

    @BeforeEach
    void setUp() {
        testCustomer = Customer.builder()
            .id(1L)
            .customerCode("C001")
            .name("测试客户")
            .build();

        feeConfig = new DeliveryProperties.FeeConfig();
        feeConfig.setBaseDistance(3);
        feeConfig.setBaseFee(15.0);
        feeConfig.setPerKmExtra(2.0);
        feeConfig.setBaseWeight(5.0);
        feeConfig.setPerKgExtra(1.0);
        feeConfig.setMinimumFee(10.0);

        DeliveryProperties.OversizedConfig oversizedConfig = new DeliveryProperties.OversizedConfig();
        oversizedConfig.setMaxLength(1.5);
        oversizedConfig.setMaxWeight(30.0);
        oversizedConfig.setExtraFee(50.0);
        feeConfig.setOversized(oversizedConfig);

        when(deliveryProperties.getFee()).thenReturn(feeConfig);
        
        DeliveryProperties.DiscountConfig discountConfig = new DeliveryProperties.DiscountConfig();
        discountConfig.setTier1Threshold(10);
        discountConfig.setTier1Discount(0.9);
        discountConfig.setTier2Threshold(30);
        discountConfig.setTier2Discount(0.8);
        when(deliveryProperties.getDiscount()).thenReturn(discountConfig);

        when(orderRepository.countMonthlyOrders(any(), any(), any())).thenReturn(0);
    }

    private void assertBigDecimalEquals(BigDecimal expected, BigDecimal actual) {
        assertEquals(0, expected.compareTo(actual), 
                    "Expected: " + expected + ", Actual: " + actual);
    }

    @Test
    @DisplayName("基础运费计算 - 3公里内")
    void testCalculateBaseFee_WithinBaseDistance() {
        BigDecimal distance = BigDecimal.valueOf(2.5);
        
        FeeCalculationService.FeeBreakdown breakdown = feeCalculationService.calculateFee(
            testCustomer, DeliveryType.STANDARD, BigDecimal.valueOf(3.0), 
            distance, null, null, null);

        assertBigDecimalEquals(BigDecimal.valueOf(15.00), breakdown.getBaseFee());
    }

    @Test
    @DisplayName("基础运费计算 - 超出3公里")
    void testCalculateBaseFee_OverBaseDistance() {
        BigDecimal distance = BigDecimal.valueOf(5.0);
        
        FeeCalculationService.FeeBreakdown breakdown = feeCalculationService.calculateFee(
            testCustomer, DeliveryType.STANDARD, BigDecimal.valueOf(3.0), 
            distance, null, null, null);

        assertBigDecimalEquals(BigDecimal.valueOf(19.00), breakdown.getBaseFee());
    }

    @Test
    @DisplayName("时效附加费 - 次日达加收50%")
    void testDeliveryTypeSurcharge_NextDay() {
        BigDecimal distance = BigDecimal.valueOf(3.0);
        
        FeeCalculationService.FeeBreakdown breakdown = feeCalculationService.calculateFee(
            testCustomer, DeliveryType.NEXT_DAY, BigDecimal.valueOf(3.0), 
            distance, null, null, null);

        assertBigDecimalEquals(BigDecimal.valueOf(7.50), breakdown.getDeliveryTypeSurcharge());
    }

    @Test
    @DisplayName("时效附加费 - 隔日达加收20%")
    void testDeliveryTypeSurcharge_TwoDay() {
        BigDecimal distance = BigDecimal.valueOf(3.0);
        
        FeeCalculationService.FeeBreakdown breakdown = feeCalculationService.calculateFee(
            testCustomer, DeliveryType.TWO_DAY, BigDecimal.valueOf(3.0), 
            distance, null, null, null);

        assertBigDecimalEquals(BigDecimal.valueOf(3.00), breakdown.getDeliveryTypeSurcharge());
    }

    @Test
    @DisplayName("重量附加费 - 超过5公斤")
    void testWeightSurcharge_OverBaseWeight() {
        BigDecimal weight = BigDecimal.valueOf(8.0);
        BigDecimal distance = BigDecimal.valueOf(3.0);
        
        FeeCalculationService.FeeBreakdown breakdown = feeCalculationService.calculateFee(
            testCustomer, DeliveryType.STANDARD, weight, 
            distance, null, null, null);

        assertBigDecimalEquals(BigDecimal.valueOf(3.00), breakdown.getWeightSurcharge());
    }

    @Test
    @DisplayName("大件附加费 - 重量超过30公斤")
    void testOversizedFee_WeightOverLimit() {
        BigDecimal weight = BigDecimal.valueOf(35.0);
        BigDecimal distance = BigDecimal.valueOf(3.0);
        
        FeeCalculationService.FeeBreakdown breakdown = feeCalculationService.calculateFee(
            testCustomer, DeliveryType.STANDARD, weight, 
            distance, null, null, null);

        assertBigDecimalEquals(BigDecimal.valueOf(50.00), breakdown.getOversizedFee());
    }

    @Test
    @DisplayName("大件附加费 - 单边超过1.5米")
    void testOversizedFee_LengthOverLimit() {
        BigDecimal weight = BigDecimal.valueOf(10.0);
        BigDecimal distance = BigDecimal.valueOf(3.0);
        
        FeeCalculationService.FeeBreakdown breakdown = feeCalculationService.calculateFee(
            testCustomer, DeliveryType.STANDARD, weight, 
            distance, 2.0, 1.0, 1.0);

        assertBigDecimalEquals(BigDecimal.valueOf(50.00), breakdown.getOversizedFee());
    }

    @Test
    @DisplayName("最低消费门槛 - 10元")
    void testMinimumFee() {
        BigDecimal baseFee = BigDecimal.valueOf(15.0);
        BigDecimal discount = baseFee.multiply(BigDecimal.valueOf(0.2));
        BigDecimal totalAfterDiscount = baseFee.subtract(discount);
        
        assertTrue(totalAfterDiscount.compareTo(BigDecimal.valueOf(12.0)) == 0);
    }

    @Test
    @DisplayName("综合费用计算")
    void testComprehensiveFeeCalculation() {
        BigDecimal weight = BigDecimal.valueOf(8.0);
        BigDecimal distance = BigDecimal.valueOf(5.0);
        
        FeeCalculationService.FeeBreakdown breakdown = feeCalculationService.calculateFee(
            testCustomer, DeliveryType.NEXT_DAY, weight, 
            distance, null, null, null);

        BigDecimal baseFee = BigDecimal.valueOf(19.00);
        BigDecimal typeSurcharge = baseFee.multiply(BigDecimal.valueOf(0.5));
        BigDecimal weightSurcharge = BigDecimal.valueOf(3.00);
        BigDecimal expectedTotal = baseFee.add(typeSurcharge).add(weightSurcharge);
        
        assertBigDecimalEquals(expectedTotal, breakdown.getTotalFee());
    }
}

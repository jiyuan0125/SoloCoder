package com.coupon;

import com.coupon.dto.CouponCalculationResult;
import com.coupon.dto.CreateCouponTemplateRequest;
import com.coupon.dto.UseCouponRequest;
import com.coupon.entity.CouponTemplate;
import com.coupon.entity.UserCoupon;
import com.coupon.enums.CouponStatus;
import com.coupon.enums.CouponType;
import com.coupon.exception.CouponException;
import com.coupon.repository.UserCouponRepository;
import com.coupon.service.*;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.boot.test.context.SpringBootTest;
import org.springframework.transaction.annotation.Transactional;

import java.math.BigDecimal;
import java.time.LocalDateTime;
import java.util.Arrays;
import java.util.List;
import java.util.Map;

import static org.junit.jupiter.api.Assertions.*;

@SpringBootTest
@Transactional
class CouponEngineApplicationTests {

    @Autowired
    private CouponTemplateService templateService;

    @Autowired
    private CouponIssueService issueService;

    @Autowired
    private CouponUsageService usageService;

    @Autowired
    private CouponQueryService queryService;

    @Autowired
    private UserCouponRepository userCouponRepository;

    private CouponTemplate fullReductionTemplate;
    private CouponTemplate discountTemplate;
    private CouponTemplate freeShippingTemplate;

    @BeforeEach
    void setUp() {
        CreateCouponTemplateRequest fullReductionRequest = new CreateCouponTemplateRequest();
        fullReductionRequest.setName("满100减20");
        fullReductionRequest.setType(CouponType.FULL_REDUCTION);
        fullReductionRequest.setValue(new BigDecimal("20"));
        fullReductionRequest.setThreshold(new BigDecimal("100"));
        fullReductionRequest.setStartTime(LocalDateTime.now().minusDays(1));
        fullReductionRequest.setEndTime(LocalDateTime.now().plusDays(30));
        fullReductionRequest.setTotalQuantity(100);
        fullReductionRequest.setLimitPerUser(2);
        fullReductionTemplate = templateService.createTemplate(fullReductionRequest);

        CreateCouponTemplateRequest discountRequest = new CreateCouponTemplateRequest();
        discountRequest.setName("8折优惠券");
        discountRequest.setType(CouponType.DISCOUNT);
        discountRequest.setValue(new BigDecimal("8"));
        discountRequest.setThreshold(new BigDecimal("0"));
        discountRequest.setStartTime(LocalDateTime.now().minusDays(1));
        discountRequest.setEndTime(LocalDateTime.now().plusDays(30));
        discountRequest.setTotalQuantity(100);
        discountRequest.setLimitPerUser(2);
        discountTemplate = templateService.createTemplate(discountRequest);

        CreateCouponTemplateRequest freeShippingRequest = new CreateCouponTemplateRequest();
        freeShippingRequest.setName("免邮券");
        freeShippingRequest.setType(CouponType.FREE_SHIPPING);
        freeShippingRequest.setValue(BigDecimal.ZERO);
        freeShippingRequest.setThreshold(new BigDecimal("0"));
        freeShippingRequest.setStartTime(LocalDateTime.now().minusDays(1));
        freeShippingRequest.setEndTime(LocalDateTime.now().plusDays(30));
        freeShippingRequest.setTotalQuantity(100);
        freeShippingRequest.setLimitPerUser(5);
        freeShippingTemplate = templateService.createTemplate(freeShippingRequest);
    }

    @Test
    void testCreateTemplate() {
        assertNotNull(fullReductionTemplate.getId());
        assertEquals("满100减20", fullReductionTemplate.getName());
        assertEquals(CouponType.FULL_REDUCTION, fullReductionTemplate.getType());
        assertEquals(new BigDecimal("20"), fullReductionTemplate.getValue());
        assertEquals(new BigDecimal("100"), fullReductionTemplate.getThreshold());
    }

    @Test
    void testClaimCoupon() {
        Long userId = 1L;
        UserCoupon userCoupon = issueService.issueCoupon(userId, fullReductionTemplate.getId());
        assertNotNull(userCoupon.getId());
        assertEquals(userId, userCoupon.getUserId());
        assertEquals(CouponStatus.AVAILABLE, userCoupon.getStatus());
        assertEquals(fullReductionTemplate.getId(), userCoupon.getCouponTemplate().getId());
    }

    @Test
    void testClaimCouponExceedLimit() {
        Long userId = 2L;
        issueService.issueCoupon(userId, fullReductionTemplate.getId());
        issueService.issueCoupon(userId, fullReductionTemplate.getId());
        CouponException exception = assertThrows(CouponException.class,
                () -> issueService.issueCoupon(userId, fullReductionTemplate.getId()));
        assertEquals("您已达到该优惠券的领取上限", exception.getMessage());
    }

    @Test
    void testUseFullReductionCoupon() {
        Long userId = 3L;
        UserCoupon userCoupon = issueService.issueCoupon(userId, fullReductionTemplate.getId());

        UseCouponRequest request = new UseCouponRequest();
        request.setUserId(userId);
        request.setUserCouponIds(Arrays.asList(userCoupon.getId()));
        request.setOrderAmount(new BigDecimal("150"));
        request.setShippingFee(new BigDecimal("10"));
        request.setOrderId("ORDER_001");

        CouponCalculationResult result = usageService.useCoupons(request);

        assertEquals(new BigDecimal("140"), result.getFinalAmount());
        assertEquals(new BigDecimal("20"), result.getProductDiscount());
        assertEquals(BigDecimal.ZERO, result.getShippingDiscount());

        UserCoupon usedCoupon = userCouponRepository.findById(userCoupon.getId()).orElseThrow();
        assertEquals(CouponStatus.USED, usedCoupon.getStatus());
        assertEquals("ORDER_001", usedCoupon.getOrderId());
    }

    @Test
    void testUseFullReductionWithFreeShipping() {
        Long userId = 4L;
        UserCoupon fullReduction = issueService.issueCoupon(userId, fullReductionTemplate.getId());
        UserCoupon freeShipping = issueService.issueCoupon(userId, freeShippingTemplate.getId());

        UseCouponRequest request = new UseCouponRequest();
        request.setUserId(userId);
        request.setUserCouponIds(Arrays.asList(fullReduction.getId(), freeShipping.getId()));
        request.setOrderAmount(new BigDecimal("150"));
        request.setShippingFee(new BigDecimal("10"));
        request.setOrderId("ORDER_002");

        CouponCalculationResult result = usageService.useCoupons(request);

        assertEquals(new BigDecimal("130"), result.getFinalAmount());
        assertEquals(new BigDecimal("20"), result.getProductDiscount());
        assertEquals(new BigDecimal("10"), result.getShippingDiscount());
    }

    @Test
    void testUseDiscountCoupon() {
        Long userId = 5L;
        UserCoupon discount = issueService.issueCoupon(userId, discountTemplate.getId());

        UseCouponRequest request = new UseCouponRequest();
        request.setUserId(userId);
        request.setUserCouponIds(Arrays.asList(discount.getId()));
        request.setOrderAmount(new BigDecimal("100"));
        request.setShippingFee(new BigDecimal("10"));
        request.setOrderId("ORDER_003");

        CouponCalculationResult result = usageService.useCoupons(request);

        assertEquals(0, new BigDecimal("90").compareTo(result.getFinalAmount()));
        assertEquals(0, new BigDecimal("20.00").compareTo(result.getProductDiscount()));
    }

    @Test
    void testMutualExclusionFullReductionAndDiscount() {
        Long userId = 6L;
        UserCoupon fullReduction = issueService.issueCoupon(userId, fullReductionTemplate.getId());
        UserCoupon discount = issueService.issueCoupon(userId, discountTemplate.getId());

        UseCouponRequest request = new UseCouponRequest();
        request.setUserId(userId);
        request.setUserCouponIds(Arrays.asList(fullReduction.getId(), discount.getId()));
        request.setOrderAmount(new BigDecimal("150"));
        request.setShippingFee(new BigDecimal("10"));
        request.setOrderId("ORDER_004");

        CouponException exception = assertThrows(CouponException.class,
                () -> usageService.validateAndCalculate(request));
        assertEquals("满减券和折扣券不能同时使用", exception.getMessage());
    }

    @Test
    void testReturnCoupon() {
        Long userId = 7L;
        UserCoupon userCoupon = issueService.issueCoupon(userId, fullReductionTemplate.getId());
        String orderId = "ORDER_005";

        UseCouponRequest request = new UseCouponRequest();
        request.setUserId(userId);
        request.setUserCouponIds(Arrays.asList(userCoupon.getId()));
        request.setOrderAmount(new BigDecimal("150"));
        request.setShippingFee(new BigDecimal("10"));
        request.setOrderId(orderId);

        usageService.useCoupons(request);

        UserCoupon usedCoupon = userCouponRepository.findById(userCoupon.getId()).orElseThrow();
        assertEquals(CouponStatus.USED, usedCoupon.getStatus());

        usageService.returnCoupons(orderId);

        UserCoupon returnedCoupon = userCouponRepository.findById(userCoupon.getId()).orElseThrow();
        assertEquals(CouponStatus.AVAILABLE, returnedCoupon.getStatus());
        assertNull(returnedCoupon.getOrderId());
    }

    @Test
    void testQueryAvailableCoupons() {
        Long userId = 8L;
        issueService.issueCoupon(userId, fullReductionTemplate.getId());
        issueService.issueCoupon(userId, discountTemplate.getId());

        List<UserCoupon> available = queryService.getAvailableCoupons(userId);
        assertEquals(2, available.size());
        assertEquals(CouponStatus.AVAILABLE, available.get(0).getStatus());
    }

    @Test
    void testTemplateStatistics() {
        Long userId = 9L;
        issueService.issueCoupon(userId, fullReductionTemplate.getId());
        issueService.issueCoupon(userId, fullReductionTemplate.getId());

        Map<String, Object> stats = templateService.getTemplateStatistics(fullReductionTemplate.getId());
        assertEquals(100, stats.get("totalQuantity"));
        assertEquals(2, stats.get("issuedQuantity"));
        assertEquals(0, stats.get("usedQuantity"));
        assertEquals(98, stats.get("remainingQuantity"));
    }

    @Test
    void testThresholdNotMet() {
        Long userId = 10L;
        UserCoupon userCoupon = issueService.issueCoupon(userId, fullReductionTemplate.getId());

        UseCouponRequest request = new UseCouponRequest();
        request.setUserId(userId);
        request.setUserCouponIds(Arrays.asList(userCoupon.getId()));
        request.setOrderAmount(new BigDecimal("50"));
        request.setShippingFee(new BigDecimal("10"));
        request.setOrderId("ORDER_006");

        CouponException exception = assertThrows(CouponException.class,
                () -> usageService.validateAndCalculate(request));
        assertEquals("订单金额不满足优惠券使用门槛", exception.getMessage());
    }
}

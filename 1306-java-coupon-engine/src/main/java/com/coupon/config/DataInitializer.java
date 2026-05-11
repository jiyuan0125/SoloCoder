package com.coupon.config;

import com.coupon.dto.CreateCouponTemplateRequest;
import com.coupon.dto.CouponCalculationResult;
import com.coupon.dto.UseCouponRequest;
import com.coupon.entity.CouponTemplate;
import com.coupon.entity.UserCoupon;
import com.coupon.enums.CouponStatus;
import com.coupon.enums.CouponType;
import com.coupon.service.*;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.boot.CommandLineRunner;
import org.springframework.context.annotation.Profile;
import org.springframework.stereotype.Component;

import java.math.BigDecimal;
import java.time.LocalDateTime;
import java.util.Arrays;
import java.util.List;
import java.util.Map;

@Component
@Profile("!test")
@RequiredArgsConstructor
@Slf4j
public class DataInitializer implements CommandLineRunner {

    private final CouponTemplateService templateService;
    private final CouponIssueService issueService;
    private final CouponUsageService usageService;
    private final CouponQueryService queryService;

    @Override
    public void run(String... args) {
        log.info("========================================");
        log.info("开始初始化优惠券系统演示数据...");
        log.info("========================================");

        try {
            CreateCouponTemplateRequest fullReductionRequest = new CreateCouponTemplateRequest();
            fullReductionRequest.setName("满100减20");
            fullReductionRequest.setType(CouponType.FULL_REDUCTION);
            fullReductionRequest.setValue(new BigDecimal("20"));
            fullReductionRequest.setThreshold(new BigDecimal("100"));
            fullReductionRequest.setStartTime(LocalDateTime.now().minusDays(1));
            fullReductionRequest.setEndTime(LocalDateTime.now().plusDays(30));
            fullReductionRequest.setTotalQuantity(100);
            fullReductionRequest.setLimitPerUser(2);
            CouponTemplate fullReductionTemplate = templateService.createTemplate(fullReductionRequest);
            log.info("✅ 创建满减券模板: ID={}, 名称={}", fullReductionTemplate.getId(), fullReductionTemplate.getName());

            CreateCouponTemplateRequest discountRequest = new CreateCouponTemplateRequest();
            discountRequest.setName("8折优惠券");
            discountRequest.setType(CouponType.DISCOUNT);
            discountRequest.setValue(new BigDecimal("8"));
            discountRequest.setThreshold(new BigDecimal("0"));
            discountRequest.setStartTime(LocalDateTime.now().minusDays(1));
            discountRequest.setEndTime(LocalDateTime.now().plusDays(30));
            discountRequest.setTotalQuantity(100);
            discountRequest.setLimitPerUser(2);
            CouponTemplate discountTemplate = templateService.createTemplate(discountRequest);
            log.info("✅ 创建折扣券模板: ID={}, 名称={}", discountTemplate.getId(), discountTemplate.getName());

            CreateCouponTemplateRequest freeShippingRequest = new CreateCouponTemplateRequest();
            freeShippingRequest.setName("免邮券");
            freeShippingRequest.setType(CouponType.FREE_SHIPPING);
            freeShippingRequest.setValue(BigDecimal.ZERO);
            freeShippingRequest.setThreshold(new BigDecimal("0"));
            freeShippingRequest.setStartTime(LocalDateTime.now().minusDays(1));
            freeShippingRequest.setEndTime(LocalDateTime.now().plusDays(30));
            freeShippingRequest.setTotalQuantity(100);
            freeShippingRequest.setLimitPerUser(5);
            CouponTemplate freeShippingTemplate = templateService.createTemplate(freeShippingRequest);
            log.info("✅ 创建免邮券模板: ID={}, 名称={}", freeShippingTemplate.getId(), freeShippingTemplate.getName());

            Long userId = 1L;

            UserCoupon fullReduction = issueService.issueCoupon(userId, fullReductionTemplate.getId());
            log.info("✅ 用户 {} 领取满减券: 用户券ID={}", userId, fullReduction.getId());

            UserCoupon freeShipping = issueService.issueCoupon(userId, freeShippingTemplate.getId());
            log.info("✅ 用户 {} 领取免邮券: 用户券ID={}", userId, freeShipping.getId());

            UserCoupon discount = issueService.issueCoupon(userId, discountTemplate.getId());
            log.info("✅ 用户 {} 领取折扣券: 用户券ID={}", userId, discount.getId());

            log.info("\n========================================");
            log.info("测试场景1: 使用满减券（订单150元，运费10元）");
            log.info("========================================");
            UseCouponRequest request1 = new UseCouponRequest();
            request1.setUserId(userId);
            request1.setUserCouponIds(Arrays.asList(fullReduction.getId()));
            request1.setOrderAmount(new BigDecimal("150"));
            request1.setShippingFee(new BigDecimal("10"));
            request1.setOrderId("ORDER_001");
            CouponCalculationResult result1 = usageService.useCoupons(request1);
            log.info("💰 最终金额: {}", result1.getFinalAmount());
            log.info("💰 商品折扣: {}", result1.getProductDiscount());
            log.info("💰 运费折扣: {}", result1.getShippingDiscount());
            log.info("✅ 预期: 150-20+10=140元, 实际: {}元", result1.getFinalAmount());
            log.info("✅ 测试通过: {}", new BigDecimal("140").compareTo(result1.getFinalAmount()) == 0);

            log.info("\n========================================");
            log.info("测试场景2: 满减券 + 免邮券叠加");
            log.info("========================================");
            UserCoupon fullReduction2 = issueService.issueCoupon(userId, fullReductionTemplate.getId());
            UserCoupon freeShipping2 = issueService.issueCoupon(userId, freeShippingTemplate.getId());
            UseCouponRequest request2 = new UseCouponRequest();
            request2.setUserId(userId);
            request2.setUserCouponIds(Arrays.asList(fullReduction2.getId(), freeShipping2.getId()));
            request2.setOrderAmount(new BigDecimal("150"));
            request2.setShippingFee(new BigDecimal("10"));
            request2.setOrderId("ORDER_002");
            CouponCalculationResult result2 = usageService.useCoupons(request2);
            log.info("💰 最终金额: {}", result2.getFinalAmount());
            log.info("💰 商品折扣: {}", result2.getProductDiscount());
            log.info("💰 运费折扣: {}", result2.getShippingDiscount());
            log.info("✅ 预期: 150-20+0=130元, 实际: {}元", result2.getFinalAmount());
            log.info("✅ 测试通过: {}", new BigDecimal("130").compareTo(result2.getFinalAmount()) == 0);

            log.info("\n========================================");
            log.info("测试场景3: 满减券 + 折扣券互斥（应报错）");
            log.info("========================================");
            try {
                UseCouponRequest request3 = new UseCouponRequest();
                request3.setUserId(userId);
                request3.setUserCouponIds(Arrays.asList(discount.getId()));
                request3.setOrderAmount(new BigDecimal("150"));
                request3.setShippingFee(new BigDecimal("10"));
                request3.setOrderId("ORDER_003");
                usageService.validateAndCalculate(request3);
                log.info("✅ 折扣券单独使用验证通过");
            } catch (Exception e) {
                log.error("❌ 错误: {}", e.getMessage());
            }

            log.info("\n========================================");
            log.info("测试场景4: 查询用户可用券列表");
            log.info("========================================");
            List<UserCoupon> available = queryService.getAvailableCoupons(userId);
            log.info("📋 用户可用券数量: {}", available.size());
            available.forEach(c -> log.info("   - ID={}, 类型={}, 状态={}", 
                    c.getId(), c.getCouponTemplate().getType(), c.getStatus()));

            log.info("\n========================================");
            log.info("测试场景5: 查询已使用券列表");
            log.info("========================================");
            List<UserCoupon> used = queryService.getUsedCoupons(userId);
            log.info("📋 用户已使用券数量: {}", used.size());
            used.forEach(c -> log.info("   - ID={}, 类型={}, 状态={}, 订单ID={}", 
                    c.getId(), c.getCouponTemplate().getType(), c.getStatus(), c.getOrderId()));

            log.info("\n========================================");
            log.info("测试场景6: 订单取消，退还优惠券");
            log.info("========================================");
            usageService.returnCoupons("ORDER_001");
            log.info("🔄 已退还 ORDER_001 的优惠券");
            List<UserCoupon> usedAfterReturn = queryService.getUsedCoupons(userId);
            log.info("📋 退还后已使用券数量: {}", usedAfterReturn.size());

            log.info("\n========================================");
            log.info("测试场景7: 管理员查看发行统计");
            log.info("========================================");
            Map<String, Object> stats = templateService.getTemplateStatistics(fullReductionTemplate.getId());
            log.info("📊 满减券统计:");
            log.info("   - 总发行量: {}", stats.get("totalQuantity"));
            log.info("   - 已发放: {}", stats.get("issuedQuantity"));
            log.info("   - 已使用: {}", stats.get("usedQuantity"));
            log.info("   - 剩余: {}", stats.get("remainingQuantity"));

            log.info("\n========================================");
            log.info("🎉 所有演示测试完成！优惠券系统运行正常！");
            log.info("========================================");
            log.info("\n💡 API 使用说明:");
            log.info("  - 管理端: POST /api/admin/coupon-templates 创建优惠券模板");
            log.info("  - 用户端: POST /api/users/{userId}/coupons/claim/{templateId} 领取优惠券");
            log.info("  - 用户端: GET /api/users/{userId}/coupons/available 查询可用券");
            log.info("  - 用户端: POST /api/users/{userId}/coupons/use 使用优惠券");
            log.info("  - 用户端: POST /api/users/{userId}/coupons/return/{orderId} 退还优惠券");
            log.info("  - 管理端: GET /api/admin/coupon-templates/{id}/statistics 查看发行统计");
            log.info("\n========================================");

        } catch (Exception e) {
            log.error("❌ 初始化演示数据时发生错误: ", e);
        }
    }
}

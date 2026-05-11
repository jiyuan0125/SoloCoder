package com.coupon.controller;

import com.coupon.dto.CouponCalculationResult;
import com.coupon.dto.UseCouponRequest;
import com.coupon.entity.UserCoupon;
import com.coupon.service.CouponIssueService;
import com.coupon.service.CouponQueryService;
import com.coupon.service.CouponUsageService;
import jakarta.validation.Valid;
import lombok.RequiredArgsConstructor;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import java.util.List;

@RestController
@RequestMapping("/api/users/{userId}/coupons")
@RequiredArgsConstructor
public class UserCouponController {
    private final CouponIssueService couponIssueService;
    private final CouponQueryService couponQueryService;
    private final CouponUsageService couponUsageService;

    @PostMapping("/claim/{templateId}")
    public ResponseEntity<UserCoupon> claimCoupon(@PathVariable Long userId, @PathVariable Long templateId) {
        UserCoupon userCoupon = couponIssueService.issueCoupon(userId, templateId);
        return ResponseEntity.ok(userCoupon);
    }

    @GetMapping("/available")
    public ResponseEntity<List<UserCoupon>> getAvailableCoupons(@PathVariable Long userId) {
        return ResponseEntity.ok(couponQueryService.getAvailableCoupons(userId));
    }

    @GetMapping("/used")
    public ResponseEntity<List<UserCoupon>> getUsedCoupons(@PathVariable Long userId) {
        return ResponseEntity.ok(couponQueryService.getUsedCoupons(userId));
    }

    @GetMapping("/expired")
    public ResponseEntity<List<UserCoupon>> getExpiredCoupons(@PathVariable Long userId) {
        return ResponseEntity.ok(couponQueryService.getExpiredCoupons(userId));
    }

    @PostMapping("/validate")
    public ResponseEntity<CouponCalculationResult> validateCoupons(@PathVariable Long userId,
                                                                    @Valid @RequestBody UseCouponRequest request) {
        request.setUserId(userId);
        CouponCalculationResult result = couponUsageService.validateAndCalculate(request);
        return ResponseEntity.ok(result);
    }

    @PostMapping("/use")
    public ResponseEntity<CouponCalculationResult> useCoupons(@PathVariable Long userId,
                                                               @Valid @RequestBody UseCouponRequest request) {
        request.setUserId(userId);
        CouponCalculationResult result = couponUsageService.useCoupons(request);
        return ResponseEntity.ok(result);
    }

    @PostMapping("/return/{orderId}")
    public ResponseEntity<Void> returnCoupons(@PathVariable Long userId, @PathVariable String orderId) {
        couponUsageService.returnCoupons(orderId);
        return ResponseEntity.ok().build();
    }
}

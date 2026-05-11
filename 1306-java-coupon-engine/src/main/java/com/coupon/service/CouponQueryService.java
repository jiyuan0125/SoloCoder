package com.coupon.service;

import com.coupon.entity.UserCoupon;
import com.coupon.enums.CouponStatus;
import com.coupon.repository.UserCouponRepository;
import lombok.RequiredArgsConstructor;
import org.springframework.stereotype.Service;

import java.time.LocalDateTime;
import java.util.List;

@Service
@RequiredArgsConstructor
public class CouponQueryService {
    private final UserCouponRepository userCouponRepository;

    public List<UserCoupon> getAvailableCoupons(Long userId) {
        LocalDateTime now = LocalDateTime.now();
        List<UserCoupon> availableCoupons = userCouponRepository
                .findByUserIdAndStatusOrderByCouponTemplate_EndTimeAsc(userId, CouponStatus.AVAILABLE);
        return availableCoupons;
    }

    public List<UserCoupon> getUsedCoupons(Long userId) {
        return userCouponRepository.findByUserIdAndStatus(userId, CouponStatus.USED);
    }

    public List<UserCoupon> getExpiredCoupons(Long userId) {
        LocalDateTime now = LocalDateTime.now();
        return userCouponRepository.findExpiredByUserId(userId, CouponStatus.AVAILABLE, now);
    }
}

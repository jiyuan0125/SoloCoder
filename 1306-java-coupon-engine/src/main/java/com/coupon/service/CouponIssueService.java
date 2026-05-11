package com.coupon.service;

import com.coupon.entity.CouponTemplate;
import com.coupon.entity.UserCoupon;
import com.coupon.enums.CouponStatus;
import com.coupon.exception.CouponException;
import com.coupon.repository.CouponTemplateRepository;
import com.coupon.repository.UserCouponRepository;
import lombok.RequiredArgsConstructor;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.time.LocalDateTime;

@Service
@RequiredArgsConstructor
public class CouponIssueService {
    private final CouponTemplateRepository couponTemplateRepository;
    private final UserCouponRepository userCouponRepository;

    @Transactional
    public UserCoupon issueCoupon(Long userId, Long templateId) {
        CouponTemplate template = couponTemplateRepository.findById(templateId)
                .orElseThrow(() -> new CouponException("优惠券模板不存在"));

        LocalDateTime now = LocalDateTime.now();
        if (now.isBefore(template.getStartTime()) || now.isAfter(template.getEndTime())) {
            throw new CouponException("优惠券不在领取时间范围内");
        }

        if (template.getIssuedQuantity() >= template.getTotalQuantity()) {
            throw new CouponException("优惠券已领完");
        }

        long userReceivedCount = userCouponRepository.countByUserIdAndCouponTemplateId(userId, templateId);
        if (userReceivedCount >= template.getLimitPerUser()) {
            throw new CouponException("您已达到该优惠券的领取上限");
        }

        template.setIssuedQuantity(template.getIssuedQuantity() + 1);
        couponTemplateRepository.save(template);

        UserCoupon userCoupon = new UserCoupon();
        userCoupon.setUserId(userId);
        userCoupon.setCouponTemplate(template);
        userCoupon.setStatus(CouponStatus.AVAILABLE);

        return userCouponRepository.save(userCoupon);
    }
}

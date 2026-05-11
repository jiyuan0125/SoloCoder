package com.coupon.service;

import com.coupon.dto.CouponCalculationResult;
import com.coupon.dto.UseCouponRequest;
import com.coupon.entity.CouponTemplate;
import com.coupon.entity.UserCoupon;
import com.coupon.enums.CouponStatus;
import com.coupon.enums.CouponType;
import com.coupon.exception.CouponException;
import com.coupon.repository.CouponTemplateRepository;
import com.coupon.repository.UserCouponRepository;
import lombok.RequiredArgsConstructor;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.math.BigDecimal;
import java.time.LocalDateTime;
import java.util.ArrayList;
import java.util.List;

@Service
@RequiredArgsConstructor
public class CouponUsageService {
    private final UserCouponRepository userCouponRepository;
    private final CouponTemplateRepository couponTemplateRepository;

    public CouponCalculationResult validateAndCalculate(UseCouponRequest request) {
        if (request.getUserCouponIds() == null || request.getUserCouponIds().isEmpty()) {
            CouponCalculationResult result = new CouponCalculationResult();
            result.setFinalAmount(request.getOrderAmount().add(request.getShippingFee()));
            result.setProductDiscount(BigDecimal.ZERO);
            result.setShippingDiscount(BigDecimal.ZERO);
            result.setAppliedCouponIds(new ArrayList<>());
            return result;
        }

        List<UserCoupon> userCoupons = new ArrayList<>();
        for (Long userCouponId : request.getUserCouponIds()) {
            UserCoupon uc = userCouponRepository.findByIdAndUserId(userCouponId, request.getUserId())
                    .orElseThrow(() -> new CouponException("优惠券不存在或不属于当前用户"));
            userCoupons.add(uc);
        }

        validateMutualExclusion(userCoupons);

        for (UserCoupon userCoupon : userCoupons) {
            validateCouponBasic(userCoupon, request.getOrderAmount());
        }

        return calculateDiscount(userCoupons, request.getOrderAmount(), request.getShippingFee());
    }

    @Transactional
    public CouponCalculationResult useCoupons(UseCouponRequest request) {
        CouponCalculationResult result = validateAndCalculate(request);

        if (request.getUserCouponIds() != null && !request.getUserCouponIds().isEmpty()) {
            LocalDateTime now = LocalDateTime.now();
            for (Long userCouponId : request.getUserCouponIds()) {
                UserCoupon userCoupon = userCouponRepository.findById(userCouponId)
                        .orElseThrow(() -> new CouponException("优惠券不存在"));
                userCoupon.setStatus(CouponStatus.USED);
                userCoupon.setOrderId(request.getOrderId());
                userCoupon.setUsedTime(now);
                userCouponRepository.save(userCoupon);

                CouponTemplate template = userCoupon.getCouponTemplate();
                template.setUsedQuantity(template.getUsedQuantity() + 1);
                couponTemplateRepository.save(template);
            }
        }

        return result;
    }

    @Transactional
    public void returnCoupons(String orderId) {
        List<UserCoupon> userCoupons = userCouponRepository.findByOrderId(orderId);
        if (userCoupons.isEmpty()) {
            return;
        }

        for (UserCoupon userCoupon : userCoupons) {
            if (userCoupon.getStatus() == CouponStatus.USED) {
                userCoupon.setStatus(CouponStatus.AVAILABLE);
                userCoupon.setOrderId(null);
                userCoupon.setUsedTime(null);
                userCouponRepository.save(userCoupon);

                CouponTemplate template = userCoupon.getCouponTemplate();
                if (template.getUsedQuantity() > 0) {
                    template.setUsedQuantity(template.getUsedQuantity() - 1);
                    couponTemplateRepository.save(template);
                }
            }
        }
    }

    private void validateMutualExclusion(List<UserCoupon> userCoupons) {
        boolean hasFullReduction = false;
        boolean hasDiscount = false;

        for (UserCoupon uc : userCoupons) {
            CouponType type = uc.getCouponTemplate().getType();
            if (type == CouponType.FULL_REDUCTION) {
                hasFullReduction = true;
            } else if (type == CouponType.DISCOUNT) {
                hasDiscount = true;
            }
        }

        if (hasFullReduction && hasDiscount) {
            throw new CouponException("满减券和折扣券不能同时使用");
        }
    }

    private void validateCouponBasic(UserCoupon userCoupon, BigDecimal orderAmount) {
        if (userCoupon.getStatus() != CouponStatus.AVAILABLE) {
            throw new CouponException("优惠券状态不可用");
        }

        CouponTemplate template = userCoupon.getCouponTemplate();
        LocalDateTime now = LocalDateTime.now();

        if (now.isBefore(template.getStartTime()) || now.isAfter(template.getEndTime())) {
            throw new CouponException("优惠券不在有效期内");
        }

        if (orderAmount.compareTo(template.getThreshold()) < 0) {
            throw new CouponException("订单金额不满足优惠券使用门槛");
        }
    }

    private CouponCalculationResult calculateDiscount(List<UserCoupon> userCoupons,
                                                      BigDecimal orderAmount,
                                                      BigDecimal shippingFee) {
        BigDecimal productDiscount = BigDecimal.ZERO;
        BigDecimal shippingDiscount = BigDecimal.ZERO;
        List<Long> appliedIds = new ArrayList<>();

        BigDecimal currentProductPrice = orderAmount;

        for (UserCoupon userCoupon : userCoupons) {
            CouponTemplate template = userCoupon.getCouponTemplate();
            CouponType type = template.getType();

            if (type == CouponType.FULL_REDUCTION) {
                if (currentProductPrice.compareTo(template.getThreshold()) >= 0) {
                    BigDecimal reduction = template.getValue();
                    if (reduction.compareTo(currentProductPrice) > 0) {
                        reduction = currentProductPrice;
                    }
                    productDiscount = productDiscount.add(reduction);
                    currentProductPrice = currentProductPrice.subtract(reduction);
                    appliedIds.add(userCoupon.getId());
                }
            } else if (type == CouponType.DISCOUNT) {
                BigDecimal discountRate = template.getValue().divide(BigDecimal.TEN, 2, java.math.RoundingMode.HALF_UP);
                BigDecimal discountAmount = currentProductPrice.multiply(BigDecimal.ONE.subtract(discountRate));
                productDiscount = productDiscount.add(discountAmount);
                currentProductPrice = currentProductPrice.multiply(discountRate);
                appliedIds.add(userCoupon.getId());
            } else if (type == CouponType.FREE_SHIPPING) {
                shippingDiscount = shippingFee;
                appliedIds.add(userCoupon.getId());
            }
        }

        BigDecimal finalProductPrice = currentProductPrice;
        BigDecimal finalShippingFee = shippingFee.subtract(shippingDiscount);
        if (finalShippingFee.compareTo(BigDecimal.ZERO) < 0) {
            finalShippingFee = BigDecimal.ZERO;
        }

        CouponCalculationResult result = new CouponCalculationResult();
        result.setFinalAmount(finalProductPrice.add(finalShippingFee));
        result.setProductDiscount(productDiscount);
        result.setShippingDiscount(shippingDiscount);
        result.setAppliedCouponIds(appliedIds);

        return result;
    }
}

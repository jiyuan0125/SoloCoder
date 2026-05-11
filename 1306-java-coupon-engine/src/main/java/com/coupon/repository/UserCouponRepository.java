package com.coupon.repository;

import com.coupon.entity.UserCoupon;
import com.coupon.enums.CouponStatus;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.data.jpa.repository.Query;
import org.springframework.data.repository.query.Param;

import java.time.LocalDateTime;
import java.util.List;
import java.util.Optional;

public interface UserCouponRepository extends JpaRepository<UserCoupon, Long> {
    long countByUserIdAndCouponTemplateId(@Param("userId") Long userId, @Param("couponTemplateId") Long couponTemplateId);

    List<UserCoupon> findByUserIdAndStatusOrderByCouponTemplate_EndTimeAsc(@Param("userId") Long userId, @Param("status") CouponStatus status);

    List<UserCoupon> findByUserIdAndStatus(@Param("userId") Long userId, @Param("status") CouponStatus status);

    @Query("SELECT uc FROM UserCoupon uc WHERE uc.userId = :userId AND uc.status = :status AND uc.couponTemplate.endTime < :now ORDER BY uc.couponTemplate.endTime DESC")
    List<UserCoupon> findExpiredByUserId(@Param("userId") Long userId, @Param("status") CouponStatus status, @Param("now") LocalDateTime now);

    Optional<UserCoupon> findByIdAndUserId(@Param("id") Long id, @Param("userId") Long userId);

    List<UserCoupon> findByOrderId(@Param("orderId") String orderId);
}

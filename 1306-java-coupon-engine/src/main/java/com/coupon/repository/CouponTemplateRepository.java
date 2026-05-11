package com.coupon.repository;

import com.coupon.entity.CouponTemplate;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.data.jpa.repository.Query;
import org.springframework.data.repository.query.Param;

import java.time.LocalDateTime;
import java.util.List;

public interface CouponTemplateRepository extends JpaRepository<CouponTemplate, Long> {
    @Query("SELECT ct FROM CouponTemplate ct WHERE ct.startTime <= :now AND ct.endTime >= :now AND ct.issuedQuantity < ct.totalQuantity")
    List<CouponTemplate> findAvailableTemplates(@Param("now") LocalDateTime now);
}

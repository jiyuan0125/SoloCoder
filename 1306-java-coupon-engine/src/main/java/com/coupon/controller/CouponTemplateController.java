package com.coupon.controller;

import com.coupon.dto.CreateCouponTemplateRequest;
import com.coupon.entity.CouponTemplate;
import com.coupon.service.CouponTemplateService;
import jakarta.validation.Valid;
import lombok.RequiredArgsConstructor;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import java.util.List;
import java.util.Map;

@RestController
@RequestMapping("/api/admin/coupon-templates")
@RequiredArgsConstructor
public class CouponTemplateController {
    private final CouponTemplateService couponTemplateService;

    @PostMapping
    public ResponseEntity<CouponTemplate> createTemplate(@Valid @RequestBody CreateCouponTemplateRequest request) {
        CouponTemplate template = couponTemplateService.createTemplate(request);
        return ResponseEntity.ok(template);
    }

    @GetMapping
    public ResponseEntity<List<CouponTemplate>> getAllTemplates() {
        return ResponseEntity.ok(couponTemplateService.getAllTemplates());
    }

    @GetMapping("/{id}")
    public ResponseEntity<CouponTemplate> getTemplateById(@PathVariable Long id) {
        return ResponseEntity.ok(couponTemplateService.getTemplateById(id));
    }

    @GetMapping("/{id}/statistics")
    public ResponseEntity<Map<String, Object>> getTemplateStatistics(@PathVariable Long id) {
        return ResponseEntity.ok(couponTemplateService.getTemplateStatistics(id));
    }
}

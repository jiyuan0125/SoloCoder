package com.coupon.service;

import com.coupon.dto.CreateCouponTemplateRequest;
import com.coupon.entity.CouponTemplate;
import com.coupon.exception.CouponException;
import com.coupon.repository.CouponTemplateRepository;
import lombok.RequiredArgsConstructor;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.util.List;
import java.util.Map;

@Service
@RequiredArgsConstructor
public class CouponTemplateService {
    private final CouponTemplateRepository couponTemplateRepository;

    @Transactional
    public CouponTemplate createTemplate(CreateCouponTemplateRequest request) {
        if (request.getStartTime().isAfter(request.getEndTime())) {
            throw new CouponException("开始时间不能晚于结束时间");
        }
        if (request.getTotalQuantity() <= 0) {
            throw new CouponException("总发行量必须大于0");
        }
        if (request.getLimitPerUser() <= 0) {
            throw new CouponException("用户限领数量必须大于0");
        }

        CouponTemplate template = new CouponTemplate();
        template.setName(request.getName());
        template.setType(request.getType());
        template.setValue(request.getValue());
        template.setThreshold(request.getThreshold());
        template.setStartTime(request.getStartTime());
        template.setEndTime(request.getEndTime());
        template.setTotalQuantity(request.getTotalQuantity());
        template.setLimitPerUser(request.getLimitPerUser());
        template.setIssuedQuantity(0);
        template.setUsedQuantity(0);

        return couponTemplateRepository.save(template);
    }

    public List<CouponTemplate> getAllTemplates() {
        return couponTemplateRepository.findAll();
    }

    public CouponTemplate getTemplateById(Long id) {
        return couponTemplateRepository.findById(id)
                .orElseThrow(() -> new CouponException("优惠券模板不存在"));
    }

    public Map<String, Object> getTemplateStatistics(Long templateId) {
        CouponTemplate template = getTemplateById(templateId);
        int remaining = template.getTotalQuantity() - template.getIssuedQuantity();
        return Map.of(
                "templateId", template.getId(),
                "name", template.getName(),
                "totalQuantity", template.getTotalQuantity(),
                "issuedQuantity", template.getIssuedQuantity(),
                "usedQuantity", template.getUsedQuantity(),
                "remainingQuantity", remaining
        );
    }
}

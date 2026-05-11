package com.safety.inspection.service;

import com.baomidou.mybatisplus.core.conditions.query.LambdaQueryWrapper;
import com.baomidou.mybatisplus.extension.service.impl.ServiceImpl;
import com.safety.inspection.entity.Hazard;
import com.safety.inspection.mapper.HazardMapper;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.stereotype.Service;

import java.time.LocalDateTime;
import java.util.List;

@Slf4j
@Service
@RequiredArgsConstructor
public class RepeatHazardService extends ServiceImpl<HazardMapper, Hazard> {

    private final HazardMapper hazardMapper;

    public Hazard checkAndMarkRepeatHazard(Hazard hazard) {
        if (hazard.getLocation() == null || hazard.getHazardType() == null) {
            return hazard;
        }

        List<Hazard> similarHazards = hazardMapper.selectList(
            new LambdaQueryWrapper<Hazard>()
                .eq(Hazard::getAreaId, hazard.getAreaId())
                .eq(Hazard::getHazardType, hazard.getHazardType())
                .eq(Hazard::getLocation, hazard.getLocation())
                .ne(Hazard::getId, hazard.getId())
                .orderByDesc(Hazard::getCreatedAt)
                .last("LIMIT 5")
        );

        if (!similarHazards.isEmpty()) {
            hazard.setIsRepeat(1);
            hazard.setRelatedHazardId(similarHazards.get(0).getId());
            log.info("检测到重复隐患：areaId={}, type={}, location={}", 
                hazard.getAreaId(), hazard.getHazardType(), hazard.getLocation());
        }

        return hazard;
    }

    public List<Hazard> getRelatedHazards(Long hazardId) {
        Hazard hazard = hazardMapper.selectById(hazardId);
        if (hazard == null || hazard.getIsRepeat() == null || hazard.getIsRepeat() == 0) {
            return List.of();
        }

        return hazardMapper.selectList(
            new LambdaQueryWrapper<Hazard>()
                .eq(Hazard::getAreaId, hazard.getAreaId())
                .eq(Hazard::getHazardType, hazard.getHazardType())
                .eq(Hazard::getLocation, hazard.getLocation())
                .orderByDesc(Hazard::getCreatedAt)
        );
    }

    public List<Hazard> getAllRepeatHazards() {
        return hazardMapper.selectList(
            new LambdaQueryWrapper<Hazard>()
                .eq(Hazard::getIsRepeat, 1)
                .ne(Hazard::getStatus, "CLOSED")
                .orderByDesc(Hazard::getCreatedAt)
        );
    }
}

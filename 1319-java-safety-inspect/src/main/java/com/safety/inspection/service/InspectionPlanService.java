package com.safety.inspection.service;

import com.baomidou.mybatisplus.core.conditions.query.LambdaQueryWrapper;
import com.baomidou.mybatisplus.extension.plugins.pagination.Page;
import com.baomidou.mybatisplus.extension.service.impl.ServiceImpl;
import com.safety.inspection.common.BusinessException;
import com.safety.inspection.dto.InspectionPlanDTO;
import com.safety.inspection.entity.InspectionPlan;
import com.safety.inspection.enums.FrequencyEnum;
import com.safety.inspection.mapper.InspectionPlanMapper;
import com.safety.inspection.security.CurrentUser;
import lombok.RequiredArgsConstructor;
import org.springframework.stereotype.Service;
import org.springframework.util.StringUtils;

import java.time.LocalDate;
import java.util.List;

@Service
@RequiredArgsConstructor
public class InspectionPlanService extends ServiceImpl<InspectionPlanMapper, InspectionPlan> {

    private final InspectionPlanMapper planMapper;
    private final CurrentUser currentUser;

    public Page<InspectionPlan> getPlanPage(int pageNum, int pageSize, String planName, Long areaId, Long inspectorId, Integer status) {
        Page<InspectionPlan> page = new Page<>(pageNum, pageSize);
        LambdaQueryWrapper<InspectionPlan> wrapper = new LambdaQueryWrapper<>();

        if (StringUtils.hasText(planName)) {
            wrapper.like(InspectionPlan::getPlanName, planName);
        }
        if (areaId != null) {
            wrapper.eq(InspectionPlan::getAreaId, areaId);
        }
        if (inspectorId != null) {
            wrapper.eq(InspectionPlan::getInspectorId, inspectorId);
        }
        if (status != null) {
            wrapper.eq(InspectionPlan::getStatus, status);
        }
        wrapper.orderByDesc(InspectionPlan::getCreatedAt);

        return planMapper.selectPage(page, wrapper);
    }

    public List<InspectionPlan> getActivePlans() {
        LocalDate today = LocalDate.now();
        LambdaQueryWrapper<InspectionPlan> wrapper = new LambdaQueryWrapper<>();
        wrapper.eq(InspectionPlan::getStatus, 1)
               .le(InspectionPlan::getStartDate, today)
               .and(w -> w.isNull(InspectionPlan::getEndDate).or().ge(InspectionPlan::getEndDate, today))
               .orderByAsc(InspectionPlan::getId);
        return planMapper.selectList(wrapper);
    }

    public InspectionPlan getPlanById(Long id) {
        return planMapper.selectById(id);
    }

    public void createPlan(InspectionPlanDTO dto) {
        validateFrequency(dto);

        InspectionPlan plan = new InspectionPlan();
        plan.setPlanName(dto.getPlanName());
        plan.setPlanType(dto.getPlanType() != null ? dto.getPlanType() : "REGULAR");
        plan.setAreaId(dto.getAreaId());
        plan.setFrequency(dto.getFrequency());
        plan.setFrequencyDays(getFrequencyDays(dto));
        plan.setStartDate(dto.getStartDate());
        plan.setEndDate(dto.getEndDate());
        plan.setInspectorId(dto.getInspectorId());
        plan.setDescription(dto.getDescription());
        plan.setStatus(dto.getStatus() != null ? dto.getStatus() : 1);
        plan.setCreatedBy(currentUser.getCurrentUserId());
        planMapper.insert(plan);
    }

    public void updatePlan(InspectionPlanDTO dto) {
        InspectionPlan plan = planMapper.selectById(dto.getId());
        if (plan == null) {
            throw new BusinessException("巡检计划不存在");
        }
        validateFrequency(dto);

        plan.setPlanName(dto.getPlanName());
        plan.setPlanType(dto.getPlanType());
        plan.setAreaId(dto.getAreaId());
        plan.setFrequency(dto.getFrequency());
        plan.setFrequencyDays(getFrequencyDays(dto));
        plan.setStartDate(dto.getStartDate());
        plan.setEndDate(dto.getEndDate());
        plan.setInspectorId(dto.getInspectorId());
        plan.setDescription(dto.getDescription());
        plan.setStatus(dto.getStatus());
        planMapper.updateById(plan);
    }

    public void deletePlan(Long id) {
        planMapper.deleteById(id);
    }

    public void togglePlanStatus(Long id) {
        InspectionPlan plan = planMapper.selectById(id);
        if (plan == null) {
            throw new BusinessException("巡检计划不存在");
        }
        plan.setStatus(plan.getStatus() == 1 ? 0 : 1);
        planMapper.updateById(plan);
    }

    private void validateFrequency(InspectionPlanDTO dto) {
        if ("CUSTOM".equals(dto.getFrequency()) && dto.getFrequencyDays() == null) {
            throw new BusinessException("自定义频率必须指定天数");
        }
    }

    private Integer getFrequencyDays(InspectionPlanDTO dto) {
        FrequencyEnum frequency = FrequencyEnum.valueOf(dto.getFrequency());
        if (frequency == FrequencyEnum.CUSTOM) {
            return dto.getFrequencyDays();
        }
        return frequency.getDays();
    }
}

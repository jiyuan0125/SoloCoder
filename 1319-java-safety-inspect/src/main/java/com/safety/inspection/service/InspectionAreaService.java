package com.safety.inspection.service;

import com.baomidou.mybatisplus.core.conditions.query.LambdaQueryWrapper;
import com.baomidou.mybatisplus.extension.plugins.pagination.Page;
import com.baomidou.mybatisplus.extension.service.impl.ServiceImpl;
import com.safety.inspection.common.BusinessException;
import com.safety.inspection.dto.InspectionAreaDTO;
import com.safety.inspection.entity.InspectionArea;
import com.safety.inspection.mapper.InspectionAreaMapper;
import lombok.RequiredArgsConstructor;
import org.springframework.stereotype.Service;
import org.springframework.util.StringUtils;

import java.util.List;

@Service
@RequiredArgsConstructor
public class InspectionAreaService extends ServiceImpl<InspectionAreaMapper, InspectionArea> {

    private final InspectionAreaMapper areaMapper;

    public List<InspectionArea> getAllAreas(String areaName, String areaType) {
        LambdaQueryWrapper<InspectionArea> wrapper = new LambdaQueryWrapper<>();
        if (StringUtils.hasText(areaName)) {
            wrapper.like(InspectionArea::getAreaName, areaName);
        }
        if (StringUtils.hasText(areaType)) {
            wrapper.eq(InspectionArea::getAreaType, areaType);
        }
        wrapper.eq(InspectionArea::getStatus, 1)
               .orderByAsc(InspectionArea::getId);
        return areaMapper.selectList(wrapper);
    }

    public Page<InspectionArea> getAreaPage(int pageNum, int pageSize, String areaName, String areaType, Long departmentId) {
        Page<InspectionArea> page = new Page<>(pageNum, pageSize);
        LambdaQueryWrapper<InspectionArea> wrapper = new LambdaQueryWrapper<>();

        if (StringUtils.hasText(areaName)) {
            wrapper.like(InspectionArea::getAreaName, areaName);
        }
        if (StringUtils.hasText(areaType)) {
            wrapper.eq(InspectionArea::getAreaType, areaType);
        }
        if (departmentId != null) {
            wrapper.eq(InspectionArea::getDepartmentId, departmentId);
        }
        wrapper.orderByDesc(InspectionArea::getCreatedAt);

        return areaMapper.selectPage(page, wrapper);
    }

    public InspectionArea getAreaById(Long id) {
        return areaMapper.selectById(id);
    }

    public void createArea(InspectionAreaDTO dto) {
        InspectionArea area = new InspectionArea();
        area.setAreaName(dto.getAreaName());
        area.setAreaType(dto.getAreaType());
        area.setLocation(dto.getLocation());
        area.setDescription(dto.getDescription());
        area.setManagerId(dto.getManagerId());
        area.setDepartmentId(dto.getDepartmentId());
        area.setStatus(dto.getStatus() != null ? dto.getStatus() : 1);
        areaMapper.insert(area);
    }

    public void updateArea(InspectionAreaDTO dto) {
        InspectionArea area = areaMapper.selectById(dto.getId());
        if (area == null) {
            throw new BusinessException("区域不存在");
        }
        area.setAreaName(dto.getAreaName());
        area.setAreaType(dto.getAreaType());
        area.setLocation(dto.getLocation());
        area.setDescription(dto.getDescription());
        area.setManagerId(dto.getManagerId());
        area.setDepartmentId(dto.getDepartmentId());
        area.setStatus(dto.getStatus());
        areaMapper.updateById(area);
    }

    public void deleteArea(Long id) {
        areaMapper.deleteById(id);
    }
}

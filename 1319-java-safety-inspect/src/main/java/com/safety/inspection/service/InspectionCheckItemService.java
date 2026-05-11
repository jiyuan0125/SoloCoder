package com.safety.inspection.service;

import com.baomidou.mybatisplus.core.conditions.query.LambdaQueryWrapper;
import com.baomidou.mybatisplus.extension.service.impl.ServiceImpl;
import com.safety.inspection.common.BusinessException;
import com.safety.inspection.dto.InspectionCheckItemDTO;
import com.safety.inspection.entity.InspectionCheckItem;
import com.safety.inspection.mapper.InspectionCheckItemMapper;
import lombok.RequiredArgsConstructor;
import org.springframework.stereotype.Service;

import java.util.List;

@Service
@RequiredArgsConstructor
public class InspectionCheckItemService extends ServiceImpl<InspectionCheckItemMapper, InspectionCheckItem> {

    private final InspectionCheckItemMapper checkItemMapper;

    public List<InspectionCheckItem> getItemsByArea(Long areaId) {
        LambdaQueryWrapper<InspectionCheckItem> wrapper = new LambdaQueryWrapper<>();
        wrapper.eq(InspectionCheckItem::getAreaId, areaId)
               .eq(InspectionCheckItem::getStatus, 1)
               .orderByAsc(InspectionCheckItem::getSort);
        return checkItemMapper.selectList(wrapper);
    }

    public InspectionCheckItem getItemById(Long id) {
        return checkItemMapper.selectById(id);
    }

    public void createCheckItem(InspectionCheckItemDTO dto) {
        InspectionCheckItem item = new InspectionCheckItem();
        item.setAreaId(dto.getAreaId());
        item.setItemName(dto.getItemName());
        item.setItemDescription(dto.getItemDescription());
        item.setStandard(dto.getStandard());
        item.setRiskLevel(dto.getRiskLevel() != null ? dto.getRiskLevel() : "LOW");
        item.setSort(dto.getSort() != null ? dto.getSort() : 0);
        item.setStatus(dto.getStatus() != null ? dto.getStatus() : 1);
        checkItemMapper.insert(item);
    }

    public void updateCheckItem(InspectionCheckItemDTO dto) {
        InspectionCheckItem item = checkItemMapper.selectById(dto.getId());
        if (item == null) {
            throw new BusinessException("检查项不存在");
        }
        item.setAreaId(dto.getAreaId());
        item.setItemName(dto.getItemName());
        item.setItemDescription(dto.getItemDescription());
        item.setStandard(dto.getStandard());
        item.setRiskLevel(dto.getRiskLevel());
        item.setSort(dto.getSort());
        item.setStatus(dto.getStatus());
        checkItemMapper.updateById(item);
    }

    public void deleteCheckItem(Long id) {
        checkItemMapper.deleteById(id);
    }

    public void batchCreateCheckItems(List<InspectionCheckItemDTO> items) {
        for (InspectionCheckItemDTO dto : items) {
            createCheckItem(dto);
        }
    }
}

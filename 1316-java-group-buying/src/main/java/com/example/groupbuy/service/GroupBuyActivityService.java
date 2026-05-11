package com.example.groupbuy.service;

import com.baomidou.mybatisplus.core.conditions.query.LambdaQueryWrapper;
import com.baomidou.mybatisplus.core.metadata.IPage;
import com.baomidou.mybatisplus.extension.plugins.pagination.Page;
import com.baomidou.mybatisplus.extension.service.impl.ServiceImpl;
import com.example.groupbuy.common.GroupBuyConstants;
import com.example.groupbuy.dto.GroupBuyActivityDTO;
import com.example.groupbuy.dto.PriceTierDTO;
import com.example.groupbuy.entity.GroupBuyActivity;
import com.example.groupbuy.entity.GroupBuyPriceTier;
import com.example.groupbuy.exception.GroupBuyException;
import com.example.groupbuy.mapper.GroupBuyActivityMapper;
import com.example.groupbuy.mapper.GroupBuyPriceTierMapper;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.math.BigDecimal;
import java.time.LocalDateTime;
import java.util.List;
import java.util.stream.Collectors;

@Slf4j
@Service
@RequiredArgsConstructor
public class GroupBuyActivityService extends ServiceImpl<GroupBuyActivityMapper, GroupBuyActivity> {
    
    private final GroupBuyPriceTierMapper priceTierMapper;
    
    @Transactional(rollbackFor = Exception.class)
    public GroupBuyActivity createActivity(GroupBuyActivityDTO dto) {
        validateActivity(dto);
        
        GroupBuyActivity activity = new GroupBuyActivity();
        activity.setProductId(dto.getProductId());
        activity.setProductName(dto.getProductName());
        activity.setOriginalPrice(dto.getOriginalPrice());
        activity.setProductType(dto.getProductType());
        activity.setMaxGroupCount(dto.getMaxGroupCount());
        activity.setUsedGroupCount(0);
        activity.setStartTime(dto.getStartTime());
        activity.setEndTime(dto.getEndTime());
        activity.setStatus(calculateStatus(dto.getStartTime(), dto.getEndTime()));
        activity.setCreateTime(LocalDateTime.now());
        activity.setUpdateTime(LocalDateTime.now());
        
        this.save(activity);
        
        savePriceTiers(activity.getId(), dto.getPriceTiers());
        
        log.info("创建拼团活动成功，活动ID: {}", activity.getId());
        return activity;
    }
    
    @Transactional(rollbackFor = Exception.class)
    public GroupBuyActivity updateActivity(GroupBuyActivityDTO dto) {
        if (dto.getId() == null) {
            throw new GroupBuyException("活动ID不能为空");
        }
        
        GroupBuyActivity activity = this.getById(dto.getId());
        if (activity == null) {
            throw new GroupBuyException("活动不存在");
        }
        
        if (activity.getStatus() != GroupBuyConstants.ACTIVITY_STATUS_NOT_STARTED) {
            throw new GroupBuyException("活动进行中或已结束，无法修改");
        }
        
        validateActivity(dto);
        
        activity.setProductId(dto.getProductId());
        activity.setProductName(dto.getProductName());
        activity.setOriginalPrice(dto.getOriginalPrice());
        activity.setProductType(dto.getProductType());
        activity.setMaxGroupCount(dto.getMaxGroupCount());
        activity.setStartTime(dto.getStartTime());
        activity.setEndTime(dto.getEndTime());
        activity.setStatus(calculateStatus(dto.getStartTime(), dto.getEndTime()));
        activity.setUpdateTime(LocalDateTime.now());
        
        this.updateById(activity);
        
        priceTierMapper.delete(new LambdaQueryWrapper<GroupBuyPriceTier>()
                .eq(GroupBuyPriceTier::getActivityId, activity.getId()));
        
        savePriceTiers(activity.getId(), dto.getPriceTiers());
        
        log.info("更新拼团活动成功，活动ID: {}", activity.getId());
        return activity;
    }
    
    public IPage<GroupBuyActivity> listActivities(int page, int size, Integer status) {
        Page<GroupBuyActivity> pageParam = new Page<>(page, size);
        LambdaQueryWrapper<GroupBuyActivity> wrapper = new LambdaQueryWrapper<>();
        
        if (status != null) {
            wrapper.eq(GroupBuyActivity::getStatus, status);
        }
        
        wrapper.orderByDesc(GroupBuyActivity::getCreateTime);
        return this.page(pageParam, wrapper);
    }
    
    public GroupBuyActivity getActivityDetail(Long activityId) {
        GroupBuyActivity activity = this.getById(activityId);
        if (activity == null) {
            throw new GroupBuyException("活动不存在");
        }
        return activity;
    }
    
    public List<GroupBuyPriceTier> getPriceTiers(Long activityId) {
        return priceTierMapper.selectList(new LambdaQueryWrapper<GroupBuyPriceTier>()
                .eq(GroupBuyPriceTier::getActivityId, activityId)
                .orderByAsc(GroupBuyPriceTier::getSortOrder));
    }
    
    @Transactional(rollbackFor = Exception.class)
    public boolean updateActivityStatus() {
        LocalDateTime now = LocalDateTime.now();
        
        List<GroupBuyActivity> activities = this.list(new LambdaQueryWrapper<GroupBuyActivity>()
                .in(GroupBuyActivity::getStatus, GroupBuyConstants.ACTIVITY_STATUS_NOT_STARTED, GroupBuyConstants.ACTIVITY_STATUS_ONGOING));
        
        int updated = 0;
        for (GroupBuyActivity activity : activities) {
            int newStatus = calculateStatus(activity.getStartTime(), activity.getEndTime());
            if (newStatus != activity.getStatus()) {
                activity.setStatus(newStatus);
                activity.setUpdateTime(now);
                this.updateById(activity);
                updated++;
            }
        }
        
        if (updated > 0) {
            log.info("更新了{}个活动的状态", updated);
        }
        
        return true;
    }
    
    private void validateActivity(GroupBuyActivityDTO dto) {
        if (dto.getEndTime().isBefore(dto.getStartTime())) {
            throw new GroupBuyException("活动结束时间不能早于开始时间");
        }
        
        dto.getPriceTiers().sort((a, b) -> a.getTargetPeopleCount().compareTo(b.getTargetPeopleCount()));
        
        for (int i = 0; i < dto.getPriceTiers().size(); i++) {
            PriceTierDTO tier = dto.getPriceTiers().get(i);
            tier.setSortOrder(i);
            
            if (tier.getTargetPeopleCount() < tier.getMinPeopleCount()) {
                throw new GroupBuyException("目标人数不能小于最低成团人数");
            }
            
            if (tier.getGroupPrice().compareTo(dto.getOriginalPrice()) >= 0) {
                throw new GroupBuyException("团购价必须低于原价");
            }
            
            if (tier.getDiscountRate() == null) {
                tier.setDiscountRate(tier.getGroupPrice().divide(dto.getOriginalPrice(), 4, BigDecimal.ROUND_HALF_UP)
                        .multiply(BigDecimal.valueOf(10)));
            }
            
            if (i > 0) {
                PriceTierDTO prevTier = dto.getPriceTiers().get(i - 1);
                if (tier.getGroupPrice().compareTo(prevTier.getGroupPrice()) >= 0) {
                    throw new GroupBuyException("人数更多的阶梯价格应该更低");
                }
            }
        }
    }
    
    private int calculateStatus(LocalDateTime startTime, LocalDateTime endTime) {
        LocalDateTime now = LocalDateTime.now();
        if (now.isBefore(startTime)) {
            return GroupBuyConstants.ACTIVITY_STATUS_NOT_STARTED;
        } else if (now.isAfter(endTime)) {
            return GroupBuyConstants.ACTIVITY_STATUS_ENDED;
        } else {
            return GroupBuyConstants.ACTIVITY_STATUS_ONGOING;
        }
    }
    
    private void savePriceTiers(Long activityId, List<PriceTierDTO> priceTiers) {
        for (PriceTierDTO tierDTO : priceTiers) {
            GroupBuyPriceTier tier = new GroupBuyPriceTier();
            tier.setActivityId(activityId);
            tier.setMinPeopleCount(tierDTO.getMinPeopleCount());
            tier.setTargetPeopleCount(tierDTO.getTargetPeopleCount());
            tier.setGroupPrice(tierDTO.getGroupPrice());
            tier.setDiscountRate(tierDTO.getDiscountRate());
            tier.setSortOrder(tierDTO.getSortOrder());
            tier.setCreateTime(LocalDateTime.now());
            priceTierMapper.insert(tier);
        }
    }
}

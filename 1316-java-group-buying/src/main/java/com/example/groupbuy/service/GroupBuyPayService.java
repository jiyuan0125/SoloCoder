package com.example.groupbuy.service;

import com.baomidou.mybatisplus.core.conditions.query.LambdaQueryWrapper;
import com.baomidou.mybatisplus.extension.service.impl.ServiceImpl;
import com.example.groupbuy.common.GroupBuyConstants;
import com.example.groupbuy.dto.PayGroupBuyDTO;
import com.example.groupbuy.entity.*;
import com.example.groupbuy.exception.GroupBuyException;
import com.example.groupbuy.mapper.*;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.data.redis.core.StringRedisTemplate;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.math.BigDecimal;
import java.time.LocalDateTime;
import java.util.Comparator;
import java.util.List;
import java.util.Optional;
import java.util.concurrent.TimeUnit;

@Slf4j
@Service
@RequiredArgsConstructor
public class GroupBuyPayService extends ServiceImpl<GroupBuyParticipantMapper, GroupBuyParticipant> {
    
    private final GroupBuyOrderMapper orderMapper;
    private final GroupBuyActivityMapper activityMapper;
    private final GroupBuyPriceTierMapper priceTierMapper;
    private final GroupBuyQueueMapper queueMapper;
    private final GroupBuyRefundService refundService;
    private final StringRedisTemplate redisTemplate;
    
    private static final String PAY_LOCK_KEY_PREFIX = "group_buy:pay:lock:";
    private static final String GROUP_FULL_KEY_PREFIX = "group_buy:group:full:";
    
    @Transactional(rollbackFor = Exception.class)
    public GroupBuyParticipant pay(PayGroupBuyDTO dto) {
        GroupBuyOrder order = orderMapper.selectById(dto.getGroupOrderId());
        if (order == null) {
            throw new GroupBuyException("拼团订单不存在");
        }
        
        if (order.getStatus() != GroupBuyConstants.ORDER_STATUS_PENDING) {
            throw new GroupBuyException("拼团订单不可支付");
        }
        
        if (order.getIsQueued() == GroupBuyConstants.IS_QUEUED_YES) {
            throw new GroupBuyException("该拼团正在排队中，请等待库存释放");
        }
        
        LocalDateTime now = LocalDateTime.now();
        if (now.isAfter(order.getEndTime())) {
            throw new GroupBuyException("拼团已过期");
        }
        
        LambdaQueryWrapper<GroupBuyParticipant> participantWrapper = new LambdaQueryWrapper<>();
        participantWrapper.eq(GroupBuyParticipant::getGroupOrderId, order.getId())
                .eq(GroupBuyParticipant::getUserId, dto.getUserId());
        GroupBuyParticipant participant = this.getOne(participantWrapper);
        
        if (participant == null) {
            throw new GroupBuyException("您未加入该拼团");
        }
        
        if (participant.getPayStatus() == GroupBuyConstants.PAY_STATUS_PAID) {
            throw new GroupBuyException("您已支付该拼团");
        }
        
        if (participant.getPayStatus() == GroupBuyConstants.PAY_STATUS_REFUNDED) {
            throw new GroupBuyException("该拼团已退款");
        }
        
        List<GroupBuyPriceTier> priceTiers = priceTierMapper.selectList(
                new LambdaQueryWrapper<GroupBuyPriceTier>()
                        .eq(GroupBuyPriceTier::getActivityId, order.getActivityId())
                        .orderByAsc(GroupBuyPriceTier::getSortOrder));
        
        if (priceTiers.isEmpty()) {
            throw new GroupBuyException("拼团价格配置错误");
        }
        
        BigDecimal expectedPrice = getExpectedPrice(order.getTargetPeopleCount(), priceTiers);
        if (dto.getPayAmount().compareTo(expectedPrice) != 0) {
            throw new GroupBuyException("支付金额不正确");
        }
        
        String payLockKey = PAY_LOCK_KEY_PREFIX + order.getId() + ":" + dto.getUserId();
        Boolean payAcquired = redisTemplate.opsForValue().setIfAbsent(
                payLockKey, 
                "locked", 
                10, 
                TimeUnit.SECONDS
        );
        
        if (!Boolean.TRUE.equals(payAcquired)) {
            throw new GroupBuyException("支付处理中，请稍后重试");
        }
        
        try {
            String groupFullKey = GROUP_FULL_KEY_PREFIX + order.getId();
            if (Boolean.TRUE.equals(redisTemplate.hasKey(groupFullKey))) {
                log.info("拼团已满，用户{}支付后将立即退款，拼团订单ID: {}", dto.getUserId(), order.getId());
                return processImmediateRefund(participant, order, dto.getPayAmount());
            }
            
            order = orderMapper.selectById(order.getId());
            LambdaQueryWrapper<GroupBuyParticipant> paidWrapper = new LambdaQueryWrapper<>();
            paidWrapper.eq(GroupBuyParticipant::getGroupOrderId, order.getId())
                    .eq(GroupBuyParticipant::getPayStatus, GroupBuyConstants.PAY_STATUS_PAID);
            long paidCountLong = this.count(paidWrapper);
            int paidCount = (int) paidCountLong;
            
            if (paidCount >= order.getTargetPeopleCount()) {
                log.info("拼团已满，用户{}支付后将立即退款，拼团订单ID: {}", dto.getUserId(), order.getId());
                redisTemplate.opsForValue().set(groupFullKey, "full", 1, TimeUnit.HOURS);
                return processImmediateRefund(participant, order, dto.getPayAmount());
            }
            
            participant.setPayStatus(GroupBuyConstants.PAY_STATUS_PAID);
            participant.setPayAmount(dto.getPayAmount());
            participant.setPayTime(now);
            participant.setUpdateTime(now);
            this.updateById(participant);
            
            log.info("用户{}支付成功，拼团订单ID: {}", dto.getUserId(), order.getId());
            
            paidCount = paidCount + 1;
            if (paidCount >= order.getTargetPeopleCount()) {
                log.info("拼团已满，开始成团处理，拼团订单ID: {}", order.getId());
                redisTemplate.opsForValue().set(groupFullKey, "full", 1, TimeUnit.HOURS);
                processGroupSuccess(order, paidCount, priceTiers);
            }
            
            return participant;
            
        } finally {
            redisTemplate.delete(payLockKey);
        }
    }
    
    @Transactional(rollbackFor = Exception.class)
    public void processGroupSuccess(GroupBuyOrder order, int actualCount, List<GroupBuyPriceTier> priceTiers) {
        LocalDateTime now = LocalDateTime.now();
        
        GroupBuyPriceTier applicableTier = findApplicablePriceTier(actualCount, priceTiers);
        log.info("拼团订单ID: {} 实际人数: {}，适用价格档位: {} 人团，价格: {}", 
                order.getId(), actualCount, applicableTier.getTargetPeopleCount(), applicableTier.getGroupPrice());
        
        order.setActualPeopleCount(actualCount);
        order.setGroupPrice(applicableTier.getGroupPrice());
        order.setStatus(GroupBuyConstants.ORDER_STATUS_SUCCESS);
        order.setUpdateTime(now);
        orderMapper.updateById(order);
        
        LambdaQueryWrapper<GroupBuyParticipant> participantWrapper = new LambdaQueryWrapper<>();
        participantWrapper.eq(GroupBuyParticipant::getGroupOrderId, order.getId());
        List<GroupBuyParticipant> participants = this.list(participantWrapper);
        
        for (GroupBuyParticipant participant : participants) {
            if (participant.getPayStatus() == GroupBuyConstants.PAY_STATUS_PAID) {
                BigDecimal priceDiff = participant.getPayAmount().subtract(applicableTier.getGroupPrice());
                if (priceDiff.compareTo(BigDecimal.ZERO) > 0) {
                    log.info("用户{}支付价格{}高于成团价格{}，需要退款差价{}", 
                            participant.getUserId(), participant.getPayAmount(), applicableTier.getGroupPrice(), priceDiff);
                    refundService.processPartialRefund(participant, priceDiff);
                }
            }
        }
        
        GroupBuyActivity activity = activityMapper.selectById(order.getActivityId());
        if (activity.getProductType() == GroupBuyConstants.PRODUCT_TYPE_VIRTUAL) {
            log.info("虚拟商品拼团成功，开始发放商品，拼团订单ID: {}", order.getId());
            deliverVirtualProducts(order);
        }
        
        log.info("拼团成功处理完成，拼团订单ID: {}", order.getId());
    }
    
    private GroupBuyParticipant processImmediateRefund(GroupBuyParticipant participant, GroupBuyOrder order, BigDecimal payAmount) {
        LocalDateTime now = LocalDateTime.now();
        
        participant.setPayStatus(GroupBuyConstants.PAY_STATUS_REFUNDED);
        participant.setPayAmount(payAmount);
        participant.setPayTime(now);
        participant.setRefundAmount(payAmount);
        participant.setRefundTime(now);
        participant.setUpdateTime(now);
        this.updateById(participant);
        
        orderMapper.decrementPeopleCount(order.getId());
        
        refundService.triggerRefund(participant);
        
        log.info("拼团已满，用户{}支付已立即退款，拼团订单ID: {}", participant.getUserId(), order.getId());
        return participant;
    }
    
    private BigDecimal getExpectedPrice(int targetPeopleCount, List<GroupBuyPriceTier> priceTiers) {
        Optional<GroupBuyPriceTier> tier = priceTiers.stream()
                .filter(t -> t.getTargetPeopleCount() == targetPeopleCount)
                .findFirst();
        
        return tier.map(GroupBuyPriceTier::getGroupPrice)
                .orElseThrow(() -> new GroupBuyException("价格档位配置错误"));
    }
    
    private GroupBuyPriceTier findApplicablePriceTier(int actualCount, List<GroupBuyPriceTier> priceTiers) {
        List<GroupBuyPriceTier> sortedTiers = priceTiers.stream()
                .sorted(Comparator.comparingInt(GroupBuyPriceTier::getTargetPeopleCount))
                .collect(java.util.stream.Collectors.toList());
        
        for (int i = sortedTiers.size() - 1; i >= 0; i--) {
            GroupBuyPriceTier tier = sortedTiers.get(i);
            if (actualCount >= tier.getMinPeopleCount()) {
                return tier;
            }
        }
        
        return sortedTiers.get(0);
    }
    
    private void deliverVirtualProducts(GroupBuyOrder order) {
        LambdaQueryWrapper<GroupBuyParticipant> wrapper = new LambdaQueryWrapper<>();
        wrapper.eq(GroupBuyParticipant::getGroupOrderId, order.getId())
                .eq(GroupBuyParticipant::getPayStatus, GroupBuyConstants.PAY_STATUS_PAID);
        
        List<GroupBuyParticipant> participants = this.list(wrapper);
        
        for (GroupBuyParticipant participant : participants) {
            if (participant.getProductDelivered() == GroupBuyConstants.PRODUCT_DELIVERED_NO) {
                participant.setProductDelivered(GroupBuyConstants.PRODUCT_DELIVERED_YES);
                participant.setUpdateTime(LocalDateTime.now());
                this.updateById(participant);
                log.info("虚拟商品已发放给用户{}，拼团订单ID: {}", participant.getUserId(), order.getId());
            }
        }
    }
    
    public List<GroupBuyParticipant> getPaidParticipants(Long groupOrderId) {
        return this.list(new LambdaQueryWrapper<GroupBuyParticipant>()
                .eq(GroupBuyParticipant::getGroupOrderId, groupOrderId)
                .eq(GroupBuyParticipant::getPayStatus, GroupBuyConstants.PAY_STATUS_PAID));
    }
}

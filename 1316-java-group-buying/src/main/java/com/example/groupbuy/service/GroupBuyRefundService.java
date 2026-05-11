package com.example.groupbuy.service;

import com.baomidou.mybatisplus.core.conditions.query.LambdaQueryWrapper;
import com.example.groupbuy.common.GroupBuyConstants;
import com.example.groupbuy.entity.*;
import com.example.groupbuy.mapper.*;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.data.redis.core.StringRedisTemplate;
import org.springframework.scheduling.annotation.Async;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.math.BigDecimal;
import java.time.LocalDateTime;
import java.util.List;

@Slf4j
@Service
@RequiredArgsConstructor
public class GroupBuyRefundService {
    
    private final GroupBuyParticipantMapper participantMapper;
    private final GroupBuyOrderMapper orderMapper;
    private final GroupBuyActivityMapper activityMapper;
    private final GroupBuyQueueService queueService;
    private final StringRedisTemplate redisTemplate;
    
    private static final String USER_LIMIT_KEY_PREFIX = "group_buy:limit:user:";
    
    @Async
    @Transactional(rollbackFor = Exception.class)
    public void triggerRefund(GroupBuyParticipant participant) {
        log.info("开始触发退款，参与者ID: {}", participant.getId());
        
        try {
            Thread.sleep(100);
            processRefundLogic(participant);
            log.info("退款处理完成，参与者ID: {}", participant.getId());
        } catch (InterruptedException e) {
            Thread.currentThread().interrupt();
            log.error("退款处理被中断，参与者ID: {}", participant.getId(), e);
            throw new RuntimeException("退款处理被中断", e);
        } catch (Exception e) {
            log.error("退款处理失败，参与者ID: {}", participant.getId(), e);
            throw new RuntimeException("退款处理失败", e);
        }
    }
    
    @Transactional(rollbackFor = Exception.class)
    public void processPartialRefund(GroupBuyParticipant participant, BigDecimal refundAmount) {
        if (refundAmount.compareTo(BigDecimal.ZERO) <= 0) {
            return;
        }
        
        LocalDateTime now = LocalDateTime.now();
        
        BigDecimal newRefundAmount = participant.getRefundAmount() != null 
                ? participant.getRefundAmount().add(refundAmount) 
                : refundAmount;
        
        participant.setRefundAmount(newRefundAmount);
        participant.setRefundTime(now);
        participant.setUpdateTime(now);
        participantMapper.updateById(participant);
        
        log.info("处理差价退款，用户ID: {}，退款金额: {}", participant.getUserId(), refundAmount);
        
        executeRefund(participant, refundAmount);
    }
    
    @Transactional(rollbackFor = Exception.class)
    public void cancelGroupBuyOrder(Long groupOrderId) {
        GroupBuyOrder order = orderMapper.selectById(groupOrderId);
        if (order == null) {
            return;
        }
        
        if (order.getStatus() != GroupBuyConstants.ORDER_STATUS_PENDING) {
            return;
        }
        
        log.info("开始取消拼团订单，订单ID: {}", groupOrderId);
        
        order.setStatus(GroupBuyConstants.ORDER_STATUS_CANCELLED);
        order.setUpdateTime(LocalDateTime.now());
        orderMapper.updateById(order);
        
        GroupBuyActivity activity = activityMapper.selectById(order.getActivityId());
        if (activity != null && activity.getUsedGroupCount() > 0) {
            activity.setUsedGroupCount(activity.getUsedGroupCount() - 1);
            activityMapper.updateById(activity);
            log.info("释放活动库存，活动ID: {}", activity.getId());
            
            queueService.processQueueOnStockRelease(activity.getId());
        }
        
        LambdaQueryWrapper<GroupBuyParticipant> wrapper = new LambdaQueryWrapper<>();
        wrapper.eq(GroupBuyParticipant::getGroupOrderId, groupOrderId)
                .in(GroupBuyParticipant::getPayStatus, 
                    GroupBuyConstants.PAY_STATUS_PENDING, 
                    GroupBuyConstants.PAY_STATUS_PAID);
        
        List<GroupBuyParticipant> participants = participantMapper.selectList(wrapper);
        
        for (GroupBuyParticipant participant : participants) {
            if (participant.getPayStatus() == GroupBuyConstants.PAY_STATUS_PAID) {
                processFullRefund(participant);
            }
            releaseUserLimitLock(order.getActivityId(), participant.getUserId());
        }
        
        log.info("拼团订单取消完成，订单ID: {}", groupOrderId);
    }
    
    @Transactional(rollbackFor = Exception.class)
    public void processFullRefund(GroupBuyParticipant participant) {
        if (participant.getPayStatus() != GroupBuyConstants.PAY_STATUS_PAID) {
            return;
        }
        
        LocalDateTime now = LocalDateTime.now();
        
        participant.setPayStatus(GroupBuyConstants.PAY_STATUS_REFUNDED);
        participant.setRefundAmount(participant.getPayAmount());
        participant.setRefundTime(now);
        participant.setUpdateTime(now);
        participantMapper.updateById(participant);
        
        log.info("处理全额退款，用户ID: {}，退款金额: {}", participant.getUserId(), participant.getPayAmount());
        
        executeRefund(participant, participant.getPayAmount());
    }
    
    public boolean canRefund(GroupBuyParticipant participant, GroupBuyOrder order, GroupBuyActivity activity) {
        if (participant.getPayStatus() != GroupBuyConstants.PAY_STATUS_PAID) {
            return false;
        }
        
        if (activity.getProductType() == GroupBuyConstants.PRODUCT_TYPE_VIRTUAL) {
            if (order.getStatus() == GroupBuyConstants.ORDER_STATUS_SUCCESS 
                    && participant.getProductDelivered() == GroupBuyConstants.PRODUCT_DELIVERED_YES) {
                return false;
            }
        } else {
            if (order.getStatus() == GroupBuyConstants.ORDER_STATUS_SUCCESS 
                    && participant.getProductDelivered() == GroupBuyConstants.PRODUCT_DELIVERED_YES) {
                return false;
            }
        }
        
        return true;
    }
    
    @Transactional(rollbackFor = Exception.class)
    public boolean userApplyRefund(Long participantId) {
        GroupBuyParticipant participant = participantMapper.selectById(participantId);
        if (participant == null) {
            throw new RuntimeException("参与者不存在");
        }
        
        GroupBuyOrder order = orderMapper.selectById(participant.getGroupOrderId());
        if (order == null) {
            throw new RuntimeException("拼团订单不存在");
        }
        
        GroupBuyActivity activity = activityMapper.selectById(order.getActivityId());
        if (activity == null) {
            throw new RuntimeException("拼团活动不存在");
        }
        
        if (!canRefund(participant, order, activity)) {
            return false;
        }
        
        processFullRefund(participant);
        releaseUserLimitLock(order.getActivityId(), participant.getUserId());
        
        orderMapper.decrementPeopleCount(order.getId());
        
        log.info("用户申请退款成功，参与者ID: {}", participantId);
        return true;
    }
    
    private void processRefundLogic(GroupBuyParticipant participant) {
        participant = participantMapper.selectById(participant.getId());
        if (participant == null) {
            return;
        }
        
        if (participant.getPayStatus() != GroupBuyConstants.PAY_STATUS_REFUNDED) {
            participant.setPayStatus(GroupBuyConstants.PAY_STATUS_REFUNDED);
            participant.setRefundAmount(participant.getPayAmount());
            participant.setRefundTime(LocalDateTime.now());
            participant.setUpdateTime(LocalDateTime.now());
            participantMapper.updateById(participant);
        }
        
        executeRefund(participant, participant.getPayAmount());
    }
    
    private void executeRefund(GroupBuyParticipant participant, BigDecimal refundAmount) {
        log.info("执行退款操作，用户ID: {}，退款金额: {}", participant.getUserId(), refundAmount);
    }
    
    private void releaseUserLimitLock(Long activityId, Long userId) {
        String key = USER_LIMIT_KEY_PREFIX + activityId + ":" + userId;
        redisTemplate.delete(key);
        log.info("释放用户限购锁，用户ID: {}, 活动ID: {}", userId, activityId);
    }
}

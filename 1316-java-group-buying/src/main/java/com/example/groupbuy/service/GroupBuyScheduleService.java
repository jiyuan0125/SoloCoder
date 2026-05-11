package com.example.groupbuy.service;

import com.baomidou.mybatisplus.core.conditions.query.LambdaQueryWrapper;
import com.example.groupbuy.common.GroupBuyConstants;
import com.example.groupbuy.entity.GroupBuyOrder;
import com.example.groupbuy.entity.GroupBuyQueue;
import com.example.groupbuy.mapper.GroupBuyOrderMapper;
import com.example.groupbuy.mapper.GroupBuyQueueMapper;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.scheduling.annotation.Scheduled;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.time.LocalDateTime;
import java.util.List;

@Slf4j
@Service
@RequiredArgsConstructor
public class GroupBuyScheduleService {
    
    private final GroupBuyOrderMapper orderMapper;
    private final GroupBuyQueueMapper queueMapper;
    private final GroupBuyRefundService refundService;
    private final GroupBuyActivityService activityService;
    
    @Scheduled(fixedDelayString = "${group-buy.check-expired-interval-seconds:30000}")
    @Transactional(rollbackFor = Exception.class)
    public void checkExpiredOrders() {
        LocalDateTime now = LocalDateTime.now();
        log.debug("开始检查过期拼团订单，当前时间: {}", now);
        
        List<GroupBuyOrder> expiredOrders = orderMapper.findExpiredOrders(now);
        
        if (expiredOrders.isEmpty()) {
            return;
        }
        
        log.info("发现{}个过期拼团订单需要处理", expiredOrders.size());
        
        for (GroupBuyOrder order : expiredOrders) {
            try {
                processExpiredOrder(order);
            } catch (Exception e) {
                log.error("处理过期拼团订单失败，订单ID: {}", order.getId(), e);
            }
        }
        
        log.info("过期拼团订单处理完成");
    }
    
    @Scheduled(fixedDelay = 60000)
    public void updateActivityStatus() {
        try {
            activityService.updateActivityStatus();
        } catch (Exception e) {
            log.error("更新活动状态失败", e);
        }
    }
    
    @Scheduled(fixedDelay = 30000)
    @Transactional(rollbackFor = Exception.class)
    public void processQueuedOrders() {
        List<GroupBuyQueue> waitingQueues = queueMapper.selectList(
                new LambdaQueryWrapper<GroupBuyQueue>()
                        .eq(GroupBuyQueue::getQueueStatus, GroupBuyConstants.QUEUE_STATUS_WAITING)
                        .orderByAsc(GroupBuyQueue::getQueueTime));
        
        if (waitingQueues.isEmpty()) {
            return;
        }
        
        log.info("发现{}个排队中的拼团订单", waitingQueues.size());
        
        for (GroupBuyQueue queue : waitingQueues) {
            try {
                processQueueItem(queue);
            } catch (Exception e) {
                log.error("处理排队订单失败，队列ID: {}", queue.getId(), e);
            }
        }
    }
    
    private void processExpiredOrder(GroupBuyOrder order) {
        log.info("处理过期拼团订单，订单ID: {}", order.getId());
        
        refundService.cancelGroupBuyOrder(order.getId());
        
        log.info("过期拼团订单处理完成，订单ID: {}", order.getId());
    }
    
    private void processQueueItem(GroupBuyQueue queue) {
        GroupBuyOrder order = orderMapper.selectById(queue.getGroupOrderId());
        if (order == null) {
            queue.setQueueStatus(GroupBuyConstants.QUEUE_STATUS_CANCELLED);
            queue.setProcessTime(LocalDateTime.now());
            queueMapper.updateById(queue);
            return;
        }
        
        if (order.getStatus() != GroupBuyConstants.ORDER_STATUS_PENDING 
                || order.getIsQueued() == GroupBuyConstants.IS_QUEUED_NO) {
            queue.setQueueStatus(GroupBuyConstants.QUEUE_STATUS_CANCELLED);
            queue.setProcessTime(LocalDateTime.now());
            queueMapper.updateById(queue);
            return;
        }
        
        if (LocalDateTime.now().isAfter(order.getEndTime())) {
            queue.setQueueStatus(GroupBuyConstants.QUEUE_STATUS_CANCELLED);
            queue.setProcessTime(LocalDateTime.now());
            queueMapper.updateById(queue);
            refundService.cancelGroupBuyOrder(order.getId());
            return;
        }
        
        log.info("处理排队订单，队列ID: {}，拼团订单ID: {}", queue.getId(), queue.getGroupOrderId());
    }
}

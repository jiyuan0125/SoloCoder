package com.example.groupbuy.service;

import com.baomidou.mybatisplus.core.conditions.query.LambdaQueryWrapper;
import com.example.groupbuy.common.GroupBuyConstants;
import com.example.groupbuy.entity.GroupBuyActivity;
import com.example.groupbuy.entity.GroupBuyOrder;
import com.example.groupbuy.entity.GroupBuyQueue;
import com.example.groupbuy.mapper.GroupBuyActivityMapper;
import com.example.groupbuy.mapper.GroupBuyOrderMapper;
import com.example.groupbuy.mapper.GroupBuyQueueMapper;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.data.redis.core.StringRedisTemplate;
import org.springframework.scheduling.annotation.Async;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.time.LocalDateTime;
import java.util.List;
import java.util.concurrent.TimeUnit;

@Slf4j
@Service
@RequiredArgsConstructor
public class GroupBuyQueueService {
    
    private final GroupBuyQueueMapper queueMapper;
    private final GroupBuyOrderMapper orderMapper;
    private final GroupBuyActivityMapper activityMapper;
    private final StringRedisTemplate redisTemplate;
    
    private static final String STOCK_LOCK_KEY_PREFIX = "group_buy:stock:lock:";
    private static final String QUEUE_PROCESSING_KEY_PREFIX = "group_buy:queue:processing:";
    
    @Async
    @Transactional(rollbackFor = Exception.class)
    public void processQueueOnStockRelease(Long activityId) {
        log.info("库存释放，开始处理排队订单，活动ID: {}", activityId);
        
        String processingKey = QUEUE_PROCESSING_KEY_PREFIX + activityId;
        Boolean processingLock = redisTemplate.opsForValue().setIfAbsent(
                processingKey, 
                "processing", 
                30, 
                TimeUnit.SECONDS
        );
        
        if (!Boolean.TRUE.equals(processingLock)) {
            log.info("排队订单正在处理中，跳过，活动ID: {}", activityId);
            return;
        }
        
        try {
            GroupBuyActivity activity = activityMapper.selectById(activityId);
            if (activity == null) {
                return;
            }
            
            while (activity.getUsedGroupCount() < activity.getMaxGroupCount()) {
                GroupBuyQueue nextQueue = getNextWaitingQueue(activityId);
                if (nextQueue == null) {
                    log.info("没有更多排队订单，活动ID: {}", activityId);
                    break;
                }
                
                if (!processNextQueueItem(nextQueue, activity)) {
                    break;
                }
                
                activity = activityMapper.selectById(activityId);
            }
            
        } finally {
            redisTemplate.delete(processingKey);
        }
        
        log.info("排队订单处理完成，活动ID: {}", activityId);
    }
    
    @Transactional(rollbackFor = Exception.class)
    public boolean processNextQueueItem(GroupBuyQueue queue, GroupBuyActivity activity) {
        log.info("处理下一个排队订单，队列ID: {}，活动ID: {}", queue.getId(), activity.getId());
        
        String stockLockKey = STOCK_LOCK_KEY_PREFIX + activity.getId();
        Boolean stockAcquired = redisTemplate.opsForValue().setIfAbsent(
                stockLockKey, 
                "locked", 
                3, 
                TimeUnit.SECONDS
        );
        
        if (!Boolean.TRUE.equals(stockAcquired)) {
            return false;
        }
        
        try {
            activity = activityMapper.selectById(activity.getId());
            if (activity.getUsedGroupCount() >= activity.getMaxGroupCount()) {
                log.info("库存已满，无法处理排队订单，活动ID: {}", activity.getId());
                return false;
            }
            
            GroupBuyOrder order = orderMapper.selectById(queue.getGroupOrderId());
            if (order == null) {
                cancelQueue(queue);
                return true;
            }
            
            if (order.getStatus() != GroupBuyConstants.ORDER_STATUS_PENDING 
                    || order.getIsQueued() == GroupBuyConstants.IS_QUEUED_NO) {
                cancelQueue(queue);
                return true;
            }
            
            if (LocalDateTime.now().isAfter(order.getEndTime())) {
                cancelQueue(queue);
                return true;
            }
            
            activity.setUsedGroupCount(activity.getUsedGroupCount() + 1);
            activityMapper.updateById(activity);
            
            order.setIsQueued(GroupBuyConstants.IS_QUEUED_NO);
            order.setUpdateTime(LocalDateTime.now());
            orderMapper.updateById(order);
            
            queue.setQueueStatus(GroupBuyConstants.QUEUE_STATUS_PROCESSED);
            queue.setProcessTime(LocalDateTime.now());
            queueMapper.updateById(queue);
            
            log.info("排队订单处理成功，拼团订单ID: {}，活动ID: {}", order.getId(), activity.getId());
            return true;
            
        } finally {
            redisTemplate.delete(stockLockKey);
        }
    }
    
    public GroupBuyQueue getNextWaitingQueue(Long activityId) {
        List<GroupBuyQueue> queues = queueMapper.selectList(
                new LambdaQueryWrapper<GroupBuyQueue>()
                        .eq(GroupBuyQueue::getActivityId, activityId)
                        .eq(GroupBuyQueue::getQueueStatus, GroupBuyConstants.QUEUE_STATUS_WAITING)
                        .orderByAsc(GroupBuyQueue::getQueueTime)
                        .last("LIMIT 1")
        );
        
        return queues.isEmpty() ? null : queues.get(0);
    }
    
    public List<GroupBuyQueue> getWaitingQueues(Long activityId) {
        return queueMapper.selectList(
                new LambdaQueryWrapper<GroupBuyQueue>()
                        .eq(GroupBuyQueue::getActivityId, activityId)
                        .eq(GroupBuyQueue::getQueueStatus, GroupBuyConstants.QUEUE_STATUS_WAITING)
                        .orderByAsc(GroupBuyQueue::getQueueTime)
        );
    }
    
    public int getQueuePosition(Long activityId, Long groupOrderId) {
        List<GroupBuyQueue> waitingQueues = getWaitingQueues(activityId);
        
        for (int i = 0; i < waitingQueues.size(); i++) {
            if (waitingQueues.get(i).getGroupOrderId().equals(groupOrderId)) {
                return i + 1;
            }
        }
        
        return -1;
    }
    
    @Transactional(rollbackFor = Exception.class)
    public boolean cancelQueuedOrder(Long groupOrderId) {
        GroupBuyOrder order = orderMapper.selectById(groupOrderId);
        if (order == null) {
            return false;
        }
        
        if (order.getIsQueued() != GroupBuyConstants.IS_QUEUED_YES) {
            return false;
        }
        
        LambdaQueryWrapper<GroupBuyQueue> wrapper = new LambdaQueryWrapper<>();
        wrapper.eq(GroupBuyQueue::getGroupOrderId, groupOrderId)
                .eq(GroupBuyQueue::getQueueStatus, GroupBuyConstants.QUEUE_STATUS_WAITING);
        GroupBuyQueue queue = queueMapper.selectOne(wrapper);
        
        if (queue != null) {
            cancelQueue(queue);
        }
        
        order.setStatus(GroupBuyConstants.ORDER_STATUS_CANCELLED);
        order.setUpdateTime(LocalDateTime.now());
        orderMapper.updateById(order);
        
        log.info("用户取消排队的拼团订单，订单ID: {}", groupOrderId);
        return true;
    }
    
    private void cancelQueue(GroupBuyQueue queue) {
        queue.setQueueStatus(GroupBuyConstants.QUEUE_STATUS_CANCELLED);
        queue.setProcessTime(LocalDateTime.now());
        queueMapper.updateById(queue);
        log.info("取消排队订单，队列ID: {}", queue.getId());
    }
}

package com.example.groupbuy.service;

import com.baomidou.mybatisplus.core.conditions.query.LambdaQueryWrapper;
import com.baomidou.mybatisplus.extension.service.impl.ServiceImpl;
import com.example.groupbuy.common.GroupBuyConstants;
import com.example.groupbuy.dto.CreateGroupBuyDTO;
import com.example.groupbuy.dto.JoinGroupBuyDTO;
import com.example.groupbuy.entity.*;
import com.example.groupbuy.exception.GroupBuyException;
import com.example.groupbuy.mapper.*;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.data.redis.core.StringRedisTemplate;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.time.LocalDateTime;
import java.util.List;
import java.util.UUID;
import java.util.concurrent.TimeUnit;

@Slf4j
@Service
@RequiredArgsConstructor
public class GroupBuyOrderService extends ServiceImpl<GroupBuyOrderMapper, GroupBuyOrder> {
    
    private final GroupBuyActivityService activityService;
    private final GroupBuyParticipantMapper participantMapper;
    private final GroupBuyPriceTierMapper priceTierMapper;
    private final GroupBuyQueueMapper queueMapper;
    private final StringRedisTemplate redisTemplate;
    
    private static final String USER_LIMIT_KEY_PREFIX = "group_buy:limit:user:";
    private static final String STOCK_LOCK_KEY_PREFIX = "group_buy:stock:lock:";
    
    @Transactional(rollbackFor = Exception.class)
    public GroupBuyOrder createGroupBuy(CreateGroupBuyDTO dto) {
        GroupBuyActivity activity = activityService.getActivityDetail(dto.getActivityId());
        
        if (activity.getStatus() != GroupBuyConstants.ACTIVITY_STATUS_ONGOING) {
            throw new GroupBuyException("拼团活动未开始或已结束");
        }
        
        if (dto.getValidityHours() <= 0 || dto.getValidityHours() > 72) {
            throw new GroupBuyException("拼团有效期必须在1-72小时之间");
        }
        
        List<GroupBuyPriceTier> priceTiers = activityService.getPriceTiers(activity.getId());
        GroupBuyPriceTier selectedTier = priceTiers.stream()
                .filter(tier -> tier.getTargetPeopleCount().equals(dto.getTargetPeopleCount()))
                .findFirst()
                .orElseThrow(() -> new GroupBuyException("所选人数档位不存在"));
        
        String userLimitKey = USER_LIMIT_KEY_PREFIX + activity.getId() + ":" + dto.getUserId();
        if (Boolean.TRUE.equals(redisTemplate.hasKey(userLimitKey))) {
            throw new GroupBuyException("您已参与过该拼团活动，请选择其他活动");
        }
        
        if (hasUserParticipated(activity.getId(), dto.getUserId())) {
            throw new GroupBuyException("您已参与过该拼团活动，请选择其他活动");
        }
        
        String stockLockKey = STOCK_LOCK_KEY_PREFIX + activity.getId();
        Boolean stockAcquired = redisTemplate.opsForValue().setIfAbsent(
                stockLockKey, 
                "locked", 
                3, 
                TimeUnit.SECONDS
        );
        
        try {
            if (!Boolean.TRUE.equals(stockAcquired)) {
                return queueForStock(activity, dto);
            }
            
            activity = activityService.getById(activity.getId());
            if (activity.getUsedGroupCount() >= activity.getMaxGroupCount()) {
                return queueForStock(activity, dto);
            }
            
            activity.setUsedGroupCount(activity.getUsedGroupCount() + 1);
            activityService.updateById(activity);
            
        } finally {
            if (Boolean.TRUE.equals(stockAcquired)) {
                redisTemplate.delete(stockLockKey);
            }
        }
        
        LocalDateTime now = LocalDateTime.now();
        LocalDateTime endTime = now.plusHours(dto.getValidityHours());
        
        GroupBuyOrder order = new GroupBuyOrder();
        order.setOrderNo(generateOrderNo());
        order.setActivityId(activity.getId());
        order.setProductId(activity.getProductId());
        order.setLeaderUserId(dto.getUserId());
        order.setTargetPeopleCount(selectedTier.getTargetPeopleCount());
        order.setActualPeopleCount(0);
        order.setValidityHours(dto.getValidityHours());
        order.setStartTime(now);
        order.setEndTime(endTime);
        order.setStatus(GroupBuyConstants.ORDER_STATUS_PENDING);
        order.setIsQueued(GroupBuyConstants.IS_QUEUED_NO);
        order.setCreateTime(now);
        order.setUpdateTime(now);
        
        this.save(order);
        
        GroupBuyParticipant participant = new GroupBuyParticipant();
        participant.setGroupOrderId(order.getId());
        participant.setUserId(dto.getUserId());
        participant.setIsLeader(GroupBuyConstants.IS_LEADER_YES);
        participant.setJoinTime(now);
        participant.setPayStatus(GroupBuyConstants.PAY_STATUS_PENDING);
        participant.setProductDelivered(GroupBuyConstants.PRODUCT_DELIVERED_NO);
        participant.setCreateTime(now);
        participant.setUpdateTime(now);
        participantMapper.insert(participant);
        
        setUserLimitLock(activity.getId(), dto.getUserId(), endTime);
        
        log.info("用户{}发起拼团成功，拼团订单ID: {}", dto.getUserId(), order.getId());
        return order;
    }
    
    @Transactional(rollbackFor = Exception.class)
    public GroupBuyParticipant joinGroupBuy(JoinGroupBuyDTO dto) {
        GroupBuyOrder order = this.getById(dto.getGroupOrderId());
        if (order == null) {
            throw new GroupBuyException("拼团订单不存在");
        }
        
        if (order.getStatus() != GroupBuyConstants.ORDER_STATUS_PENDING) {
            throw new GroupBuyException("拼团订单不可加入");
        }
        
        if (order.getIsQueued() == GroupBuyConstants.IS_QUEUED_YES) {
            throw new GroupBuyException("该拼团正在排队中，暂时无法加入");
        }
        
        LocalDateTime now = LocalDateTime.now();
        if (now.isAfter(order.getEndTime())) {
            throw new GroupBuyException("拼团已过期");
        }
        
        GroupBuyActivity activity = activityService.getById(order.getActivityId());
        if (activity.getStatus() != GroupBuyConstants.ACTIVITY_STATUS_ONGOING) {
            throw new GroupBuyException("拼团活动已结束");
        }
        
        String userLimitKey = USER_LIMIT_KEY_PREFIX + activity.getId() + ":" + dto.getUserId();
        if (Boolean.TRUE.equals(redisTemplate.hasKey(userLimitKey))) {
            throw new GroupBuyException("您已参与过该拼团活动，请选择其他活动");
        }
        
        if (hasUserParticipated(activity.getId(), dto.getUserId())) {
            throw new GroupBuyException("您已参与过该拼团活动，请选择其他活动");
        }
        
        if (dto.getUserId().equals(order.getLeaderUserId())) {
            throw new GroupBuyException("不能加入自己发起的拼团");
        }
        
        LambdaQueryWrapper<GroupBuyParticipant> participantWrapper = new LambdaQueryWrapper<>();
        participantWrapper.eq(GroupBuyParticipant::getGroupOrderId, order.getId())
                .eq(GroupBuyParticipant::getUserId, dto.getUserId());
        if (participantMapper.selectCount(participantWrapper) > 0) {
            throw new GroupBuyException("您已经加入该拼团");
        }
        
        String orderLockKey = "group_buy:order:lock:" + order.getId();
        Boolean orderAcquired = redisTemplate.opsForValue().setIfAbsent(
                orderLockKey, 
                "locked", 
                3, 
                TimeUnit.SECONDS
        );
        
        if (!Boolean.TRUE.equals(orderAcquired)) {
            throw new GroupBuyException("拼团人数已满，请选择其他拼团");
        }
        
        try {
            order = this.getById(order.getId());
            if (order.getActualPeopleCount() >= order.getTargetPeopleCount()) {
                throw new GroupBuyException("拼团人数已满，请选择其他拼团");
            }
            
            int affected = baseMapper.incrementPeopleCount(order.getId());
            if (affected == 0) {
                throw new GroupBuyException("拼团人数已满，请选择其他拼团");
            }
            
        } finally {
            redisTemplate.delete(orderLockKey);
        }
        
        GroupBuyParticipant participant = new GroupBuyParticipant();
        participant.setGroupOrderId(order.getId());
        participant.setUserId(dto.getUserId());
        participant.setIsLeader(GroupBuyConstants.IS_LEADER_NO);
        participant.setJoinTime(now);
        participant.setPayStatus(GroupBuyConstants.PAY_STATUS_PENDING);
        participant.setProductDelivered(GroupBuyConstants.PRODUCT_DELIVERED_NO);
        participant.setCreateTime(now);
        participant.setUpdateTime(now);
        participantMapper.insert(participant);
        
        setUserLimitLock(activity.getId(), dto.getUserId(), order.getEndTime());
        
        log.info("用户{}加入拼团成功，拼团订单ID: {}", dto.getUserId(), order.getId());
        return participant;
    }
    
    public List<GroupBuyOrder> getAvailableOrders(Long activityId, int page, int size) {
        LambdaQueryWrapper<GroupBuyOrder> wrapper = new LambdaQueryWrapper<>();
        wrapper.eq(GroupBuyOrder::getActivityId, activityId)
                .eq(GroupBuyOrder::getStatus, GroupBuyConstants.ORDER_STATUS_PENDING)
                .eq(GroupBuyOrder::getIsQueued, GroupBuyConstants.IS_QUEUED_NO)
                .gt(GroupBuyOrder::getEndTime, LocalDateTime.now())
                .apply("actual_people_count < target_people_count")
                .orderByDesc(GroupBuyOrder::getCreateTime);
        
        return this.list(wrapper);
    }
    
    public GroupBuyOrder getOrderDetail(Long orderId) {
        return this.getById(orderId);
    }
    
    public List<GroupBuyParticipant> getOrderParticipants(Long orderId) {
        return participantMapper.selectList(new LambdaQueryWrapper<GroupBuyParticipant>()
                .eq(GroupBuyParticipant::getGroupOrderId, orderId));
    }
    
    private GroupBuyOrder queueForStock(GroupBuyActivity activity, CreateGroupBuyDTO dto) {
        LocalDateTime now = LocalDateTime.now();
        LocalDateTime endTime = now.plusHours(dto.getValidityHours());
        
        GroupBuyOrder order = new GroupBuyOrder();
        order.setOrderNo(generateOrderNo());
        order.setActivityId(activity.getId());
        order.setProductId(activity.getProductId());
        order.setLeaderUserId(dto.getUserId());
        order.setTargetPeopleCount(dto.getTargetPeopleCount());
        order.setActualPeopleCount(0);
        order.setValidityHours(dto.getValidityHours());
        order.setStartTime(now);
        order.setEndTime(endTime);
        order.setStatus(GroupBuyConstants.ORDER_STATUS_PENDING);
        order.setIsQueued(GroupBuyConstants.IS_QUEUED_YES);
        order.setCreateTime(now);
        order.setUpdateTime(now);
        
        this.save(order);
        
        Long queueOrder = queueMapper.selectCount(new LambdaQueryWrapper<GroupBuyQueue>()
                .eq(GroupBuyQueue::getActivityId, activity.getId())
                .eq(GroupBuyQueue::getQueueStatus, GroupBuyConstants.QUEUE_STATUS_WAITING)) + 1;
        
        order.setQueueOrder(queueOrder.intValue());
        this.updateById(order);
        
        GroupBuyQueue queue = new GroupBuyQueue();
        queue.setGroupOrderId(order.getId());
        queue.setActivityId(activity.getId());
        queue.setUserId(dto.getUserId());
        queue.setQueueStatus(GroupBuyConstants.QUEUE_STATUS_WAITING);
        queue.setQueueTime(now);
        queue.setCreateTime(now);
        queueMapper.insert(queue);
        
        GroupBuyParticipant participant = new GroupBuyParticipant();
        participant.setGroupOrderId(order.getId());
        participant.setUserId(dto.getUserId());
        participant.setIsLeader(GroupBuyConstants.IS_LEADER_YES);
        participant.setJoinTime(now);
        participant.setPayStatus(GroupBuyConstants.PAY_STATUS_PENDING);
        participant.setProductDelivered(GroupBuyConstants.PRODUCT_DELIVERED_NO);
        participant.setCreateTime(now);
        participant.setUpdateTime(now);
        participantMapper.insert(participant);
        
        setUserLimitLock(activity.getId(), dto.getUserId(), endTime);
        
        log.info("用户{}发起拼团，库存不足已排队，拼团订单ID: {}", dto.getUserId(), order.getId());
        return order;
    }
    
    private boolean hasUserParticipated(Long activityId, Long userId) {
        LambdaQueryWrapper<GroupBuyOrder> orderWrapper = new LambdaQueryWrapper<>();
        orderWrapper.eq(GroupBuyOrder::getActivityId, activityId)
                .eq(GroupBuyOrder::getLeaderUserId, userId)
                .in(GroupBuyOrder::getStatus, 
                    GroupBuyConstants.ORDER_STATUS_PENDING, 
                    GroupBuyConstants.ORDER_STATUS_SUCCESS);
        
        if (this.count(orderWrapper) > 0) {
            return true;
        }
        
        List<Long> orderIds = this.list(new LambdaQueryWrapper<GroupBuyOrder>()
                .eq(GroupBuyOrder::getActivityId, activityId)
                .in(GroupBuyOrder::getStatus, 
                    GroupBuyConstants.ORDER_STATUS_PENDING, 
                    GroupBuyConstants.ORDER_STATUS_SUCCESS))
                .stream()
                .map(GroupBuyOrder::getId)
                .collect(java.util.stream.Collectors.toList());
        
        if (!orderIds.isEmpty()) {
            LambdaQueryWrapper<GroupBuyParticipant> participantWrapper = new LambdaQueryWrapper<>();
            participantWrapper.in(GroupBuyParticipant::getGroupOrderId, orderIds)
                    .eq(GroupBuyParticipant::getUserId, userId)
                    .in(GroupBuyParticipant::getPayStatus, 
                        GroupBuyConstants.PAY_STATUS_PENDING, 
                        GroupBuyConstants.PAY_STATUS_PAID);
            
            return participantMapper.selectCount(participantWrapper) > 0;
        }
        
        return false;
    }
    
    private void setUserLimitLock(Long activityId, Long userId, LocalDateTime expireTime) {
        String key = USER_LIMIT_KEY_PREFIX + activityId + ":" + userId;
        long ttlSeconds = java.time.Duration.between(LocalDateTime.now(), expireTime).getSeconds();
        if (ttlSeconds > 0) {
            redisTemplate.opsForValue().set(key, "locked", ttlSeconds, TimeUnit.SECONDS);
        }
    }
    
    private String generateOrderNo() {
        return "GB" + System.currentTimeMillis() + UUID.randomUUID().toString().replace("-", "").substring(0, 6).toUpperCase();
    }
}

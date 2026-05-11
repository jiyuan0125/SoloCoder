package com.example.groupbuy.controller;

import com.example.groupbuy.common.Result;
import com.example.groupbuy.dto.CreateGroupBuyDTO;
import com.example.groupbuy.dto.JoinGroupBuyDTO;
import com.example.groupbuy.entity.GroupBuyOrder;
import com.example.groupbuy.entity.GroupBuyParticipant;
import com.example.groupbuy.service.GroupBuyOrderService;
import lombok.RequiredArgsConstructor;
import org.springframework.web.bind.annotation.*;

import javax.validation.Valid;
import java.util.List;

@RestController
@RequestMapping("/api/group-buy/orders")
@RequiredArgsConstructor
public class GroupBuyOrderController {
    
    private final GroupBuyOrderService orderService;
    
    @PostMapping("/create")
    public Result<GroupBuyOrder> createGroupBuy(@Valid @RequestBody CreateGroupBuyDTO dto) {
        GroupBuyOrder order = orderService.createGroupBuy(dto);
        return Result.success(order);
    }
    
    @PostMapping("/join")
    public Result<GroupBuyParticipant> joinGroupBuy(@Valid @RequestBody JoinGroupBuyDTO dto) {
        GroupBuyParticipant participant = orderService.joinGroupBuy(dto);
        return Result.success(participant);
    }
    
    @GetMapping("/available")
    public Result<List<GroupBuyOrder>> getAvailableOrders(
            @RequestParam Long activityId,
            @RequestParam(defaultValue = "1") int page,
            @RequestParam(defaultValue = "10") int size) {
        List<GroupBuyOrder> orders = orderService.getAvailableOrders(activityId, page, size);
        return Result.success(orders);
    }
    
    @GetMapping("/{orderId}")
    public Result<GroupBuyOrder> getOrderDetail(@PathVariable Long orderId) {
        GroupBuyOrder order = orderService.getOrderDetail(orderId);
        return Result.success(order);
    }
    
    @GetMapping("/{orderId}/participants")
    public Result<List<GroupBuyParticipant>> getOrderParticipants(@PathVariable Long orderId) {
        List<GroupBuyParticipant> participants = orderService.getOrderParticipants(orderId);
        return Result.success(participants);
    }
}

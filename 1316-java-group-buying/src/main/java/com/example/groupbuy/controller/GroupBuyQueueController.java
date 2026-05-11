package com.example.groupbuy.controller;

import com.example.groupbuy.common.Result;
import com.example.groupbuy.entity.GroupBuyQueue;
import com.example.groupbuy.service.GroupBuyQueueService;
import lombok.RequiredArgsConstructor;
import org.springframework.web.bind.annotation.*;

import java.util.List;

@RestController
@RequestMapping("/api/group-buy/queue")
@RequiredArgsConstructor
public class GroupBuyQueueController {
    
    private final GroupBuyQueueService queueService;
    
    @GetMapping("/waiting/{activityId}")
    public Result<List<GroupBuyQueue>> getWaitingQueues(@PathVariable Long activityId) {
        List<GroupBuyQueue> queues = queueService.getWaitingQueues(activityId);
        return Result.success(queues);
    }
    
    @GetMapping("/position/{activityId}/{groupOrderId}")
    public Result<Integer> getQueuePosition(
            @PathVariable Long activityId,
            @PathVariable Long groupOrderId) {
        int position = queueService.getQueuePosition(activityId, groupOrderId);
        return Result.success(position);
    }
    
    @PostMapping("/cancel/{groupOrderId}")
    public Result<Boolean> cancelQueuedOrder(@PathVariable Long groupOrderId) {
        boolean success = queueService.cancelQueuedOrder(groupOrderId);
        return Result.success(success);
    }
}

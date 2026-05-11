package com.example.groupbuy.controller;

import com.baomidou.mybatisplus.core.metadata.IPage;
import com.example.groupbuy.common.Result;
import com.example.groupbuy.dto.GroupBuyActivityDTO;
import com.example.groupbuy.entity.GroupBuyActivity;
import com.example.groupbuy.entity.GroupBuyPriceTier;
import com.example.groupbuy.service.GroupBuyActivityService;
import lombok.RequiredArgsConstructor;
import org.springframework.web.bind.annotation.*;

import javax.validation.Valid;
import java.util.List;

@RestController
@RequestMapping("/api/group-buy/activities")
@RequiredArgsConstructor
public class GroupBuyActivityController {
    
    private final GroupBuyActivityService activityService;
    
    @PostMapping
    public Result<GroupBuyActivity> createActivity(@Valid @RequestBody GroupBuyActivityDTO dto) {
        GroupBuyActivity activity = activityService.createActivity(dto);
        return Result.success(activity);
    }
    
    @PutMapping
    public Result<GroupBuyActivity> updateActivity(@Valid @RequestBody GroupBuyActivityDTO dto) {
        GroupBuyActivity activity = activityService.updateActivity(dto);
        return Result.success(activity);
    }
    
    @GetMapping
    public Result<IPage<GroupBuyActivity>> listActivities(
            @RequestParam(defaultValue = "1") int page,
            @RequestParam(defaultValue = "10") int size,
            @RequestParam(required = false) Integer status) {
        IPage<GroupBuyActivity> activities = activityService.listActivities(page, size, status);
        return Result.success(activities);
    }
    
    @GetMapping("/{activityId}")
    public Result<GroupBuyActivity> getActivityDetail(@PathVariable Long activityId) {
        GroupBuyActivity activity = activityService.getActivityDetail(activityId);
        return Result.success(activity);
    }
    
    @GetMapping("/{activityId}/price-tiers")
    public Result<List<GroupBuyPriceTier>> getPriceTiers(@PathVariable Long activityId) {
        List<GroupBuyPriceTier> priceTiers = activityService.getPriceTiers(activityId);
        return Result.success(priceTiers);
    }
}

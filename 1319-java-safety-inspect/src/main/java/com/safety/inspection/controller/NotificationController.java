package com.safety.inspection.controller;

import com.baomidou.mybatisplus.extension.plugins.pagination.Page;
import com.safety.inspection.common.Result;
import com.safety.inspection.entity.Hazard;
import com.safety.inspection.entity.SystemNotification;
import com.safety.inspection.service.NotificationService;
import com.safety.inspection.service.RepeatHazardService;
import lombok.RequiredArgsConstructor;
import org.springframework.web.bind.annotation.*;

import java.util.List;

@RestController
@RequestMapping("/api/notifications")
@RequiredArgsConstructor
public class NotificationController {

    private final NotificationService notificationService;

    @GetMapping("/page")
    public Result<Page<SystemNotification>> getMyNotifications(
            @RequestParam(defaultValue = "1") int pageNum,
            @RequestParam(defaultValue = "10") int pageSize,
            @RequestParam(required = false) Boolean isRead) {
        return Result.success(notificationService.getMyNotifications(pageNum, pageSize, isRead));
    }

    @GetMapping("/unread-count")
    public Result<Long> getUnreadCount() {
        return Result.success(notificationService.getUnreadCount());
    }

    @PutMapping("/{id}/read")
    public Result<Void> markAsRead(@PathVariable Long id) {
        notificationService.markAsRead(id);
        return Result.success();
    }

    @PutMapping("/read-all")
    public Result<Void> markAllAsRead() {
        notificationService.markAllAsRead();
        return Result.success();
    }
}

@RestController
@RequestMapping("/api/repeat-hazards")
@RequiredArgsConstructor
class RepeatHazardController {

    private final RepeatHazardService repeatHazardService;

    @GetMapping("/all")
    public Result<List<Hazard>> getAllRepeatHazards() {
        return Result.success(repeatHazardService.getAllRepeatHazards());
    }

    @GetMapping("/{hazardId}/related")
    public Result<List<Hazard>> getRelatedHazards(@PathVariable Long hazardId) {
        return Result.success(repeatHazardService.getRelatedHazards(hazardId));
    }
}

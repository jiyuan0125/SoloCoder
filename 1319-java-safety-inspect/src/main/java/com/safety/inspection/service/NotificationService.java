package com.safety.inspection.service;

import com.baomidou.mybatisplus.core.conditions.query.LambdaQueryWrapper;
import com.baomidou.mybatisplus.core.conditions.update.LambdaUpdateWrapper;
import com.baomidou.mybatisplus.extension.plugins.pagination.Page;
import com.baomidou.mybatisplus.extension.service.impl.ServiceImpl;
import com.safety.inspection.entity.Hazard;
import com.safety.inspection.entity.InspectionTask;
import com.safety.inspection.entity.SystemNotification;
import com.safety.inspection.enums.HazardLevelEnum;
import com.safety.inspection.enums.NotificationTypeEnum;
import com.safety.inspection.mapper.SystemNotificationMapper;
import com.safety.inspection.security.CurrentUser;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.stereotype.Service;

import java.time.LocalDateTime;

@Slf4j
@Service
@RequiredArgsConstructor
public class NotificationService extends ServiceImpl<SystemNotificationMapper, SystemNotification> {

    private final SystemNotificationMapper notificationMapper;
    private final CurrentUser currentUser;

    public Page<SystemNotification> getMyNotifications(int pageNum, int pageSize, Boolean isRead) {
        Long userId = currentUser.getCurrentUserId();
        Page<SystemNotification> page = new Page<>(pageNum, pageSize);
        LambdaQueryWrapper<SystemNotification> wrapper = new LambdaQueryWrapper<>();
        wrapper.eq(SystemNotification::getRecipientId, userId);
        if (isRead != null) {
            wrapper.eq(SystemNotification::getIsRead, isRead ? 1 : 0);
        }
        wrapper.orderByDesc(SystemNotification::getCreatedAt);
        return notificationMapper.selectPage(page, wrapper);
    }

    public void markAsRead(Long id) {
        SystemNotification notification = notificationMapper.selectById(id);
        if (notification != null) {
            notification.setIsRead(1);
            notification.setReadTime(LocalDateTime.now());
            notificationMapper.updateById(notification);
        }
    }

    public void markAllAsRead() {
        Long userId = currentUser.getCurrentUserId();
        notificationMapper.update(
            null,
            new LambdaUpdateWrapper<SystemNotification>()
                .eq(SystemNotification::getRecipientId, userId)
                .eq(SystemNotification::getIsRead, 0)
                .set(SystemNotification::getIsRead, 1)
                .set(SystemNotification::getReadTime, LocalDateTime.now())
        );
    }

    public long getUnreadCount() {
        Long userId = currentUser.getCurrentUserId();
        return notificationMapper.selectCount(
            new LambdaQueryWrapper<SystemNotification>()
                .eq(SystemNotification::getRecipientId, userId)
                .eq(SystemNotification::getIsRead, 0)
        );
    }

    public void createHazardCreatedNotification(Hazard hazard) {
        String title = "新隐患待处理";
        String content = String.format("发现新的%s：%s，请及时安排整改。",
            HazardLevelEnum.getByCode(hazard.getHazardLevel()).getDesc(),
            hazard.getDescription());

        createNotification(NotificationTypeEnum.HAZARD_CREATED, title, content,
            hazard.getResponsiblePersonId(), "HAZARD", hazard.getId());
    }

    public void createHazardAssignedNotification(Hazard hazard) {
        String title = "隐患整改任务分配";
        String content = String.format("您被分配为隐患整改责任人：%s，请在%s前完成整改。",
            hazard.getDescription(),
            hazard.getRectificationDeadline());

        createNotification(NotificationTypeEnum.HAZARD_ASSIGNED, title, content,
            hazard.getResponsiblePersonId(), "HAZARD", hazard.getId());
    }

    public void createHazardRectifiedNotification(Hazard hazard) {
        String title = "隐患整改完成待复检";
        String content = String.format("隐患已完成整改，等待复检：%s", hazard.getDescription());

        createNotification(NotificationTypeEnum.HAZARD_RECHECK_NEEDED, title, content,
            null, "HAZARD", hazard.getId());
    }

    public void createHazardRecheckPassedNotification(Hazard hazard) {
        String title = "隐患复检通过";
        String content = String.format("隐患复检通过，已关闭：%s", hazard.getDescription());

        createNotification(NotificationTypeEnum.HAZARD_RECHECK_PASSED, title, content,
            hazard.getResponsiblePersonId(), "HAZARD", hazard.getId());
    }

    public void createHazardRecheckFailedNotification(Hazard hazard) {
        String title = "隐患复检不通过";
        String content = String.format("隐患复检不通过，需重新整改：%s", hazard.getDescription());

        createNotification(NotificationTypeEnum.HAZARD_RECHECK_FAILED, title, content,
            hazard.getResponsiblePersonId(), "HAZARD", hazard.getId());
    }

    public void createHazardUpgradedNotification(Hazard hazard, String oldLevel) {
        String title = "隐患等级升级";
        String content = String.format("隐患等级已从%s升级为%s：%s",
            HazardLevelEnum.getByCode(oldLevel).getDesc(),
            HazardLevelEnum.getByCode(hazard.getHazardLevel()).getDesc(),
            hazard.getDescription());

        createNotification(NotificationTypeEnum.HAZARD_UPGRADED, title, content,
            hazard.getResponsiblePersonId(), "HAZARD", hazard.getId());
    }

    public void createHazardOverdueNotification(Hazard hazard) {
        String title = "隐患整改超期";
        String content = String.format("隐患整改已超期：%s，请立即处理！", hazard.getDescription());

        createNotification(NotificationTypeEnum.HAZARD_OVERDUE, title, content,
            hazard.getResponsiblePersonId(), "HAZARD", hazard.getId());
    }

    public void createTaskAssignedNotification(InspectionTask task) {
        String title = "新巡检任务分配";
        String content = String.format("您有新的巡检任务需要执行，任务日期：%s", task.getTaskDate());

        createNotification(NotificationTypeEnum.TASK_ASSIGNED, title, content,
            task.getInspectorId(), "TASK", task.getId());
    }

    public void createMissedTaskNotification(InspectionTask task) {
        String title = "漏检提醒";
        String content = String.format("您有巡检任务未按时完成，已记录为漏检：%s", task.getTaskNo());

        createNotification(NotificationTypeEnum.TASK_MISSED, title, content,
            task.getInspectorId(), "TASK", task.getId());
    }

    private void createNotification(NotificationTypeEnum type, String title, String content,
                                     Long recipientId, String relatedType, Long relatedId) {
        if (recipientId == null) {
            log.warn("通知接收人为空，跳过通知创建");
            return;
        }

        SystemNotification notification = new SystemNotification();
        notification.setNotificationType(type.getCode());
        notification.setTitle(title);
        notification.setContent(content);
        notification.setRecipientId(recipientId);
        notification.setRelatedType(relatedType);
        notification.setRelatedId(relatedId);
        notification.setIsRead(0);
        notification.setCreatedAt(LocalDateTime.now());
        notificationMapper.insert(notification);

        log.info("创建通知：类型={}, 接收人={}, 标题={}", type.getCode(), recipientId, title);
    }
}

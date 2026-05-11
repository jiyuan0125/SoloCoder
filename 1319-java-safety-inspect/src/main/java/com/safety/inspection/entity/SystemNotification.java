package com.safety.inspection.entity;

import com.baomidou.mybatisplus.annotation.TableName;
import lombok.Data;

import java.time.LocalDateTime;

@Data
@TableName("system_notification")
public class SystemNotification {

    private Long id;

    private String notificationType;

    private String title;

    private String content;

    private Long recipientId;

    private String relatedType;

    private Long relatedId;

    private Integer isRead;

    private LocalDateTime readTime;

    private LocalDateTime createdAt;
}

package com.safety.inspection.entity;

import com.baomidou.mybatisplus.annotation.TableName;
import lombok.Data;

@Data
@TableName("sys_user_role")
public class UserRole {

    private Long id;

    private Long userId;

    private Long roleId;

    private java.time.LocalDateTime createdAt;
}

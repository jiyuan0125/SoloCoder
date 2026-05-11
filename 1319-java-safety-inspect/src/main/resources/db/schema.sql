CREATE DATABASE IF NOT EXISTS safety_inspection DEFAULT CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;

USE safety_inspection;

CREATE TABLE IF NOT EXISTS sys_user (
    id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '用户ID',
    username VARCHAR(50) NOT NULL UNIQUE COMMENT '用户名',
    password VARCHAR(100) NOT NULL COMMENT '密码',
    real_name VARCHAR(50) NOT NULL COMMENT '真实姓名',
    phone VARCHAR(20) COMMENT '手机号',
    email VARCHAR(100) COMMENT '邮箱',
    department_id BIGINT COMMENT '部门ID',
    status TINYINT DEFAULT 1 COMMENT '状态 1正常 0禁用',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    deleted TINYINT DEFAULT 0 COMMENT '逻辑删除 0否 1是',
    INDEX idx_username (username),
    INDEX idx_department (department_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户表';

CREATE TABLE IF NOT EXISTS sys_role (
    id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '角色ID',
    role_code VARCHAR(50) NOT NULL UNIQUE COMMENT '角色编码',
    role_name VARCHAR(50) NOT NULL COMMENT '角色名称',
    description VARCHAR(200) COMMENT '描述',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    deleted TINYINT DEFAULT 0 COMMENT '逻辑删除 0否 1是',
    INDEX idx_role_code (role_code)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='角色表';

CREATE TABLE IF NOT EXISTS sys_user_role (
    id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    user_id BIGINT NOT NULL COMMENT '用户ID',
    role_id BIGINT NOT NULL COMMENT '角色ID',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    UNIQUE KEY uk_user_role (user_id, role_id),
    INDEX idx_user (user_id),
    INDEX idx_role (role_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='用户角色关联表';

CREATE TABLE IF NOT EXISTS sys_department (
    id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '部门ID',
    dept_name VARCHAR(50) NOT NULL COMMENT '部门名称',
    parent_id BIGINT DEFAULT 0 COMMENT '父部门ID',
    leader_id BIGINT COMMENT '负责人ID',
    sort INT DEFAULT 0 COMMENT '排序',
    status TINYINT DEFAULT 1 COMMENT '状态 1正常 0禁用',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    deleted TINYINT DEFAULT 0 COMMENT '逻辑删除 0否 1是',
    INDEX idx_parent (parent_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='部门表';

CREATE TABLE IF NOT EXISTS inspection_area (
    id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '区域ID',
    area_name VARCHAR(100) NOT NULL COMMENT '区域名称',
    area_type VARCHAR(50) COMMENT '区域类型：车间、仓库、办公楼、食堂等',
    location VARCHAR(200) COMMENT '具体位置',
    description VARCHAR(500) COMMENT '区域描述',
    manager_id BIGINT COMMENT '区域负责人ID',
    department_id BIGINT COMMENT '所属部门ID',
    status TINYINT DEFAULT 1 COMMENT '状态 1正常 0禁用',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    deleted TINYINT DEFAULT 0 COMMENT '逻辑删除 0否 1是',
    INDEX idx_department (department_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='巡检区域表';

CREATE TABLE IF NOT EXISTS inspection_check_item (
    id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '检查项ID',
    area_id BIGINT NOT NULL COMMENT '关联区域ID',
    item_name VARCHAR(200) NOT NULL COMMENT '检查项名称',
    item_description VARCHAR(500) COMMENT '检查项描述',
    standard VARCHAR(500) COMMENT '检查标准',
    risk_level VARCHAR(20) DEFAULT 'LOW' COMMENT '风险等级：LOW一般 MEDIUM较高 HIGH高',
    sort INT DEFAULT 0 COMMENT '排序',
    status TINYINT DEFAULT 1 COMMENT '状态 1正常 0禁用',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    deleted TINYINT DEFAULT 0 COMMENT '逻辑删除 0否 1是',
    INDEX idx_area (area_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='检查项表';

CREATE TABLE IF NOT EXISTS inspection_plan (
    id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '计划ID',
    plan_name VARCHAR(100) NOT NULL COMMENT '计划名称',
    plan_type VARCHAR(20) DEFAULT 'REGULAR' COMMENT '计划类型：REGULAR常规 TEMPORARY临时',
    area_id BIGINT NOT NULL COMMENT '巡检区域ID',
    frequency VARCHAR(20) NOT NULL COMMENT '巡检频率：DAILY每日 WEEKLY每周 MONTHLY每月 CUSTOM自定义',
    frequency_days INT COMMENT '自定义频率（天数）',
    start_date DATE NOT NULL COMMENT '开始日期',
    end_date DATE COMMENT '结束日期',
    inspector_id BIGINT NOT NULL COMMENT '巡检员ID',
    description VARCHAR(500) COMMENT '计划描述',
    status TINYINT DEFAULT 1 COMMENT '状态 1启用 0停用',
    created_by BIGINT COMMENT '创建人ID',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    deleted TINYINT DEFAULT 0 COMMENT '逻辑删除 0否 1是',
    INDEX idx_area (area_id),
    INDEX idx_inspector (inspector_id),
    INDEX idx_status (status)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='巡检计划表';

CREATE TABLE IF NOT EXISTS inspection_task (
    id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '任务ID',
    task_no VARCHAR(50) NOT NULL UNIQUE COMMENT '任务编号',
    plan_id BIGINT COMMENT '关联计划ID',
    area_id BIGINT NOT NULL COMMENT '巡检区域ID',
    inspector_id BIGINT NOT NULL COMMENT '巡检员ID',
    task_date DATE NOT NULL COMMENT '任务日期',
    start_time DATETIME COMMENT '开始时间',
    end_time DATETIME COMMENT '结束时间',
    task_status VARCHAR(20) DEFAULT 'PENDING' COMMENT '任务状态：PENDING待执行 IN_PROGRESS进行中 COMPLETED已完成 MISSED漏检',
    remark VARCHAR(500) COMMENT '备注',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    deleted TINYINT DEFAULT 0 COMMENT '逻辑删除 0否 1是',
    INDEX idx_plan (plan_id),
    INDEX idx_inspector (inspector_id),
    INDEX idx_status (task_status),
    INDEX idx_task_date (task_date)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='巡检任务表';

CREATE TABLE IF NOT EXISTS inspection_record (
    id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '记录ID',
    task_id BIGINT NOT NULL COMMENT '任务ID',
    check_item_id BIGINT NOT NULL COMMENT '检查项ID',
    check_result VARCHAR(20) NOT NULL COMMENT '检查结果：NORMAL正常 ABNORMAL异常',
    description VARCHAR(500) COMMENT '问题描述',
    photos VARCHAR(1000) COMMENT '照片URL（多个用逗号分隔）',
    created_by BIGINT COMMENT '巡检员ID',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '检查时间',
    deleted TINYINT DEFAULT 0 COMMENT '逻辑删除 0否 1是',
    INDEX idx_task (task_id),
    INDEX idx_check_item (check_item_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='巡检记录表';

CREATE TABLE IF NOT EXISTS hazard (
    id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '隐患ID',
    hazard_no VARCHAR(50) NOT NULL UNIQUE COMMENT '隐患编号',
    inspection_record_id BIGINT COMMENT '关联巡检记录ID',
    task_id BIGINT COMMENT '任务ID',
    area_id BIGINT NOT NULL COMMENT '区域ID',
    hazard_level VARCHAR(20) NOT NULL COMMENT '隐患等级：GENERAL一般 LARGER较大 MAJOR重大',
    original_level VARCHAR(20) COMMENT '原始等级',
    upgrade_count INT DEFAULT 0 COMMENT '升级次数',
    hazard_type VARCHAR(100) COMMENT '隐患类型',
    location VARCHAR(200) COMMENT '具体位置',
    description VARCHAR(1000) NOT NULL COMMENT '隐患描述',
    photos VARCHAR(1000) COMMENT '隐患照片',
    status VARCHAR(20) DEFAULT 'PENDING_RECTIFICATION' COMMENT '状态：PENDING_RECTIFICATION待整改 RECTIFYING整改中 PENDING_RECHECK待复检 RECHECKED已复检 CLOSED已关闭',
    rectification_deadline DATETIME COMMENT '整改期限',
    rectification_plan VARCHAR(1000) COMMENT '整改方案',
    responsible_person_id BIGINT COMMENT '整改责任人ID',
    rectification_result VARCHAR(1000) COMMENT '整改结果',
    rectification_photos VARCHAR(1000) COMMENT '整改后照片',
    rectification_time DATETIME COMMENT '整改完成时间',
    rechecker_id BIGINT COMMENT '复检人ID',
    recheck_result VARCHAR(20) COMMENT '复检结果：PASS通过 FAIL未通过',
    recheck_time DATETIME COMMENT '复检时间',
    recheck_remark VARCHAR(500) COMMENT '复检备注',
    is_repeat TINYINT DEFAULT 0 COMMENT '是否重复隐患 0否 1是',
    related_hazard_id BIGINT COMMENT '关联重复隐患ID',
    created_by BIGINT COMMENT '发现人ID',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    closed_at DATETIME COMMENT '关闭时间',
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
    deleted TINYINT DEFAULT 0 COMMENT '逻辑删除 0否 1是',
    INDEX idx_task (task_id),
    INDEX idx_area (area_id),
    INDEX idx_status (status),
    INDEX idx_level (hazard_level),
    INDEX idx_responsible (responsible_person_id),
    INDEX idx_created_at (created_at)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='隐患表';

CREATE TABLE IF NOT EXISTS hazard_upgrade_record (
    id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    hazard_id BIGINT NOT NULL COMMENT '隐患ID',
    old_level VARCHAR(20) NOT NULL COMMENT '原等级',
    new_level VARCHAR(20) NOT NULL COMMENT '新等级',
    reason VARCHAR(500) COMMENT '升级原因',
    new_deadline DATETIME COMMENT '新整改期限',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '升级时间',
    INDEX idx_hazard (hazard_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='隐患升级记录表';

CREATE TABLE IF NOT EXISTS system_notification (
    id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '通知ID',
    notification_type VARCHAR(50) NOT NULL COMMENT '通知类型',
    title VARCHAR(200) NOT NULL COMMENT '通知标题',
    content VARCHAR(2000) NOT NULL COMMENT '通知内容',
    recipient_id BIGINT NOT NULL COMMENT '接收人ID',
    related_type VARCHAR(50) COMMENT '关联类型：HAZARD隐患 TASK任务',
    related_id BIGINT COMMENT '关联ID',
    is_read TINYINT DEFAULT 0 COMMENT '是否已读 0否 1是',
    read_time DATETIME COMMENT '阅读时间',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    INDEX idx_recipient (recipient_id),
    INDEX idx_is_read (is_read),
    INDEX idx_related (related_type, related_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='系统通知表';

CREATE TABLE IF NOT EXISTS monthly_report (
    id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT '报告ID',
    report_year INT NOT NULL COMMENT '年份',
    report_month INT NOT NULL COMMENT '月份',
    total_hazards INT DEFAULT 0 COMMENT '新增隐患数',
    rectified_count INT DEFAULT 0 COMMENT '已整改数',
    rectifying_count INT DEFAULT 0 COMMENT '整改中数量',
    overdue_count INT DEFAULT 0 COMMENT '超期数量',
    general_count INT DEFAULT 0 COMMENT '一般隐患数',
    larger_count INT DEFAULT 0 COMMENT '较大隐患数',
    major_count INT DEFAULT 0 COMMENT '重大隐患数',
    generated_at DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '生成时间',
    UNIQUE KEY uk_year_month (report_year, report_month)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='月度报告表';

CREATE TABLE IF NOT EXISTS department_hazard_stat (
    id BIGINT PRIMARY KEY AUTO_INCREMENT COMMENT 'ID',
    report_id BIGINT NOT NULL COMMENT '报告ID',
    department_id BIGINT NOT NULL COMMENT '部门ID',
    total_count INT DEFAULT 0 COMMENT '隐患总数',
    rectified_count INT DEFAULT 0 COMMENT '已整改数',
    rectification_rate DECIMAL(5,2) DEFAULT 0 COMMENT '整改率(%)',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
    INDEX idx_report (report_id),
    INDEX idx_department (department_id)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='部门隐患统计表';

INSERT INTO sys_role (role_code, role_name, description) VALUES
('ADMIN', '安全管理员', '系统管理员，负责巡检计划制定、系统管理'),
('INSPECTOR', '巡检员', '执行巡检任务的人员'),
('RESPONSIBLE', '整改责任人', '负责隐患整改的人员'),
('DEPT_LEADER', '部门负责人', '部门主管，接收隐患通知'),
('SAFETY_DIRECTOR', '安全总监', '安全管理负责人，接收升级隐患通知'),
('FACTORY_MANAGER', '厂长', '工厂最高负责人，接收重大隐患通知');

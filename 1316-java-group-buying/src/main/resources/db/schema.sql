-- 拼团活动表
CREATE TABLE IF NOT EXISTS `group_buy_activity` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `product_id` BIGINT NOT NULL COMMENT '商品ID',
  `product_name` VARCHAR(200) NOT NULL COMMENT '商品名称',
  `original_price` DECIMAL(10,2) NOT NULL COMMENT '原价',
  `product_type` TINYINT NOT NULL DEFAULT 1 COMMENT '商品类型：1-实物商品，2-虚拟商品',
  `max_group_count` INT NOT NULL COMMENT '最大团数',
  `used_group_count` INT NOT NULL DEFAULT 0 COMMENT '已使用团数',
  `start_time` DATETIME NOT NULL COMMENT '活动开始时间',
  `end_time` DATETIME NOT NULL COMMENT '活动结束时间',
  `status` TINYINT NOT NULL DEFAULT 0 COMMENT '状态：0-未开始，1-进行中，2-已结束',
  `create_time` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_time` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  KEY `idx_product_id` (`product_id`),
  KEY `idx_status` (`status`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='拼团活动表';

-- 拼团阶梯价格表
CREATE TABLE IF NOT EXISTS `group_buy_price_tier` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `activity_id` BIGINT NOT NULL COMMENT '拼团活动ID',
  `min_people_count` INT NOT NULL COMMENT '最低成团人数',
  `target_people_count` INT NOT NULL COMMENT '目标人数档位',
  `group_price` DECIMAL(10,2) NOT NULL COMMENT '团购价',
  `discount_rate` DECIMAL(5,2) DEFAULT NULL COMMENT '折扣率',
  `sort_order` INT NOT NULL DEFAULT 0 COMMENT '排序',
  `create_time` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  PRIMARY KEY (`id`),
  KEY `idx_activity_id` (`activity_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='拼团阶梯价格表';

-- 拼团订单表
CREATE TABLE IF NOT EXISTS `group_buy_order` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `order_no` VARCHAR(32) NOT NULL COMMENT '拼团订单号',
  `activity_id` BIGINT NOT NULL COMMENT '拼团活动ID',
  `product_id` BIGINT NOT NULL COMMENT '商品ID',
  `leader_user_id` BIGINT NOT NULL COMMENT '团长用户ID',
  `target_people_count` INT NOT NULL COMMENT '目标成团人数',
  `actual_people_count` INT NOT NULL DEFAULT 0 COMMENT '实际成团人数',
  `group_price` DECIMAL(10,2) DEFAULT NULL COMMENT '实际团购价',
  `validity_hours` INT NOT NULL COMMENT '拼团有效期（小时）',
  `start_time` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '拼团开始时间',
  `end_time` DATETIME NOT NULL COMMENT '拼团结束时间',
  `status` TINYINT NOT NULL DEFAULT 0 COMMENT '状态：0-待成团，1-已成团，2-已取消，3-已退款',
  `queue_order` INT DEFAULT NULL COMMENT '排队顺序（库存不足时）',
  `is_queued` TINYINT NOT NULL DEFAULT 0 COMMENT '是否在排队：0-否，1-是',
  `create_time` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_time` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_order_no` (`order_no`),
  KEY `idx_activity_id` (`activity_id`),
  KEY `idx_leader_user_id` (`leader_user_id`),
  KEY `idx_status` (`status`),
  KEY `idx_end_time` (`end_time`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='拼团订单表';

-- 拼团参与者表
CREATE TABLE IF NOT EXISTS `group_buy_participant` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `group_order_id` BIGINT NOT NULL COMMENT '拼团订单ID',
  `user_id` BIGINT NOT NULL COMMENT '用户ID',
  `is_leader` TINYINT NOT NULL DEFAULT 0 COMMENT '是否团长：0-否，1-是',
  `join_time` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '加入时间',
  `pay_time` DATETIME DEFAULT NULL COMMENT '支付时间',
  `pay_status` TINYINT NOT NULL DEFAULT 0 COMMENT '支付状态：0-待支付，1-已支付，2-已退款',
  `pay_amount` DECIMAL(10,2) DEFAULT NULL COMMENT '支付金额',
  `refund_amount` DECIMAL(10,2) DEFAULT NULL COMMENT '退款金额',
  `refund_time` DATETIME DEFAULT NULL COMMENT '退款时间',
  `product_delivered` TINYINT NOT NULL DEFAULT 0 COMMENT '商品是否已发放/发货：0-否，1-是',
  `create_time` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  `update_time` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',
  PRIMARY KEY (`id`),
  UNIQUE KEY `uk_group_order_user` (`group_order_id`, `user_id`),
  KEY `idx_user_id` (`user_id`),
  KEY `idx_group_order_id` (`group_order_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='拼团参与者表';

-- 库存排队表
CREATE TABLE IF NOT EXISTS `group_buy_queue` (
  `id` BIGINT NOT NULL AUTO_INCREMENT COMMENT '主键ID',
  `group_order_id` BIGINT NOT NULL COMMENT '拼团订单ID',
  `activity_id` BIGINT NOT NULL COMMENT '拼团活动ID',
  `user_id` BIGINT NOT NULL COMMENT '用户ID',
  `queue_status` TINYINT NOT NULL DEFAULT 0 COMMENT '排队状态：0-排队中，1-已处理，2-已取消',
  `queue_time` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '排队时间',
  `process_time` DATETIME DEFAULT NULL COMMENT '处理时间',
  `create_time` DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',
  PRIMARY KEY (`id`),
  KEY `idx_activity_id` (`activity_id`),
  KEY `idx_group_order_id` (`group_order_id`)
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COMMENT='库存排队表';

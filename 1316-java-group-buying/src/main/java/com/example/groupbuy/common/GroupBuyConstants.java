package com.example.groupbuy.common;

public class GroupBuyConstants {
    
    public static final int ACTIVITY_STATUS_NOT_STARTED = 0;
    public static final int ACTIVITY_STATUS_ONGOING = 1;
    public static final int ACTIVITY_STATUS_ENDED = 2;
    
    public static final int ORDER_STATUS_PENDING = 0;
    public static final int ORDER_STATUS_SUCCESS = 1;
    public static final int ORDER_STATUS_CANCELLED = 2;
    public static final int ORDER_STATUS_REFUNDED = 3;
    
    public static final int PAY_STATUS_PENDING = 0;
    public static final int PAY_STATUS_PAID = 1;
    public static final int PAY_STATUS_REFUNDED = 2;
    
    public static final int PRODUCT_TYPE_PHYSICAL = 1;
    public static final int PRODUCT_TYPE_VIRTUAL = 2;
    
    public static final int QUEUE_STATUS_WAITING = 0;
    public static final int QUEUE_STATUS_PROCESSED = 1;
    public static final int QUEUE_STATUS_CANCELLED = 2;
    
    public static final int IS_LEADER_YES = 1;
    public static final int IS_LEADER_NO = 0;
    
    public static final int PRODUCT_DELIVERED_YES = 1;
    public static final int PRODUCT_DELIVERED_NO = 0;
    
    public static final int IS_QUEUED_YES = 1;
    public static final int IS_QUEUED_NO = 0;
}

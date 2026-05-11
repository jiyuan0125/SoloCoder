package com.safety.inspection.dto;

import lombok.Data;

@Data
public class DashboardStatsVO {
    private Integer pendingRectification;
    private Integer rectifying;
    private Integer pendingRecheck;
    private Integer overdue;
    private Integer todayTasks;
    private Integer completedTasks;
    private Integer missedTasks;
}

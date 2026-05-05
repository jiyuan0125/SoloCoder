package com.orgchart.server.util;

import java.util.concurrent.atomic.AtomicLong;

public class IdGenerator {

    private static final AtomicLong DEPARTMENT_ID = new AtomicLong(1);
    private static final AtomicLong EMPLOYEE_ID = new AtomicLong(1);
    private static final AtomicLong VIRTUAL_TEAM_ID = new AtomicLong(1);
    private static final AtomicLong TRANSFER_HISTORY_ID = new AtomicLong(1);
    private static final AtomicLong OPERATION_LOG_ID = new AtomicLong(1);

    private IdGenerator() {
    }

    public static String generateDepartmentId() {
        return "DEP" + String.format("%06d", DEPARTMENT_ID.getAndIncrement());
    }

    public static String generateEmployeeId() {
        return "EMP" + String.format("%06d", EMPLOYEE_ID.getAndIncrement());
    }

    public static String generateVirtualTeamId() {
        return "VT" + String.format("%06d", VIRTUAL_TEAM_ID.getAndIncrement());
    }

    public static String generateTransferHistoryId() {
        return "TH" + String.format("%06d", TRANSFER_HISTORY_ID.getAndIncrement());
    }

    public static String generateOperationLogId() {
        return "OL" + String.format("%06d", OPERATION_LOG_ID.getAndIncrement());
    }
}

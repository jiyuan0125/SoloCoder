package com.reimbursement.common.constant;

public class ErrorCode {
    public static final int SUCCESS = 200;
    public static final int BAD_REQUEST = 400;
    public static final int NOT_FOUND = 404;
    public static final int INTERNAL_ERROR = 500;

    public static final int INVALID_PERCENTAGE_SUM = 1001;
    public static final int INVALID_PERCENTAGE_RANGE = 1002;
    public static final int TOO_MANY_COST_CENTERS = 1003;
    public static final int REIMBURSEMENT_NOT_FOUND = 1004;
    public static final int INVALID_STATUS_FOR_OPERATION = 1005;
    public static final int COST_CENTER_NOT_FOUND = 1006;
    public static final int ALREADY_APPROVED = 1007;
    public static final int CANNOT_MODIFY_WHILE_APPROVING = 1008;
    public static final int SPECIAL_APPROVAL_REQUIRED = 1009;
}

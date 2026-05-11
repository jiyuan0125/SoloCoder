package com.factory.workorder.exception;

public class WorkorderNotFoundException extends BusinessException {

    public WorkorderNotFoundException(Long workorderId) {
        super("工单不存在，ID: " + workorderId);
    }

    public WorkorderNotFoundException(String orderNo) {
        super("工单不存在，工单编号: " + orderNo);
    }
}

package com.factory.workorder.enums;

public enum OperationType {
    CREATE("创建工单"),
    ASSIGN("分派工单"),
    SUBMIT_ACCEPTANCE("提交验收"),
    ACCEPT("验收通过"),
    REJECT("验收驳回"),
    UPGRADE_PRIORITY("优先级升级");

    private final String description;

    OperationType(String description) {
        this.description = description;
    }

    public String getDescription() {
        return description;
    }
}

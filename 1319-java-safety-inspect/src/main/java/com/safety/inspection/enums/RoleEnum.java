package com.safety.inspection.enums;

import lombok.Getter;

@Getter
public enum RoleEnum {
    ADMIN("ADMIN", "安全管理员"),
    INSPECTOR("INSPECTOR", "巡检员"),
    RESPONSIBLE("RESPONSIBLE", "整改责任人"),
    DEPT_LEADER("DEPT_LEADER", "部门负责人"),
    SAFETY_DIRECTOR("SAFETY_DIRECTOR", "安全总监"),
    FACTORY_MANAGER("FACTORY_MANAGER", "厂长");

    private final String code;
    private final String desc;

    RoleEnum(String code, String desc) {
        this.code = code;
        this.desc = desc;
    }
}

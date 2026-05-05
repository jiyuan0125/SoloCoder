package com.orgchart.common.enums;

public enum ErrorCode {

    SUCCESS(0, "操作成功"),
    PARAM_ERROR(1001, "参数错误"),
    DEPARTMENT_NOT_FOUND(2001, "部门不存在"),
    DEPARTMENT_NAME_DUPLICATE(2002, "同级部门名称重复"),
    DEPARTMENT_HAS_EMPLOYEES(2003, "部门下存在员工，无法删除"),
    PARENT_DEPARTMENT_NOT_FOUND(2004, "父部门不存在"),
    CIRCULAR_REFERENCE(2005, "存在循环引用"),
    DEPARTMENT_CANNOT_BE_OWN_PARENT(2006, "部门不能作为自己的父部门"),
    EMPLOYEE_NOT_FOUND(3001, "员工不存在"),
    EMPLOYEE_ALREADY_EXISTS(3002, "员工已存在"),
    EMPLOYEE_DEPARTMENT_NOT_FOUND(3003, "员工所属部门不存在"),
    VIRTUAL_TEAM_NOT_FOUND(4001, "虚拟团队不存在"),
    INTERNAL_ERROR(9999, "系统内部错误");

    private final int code;
    private final String message;

    ErrorCode(int code, String message) {
        this.code = code;
        this.message = message;
    }

    public int getCode() {
        return code;
    }

    public String getMessage() {
        return message;
    }
}

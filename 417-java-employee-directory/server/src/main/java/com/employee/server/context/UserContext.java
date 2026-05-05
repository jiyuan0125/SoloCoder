package com.employee.server.context;

import com.employee.common.constant.EmployeeConstants;

public class UserContext {
    private static final ThreadLocal<String> currentUserId = new ThreadLocal<>();
    private static final ThreadLocal<String> currentUserName = new ThreadLocal<>();
    private static final ThreadLocal<String> currentRole = new ThreadLocal<>();

    public static void setCurrentUser(String userId, String userName, String role) {
        currentUserId.set(userId);
        currentUserName.set(userName);
        currentRole.set(role);
    }

    public static String getCurrentUserId() {
        return currentUserId.get();
    }

    public static String getCurrentUserName() {
        return currentUserName.get();
    }

    public static String getCurrentRole() {
        return currentRole.get();
    }

    public static void clear() {
        currentUserId.remove();
        currentUserName.remove();
        currentRole.remove();
    }

    public static boolean isHr() {
        return EmployeeConstants.ROLE_HR.equals(getCurrentRole());
    }

    public static boolean isAdmin() {
        return EmployeeConstants.ROLE_ADMIN.equals(getCurrentRole());
    }

    public static boolean isEmployee() {
        return EmployeeConstants.ROLE_EMPLOYEE.equals(getCurrentRole());
    }

    public static boolean canModifyAllFields() {
        return isHr() || isAdmin();
    }

    public static boolean canViewResigned() {
        return isAdmin() || isHr();
    }
}

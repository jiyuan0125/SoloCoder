package com.employee.server.interceptor;

import com.employee.common.constant.EmployeeConstants;
import com.employee.server.context.UserContext;
import jakarta.servlet.http.HttpServletRequest;
import jakarta.servlet.http.HttpServletResponse;
import org.springframework.stereotype.Component;
import org.springframework.web.servlet.HandlerInterceptor;

@Component
public class AuthInterceptor implements HandlerInterceptor {

    @Override
    public boolean preHandle(HttpServletRequest request, HttpServletResponse response, Object handler) {
        String userId = request.getHeader("X-User-Id");
        String userName = request.getHeader("X-User-Name");
        String userRole = request.getHeader("X-User-Role");

        if (userId == null || userId.isEmpty()) {
            userId = "anonymous";
            userName = "Anonymous";
            userRole = EmployeeConstants.ROLE_EMPLOYEE;
        }

        if (userRole == null || userRole.isEmpty()) {
            userRole = EmployeeConstants.ROLE_EMPLOYEE;
        }

        UserContext.setCurrentUser(userId, userName, userRole);
        return true;
    }

    @Override
    public void afterCompletion(HttpServletRequest request, HttpServletResponse response, Object handler, Exception ex) {
        UserContext.clear();
    }
}

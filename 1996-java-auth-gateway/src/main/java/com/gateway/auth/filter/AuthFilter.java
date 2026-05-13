package com.gateway.auth.filter;

import com.gateway.auth.model.AuditLog;
import com.gateway.auth.model.JwtValidationResult;
import com.gateway.auth.service.AuditLogService;
import com.gateway.auth.service.JwtService;
import org.springframework.core.Ordered;
import org.springframework.core.annotation.Order;
import org.springframework.http.HttpStatus;
import org.springframework.stereotype.Component;

import javax.servlet.*;
import javax.servlet.http.HttpServletRequest;
import javax.servlet.http.HttpServletResponse;
import java.io.IOException;
import java.time.Instant;

@Component
@Order(Ordered.HIGHEST_PRECEDENCE)
public class AuthFilter implements Filter {

    private static final String ADMIN_PATH_PREFIX = "/api/admin/";

    private final JwtService jwtService;
    private final AuditLogService auditLogService;

    public AuthFilter(JwtService jwtService, AuditLogService auditLogService) {
        this.jwtService = jwtService;
        this.auditLogService = auditLogService;
    }

    @Override
    public void doFilter(ServletRequest request, ServletResponse response, FilterChain chain)
            throws IOException, ServletException {

        HttpServletRequest httpRequest = (HttpServletRequest) request;
        HttpServletResponse httpResponse = (HttpServletResponse) response;

        String path = httpRequest.getRequestURI();
        String clientId = null;

        try {
            String authorizationHeader = httpRequest.getHeader("Authorization");
            String token = jwtService.extractTokenFromHeader(authorizationHeader);

            JwtValidationResult validationResult;
            if (token == null) {
                validationResult = JwtValidationResult.invalid("Missing Authorization header");
            } else {
                validationResult = jwtService.validateToken(token);
            }

            clientId = validationResult.getClientId();

            switch (validationResult.getStatus()) {
                case VALID:
                    String role = validationResult.getRole();

                    if (path.startsWith(ADMIN_PATH_PREFIX) && !"admin".equals(role)) {
                        auditLogService.log(new AuditLog(Instant.now(), clientId, path, AuditLog.Result.REJECTED));
                        httpResponse.setStatus(HttpStatus.FORBIDDEN.value());
                        httpResponse.getWriter().write("Forbidden: Admin role required");
                        return;
                    }

                    auditLogService.log(new AuditLog(Instant.now(), clientId, path, AuditLog.Result.PASSED));

                    httpRequest.setAttribute("clientId", clientId);
                    httpRequest.setAttribute("role", role);

                    chain.doFilter(request, response);
                    return;

                case EXPIRED:
                    auditLogService.log(new AuditLog(Instant.now(), clientId, path, AuditLog.Result.EXPIRED));
                    httpResponse.setStatus(HttpStatus.UNAUTHORIZED.value());
                    httpResponse.getWriter().write("Unauthorized: Token expired");
                    return;

                case INVALID:
                default:
                    auditLogService.log(new AuditLog(Instant.now(), clientId, path, AuditLog.Result.REJECTED));
                    httpResponse.setStatus(HttpStatus.UNAUTHORIZED.value());
                    httpResponse.getWriter().write("Unauthorized: Invalid token");
                    return;
            }

        } catch (Exception e) {
            auditLogService.log(new AuditLog(Instant.now(), clientId, path, AuditLog.Result.REJECTED));
            httpResponse.setStatus(HttpStatus.INTERNAL_SERVER_ERROR.value());
            httpResponse.getWriter().write("Internal Server Error");
        }
    }
}

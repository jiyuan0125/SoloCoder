package com.gateway.core;

import com.gateway.forward.ForwardException;
import lombok.extern.slf4j.Slf4j;
import org.springframework.stereotype.Component;
import org.springframework.web.servlet.HandlerInterceptor;

import javax.servlet.http.HttpServletRequest;
import javax.servlet.http.HttpServletResponse;

@Slf4j
@Component
public class GatewayServlet implements HandlerInterceptor {

    private final GatewayProcessor gatewayProcessor;
    private final ErrorResponseWriter errorResponseWriter;

    public GatewayServlet(GatewayProcessor gatewayProcessor, ErrorResponseWriter errorResponseWriter) {
        this.gatewayProcessor = gatewayProcessor;
        this.errorResponseWriter = errorResponseWriter;
    }

    @Override
    public boolean preHandle(HttpServletRequest request, HttpServletResponse response, Object handler) throws Exception {
        try {
            gatewayProcessor.process(request, response);
        } catch (ForwardException e) {
            log.error("转发异常: {} - {} (后端: {})", e.getErrorCode(), e.getMessage(), e.getBackendUrl());
            errorResponseWriter.writeError(
                    response,
                    e.getHttpStatus(),
                    e.getErrorCode(),
                    e.getMessage(),
                    e.getBackendUrl()
            );
        } catch (GatewayProcessor.FilterRejectException e) {
            log.warn("过滤器拒绝请求: {} - {}", e.getErrorCode(), e.getMessage());
            errorResponseWriter.writeError(
                    response,
                    e.getHttpStatus(),
                    e.getErrorCode(),
                    e.getMessage()
            );
        } catch (Exception e) {
            log.error("服务器内部错误", e);
            errorResponseWriter.writeError(
                    response,
                    500,
                    "INTERNAL_ERROR",
                    "服务器内部错误"
            );
        }
        return false;
    }
}

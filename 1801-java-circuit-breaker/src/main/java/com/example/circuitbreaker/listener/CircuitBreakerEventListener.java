package com.example.circuitbreaker.listener;

import com.example.circuitbreaker.core.CircuitStateChangedEvent;
import lombok.extern.slf4j.Slf4j;
import org.springframework.context.event.EventListener;
import org.springframework.stereotype.Component;

@Slf4j
@Component
public class CircuitBreakerEventListener {

    @EventListener
    public void handleCircuitStateChange(CircuitStateChangedEvent event) {
        log.info("收到熔断器状态变更事件: 服务={}, 从={}, 到={}, 原因={}",
                event.getServiceName(),
                event.getFromState(),
                event.getToState(),
                event.getReason());

        try {
            sendDingTalkAlert(event);
        } catch (Exception e) {
            log.error("发送钉钉告警失败，但不影响状态转换", e);
        }

        try {
            recordAuditLog(event);
        } catch (Exception e) {
            log.error("记录审计日志失败，但不影响状态转换", e);
        }

        try {
            triggerDegradation(event);
        } catch (Exception e) {
            log.error("触发降级失败，但不影响状态转换", e);
        }
    }

    private void sendDingTalkAlert(CircuitStateChangedEvent event) {
        String message = String.format("【熔断器告警】服务: %s, 状态: %s -> %s, 原因: %s",
                event.getServiceName(),
                event.getFromState(),
                event.getToState(),
                event.getReason());
        log.info("模拟发送钉钉告警: {}", message);
    }

    private void recordAuditLog(CircuitStateChangedEvent event) {
        log.info("记录审计日志 - 服务: {}, 状态变更: {} -> {}, 原因: {}",
                event.getServiceName(),
                event.getFromState(),
                event.getToState(),
                event.getReason());
    }

    private void triggerDegradation(CircuitStateChangedEvent event) {
        switch (event.getToState()) {
            case OPEN:
                log.info("触发服务 {} 的降级策略 - 切换到备用方案", event.getServiceName());
                break;
            case HALF_OPEN:
                log.info("服务 {} 进入半开状态，开始放行探测请求", event.getServiceName());
                break;
            case CLOSED:
                log.info("服务 {} 恢复正常，关闭降级策略", event.getServiceName());
                break;
        }
    }
}

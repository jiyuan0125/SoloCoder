package com.example.circuitbreaker.service;

import com.example.circuitbreaker.core.CircuitBreaker;
import com.example.circuitbreaker.core.CircuitBreakerManager;
import com.example.circuitbreaker.core.FailureClassifier;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.stereotype.Service;

import java.util.function.Supplier;

@Slf4j
@Service
@RequiredArgsConstructor
public class DownstreamServiceCaller {
    private final CircuitBreakerManager manager;

    public <T> T callWithCircuitBreaker(String serviceName, Supplier<T> action, T fallback) {
        CircuitBreaker cb = manager.getOrCreate(serviceName);

        if (!cb.isCallPermitted()) {
            log.warn("熔断器 [{}] 已打开，快速失败，使用降级策略", serviceName);
            return fallback;
        }

        try {
            T result = action.get();
            cb.recordSuccess();
            return result;
        } catch (Exception e) {
            if (FailureClassifier.shouldCountAsFailure(e)) {
                log.error("下游服务 [{}] 调用失败（计数为熔断失败）: {}", serviceName, e.getMessage());
                cb.recordFailure();
            } else {
                log.warn("下游服务 [{}] 调用失败（不计入熔断失败）: {}", serviceName, e.getMessage());
                cb.recordSuccess();
            }
            return fallback;
        }
    }

    public String callPaymentService(String orderId) {
        return callWithCircuitBreaker("payment-service",
                () -> {
                    log.info("调用支付服务处理订单: {}", orderId);
                    return "支付成功: " + orderId;
                },
                "支付降级：订单 " + orderId + " 已进入队列，稍后重试"
        );
    }

    public String callInventoryService(String productId) {
        return callWithCircuitBreaker("inventory-service",
                () -> {
                    log.info("调用库存服务查询产品: {}", productId);
                    return "库存充足: " + productId;
                },
                "库存降级：产品 " + productId + " 库存信息暂时不可用"
        );
    }
}

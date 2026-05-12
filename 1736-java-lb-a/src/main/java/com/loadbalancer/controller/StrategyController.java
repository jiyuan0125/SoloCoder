package com.loadbalancer.controller;

import com.loadbalancer.model.LoadBalancingStrategy;
import com.loadbalancer.model.StrategyState;
import com.loadbalancer.service.StrategyManager;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import java.util.Map;

@RestController
@RequestMapping("/api/strategy")
public class StrategyController {

    private final StrategyManager strategyManager;

    public StrategyController(StrategyManager strategyManager) {
        this.strategyManager = strategyManager;
    }

    @GetMapping
    public ResponseEntity<Map<String, Object>> getStatus() {
        return ResponseEntity.ok(Map.of(
                "currentStrategy", strategyManager.getCurrentStrategy(),
                "pendingStrategy", strategyManager.getPendingStrategy() != null ? strategyManager.getPendingStrategy() : "none",
                "state", strategyManager.getCurrentState()
        ));
    }

    @PostMapping("/configure")
    public ResponseEntity<?> configureStrategy(@RequestBody Map<String, String> request) {
        String strategyName = request.get("strategy");
        if (strategyName == null) {
            return ResponseEntity.badRequest()
                    .body(Map.of("error", "strategy is required"));
        }

        try {
            LoadBalancingStrategy strategy = LoadBalancingStrategy.valueOf(strategyName);
            boolean success = strategyManager.configureStrategy(strategy);

            if (!success) {
                return ResponseEntity.badRequest()
                        .body(Map.of("error", "Cannot configure strategy in current state: " + strategyManager.getCurrentState()));
            }

            return ResponseEntity.ok(Map.of(
                    "status", "configured",
                    "pendingStrategy", strategy,
                    "state", StrategyState.CONFIGURING
            ));
        } catch (IllegalArgumentException e) {
            return ResponseEntity.badRequest()
                    .body(Map.of("error", "Invalid strategy: " + strategyName));
        }
    }

    @PostMapping("/activate")
    public ResponseEntity<?> startActivation() {
        boolean success = strategyManager.startActivation();

        if (!success) {
            return ResponseEntity.badRequest()
                    .body(Map.of("error", "Cannot start activation in current state: " + strategyManager.getCurrentState()));
        }

        return ResponseEntity.ok(Map.of(
                "status", "activating",
                "pendingStrategy", strategyManager.getPendingStrategy(),
                "state", StrategyState.ACTIVATING
        ));
    }

    @PostMapping("/commit")
    public ResponseEntity<?> commit() {
        boolean success = strategyManager.commit();

        if (!success) {
            return ResponseEntity.badRequest()
                    .body(Map.of("error", "Cannot commit in current state: " + strategyManager.getCurrentState()));
        }

        return ResponseEntity.ok(Map.of(
                "status", "committed",
                "currentStrategy", strategyManager.getCurrentStrategy(),
                "state", StrategyState.ACTIVE
        ));
    }

    @PostMapping("/cancel")
    public ResponseEntity<?> cancel() {
        boolean success = strategyManager.cancelConfiguration();

        if (!success) {
            return ResponseEntity.badRequest()
                    .body(Map.of("error", "No pending configuration to cancel"));
        }

        return ResponseEntity.ok(Map.of(
                "status", "cancelled",
                "currentStrategy", strategyManager.getCurrentStrategy(),
                "state", StrategyState.ACTIVE
        ));
    }
}

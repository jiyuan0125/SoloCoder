package com.example.loadbalancer.controller;

import com.example.loadbalancer.dto.StrategyRequest;
import com.example.loadbalancer.strategy.StrategyManager;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import javax.validation.Valid;
import java.util.Map;

@RestController
@RequestMapping("/api/strategy")
public class StrategyController {

    private final StrategyManager strategyManager;

    public StrategyController(StrategyManager strategyManager) {
        this.strategyManager = strategyManager;
    }

    @PutMapping
    public ResponseEntity<?> switchStrategy(@Valid @RequestBody StrategyRequest request) {
        String strategyName = request.getStrategy();
        
        if (!isValidStrategy(strategyName)) {
            return ResponseEntity.status(HttpStatus.BAD_REQUEST)
                    .body(Map.of(
                            "error", "Invalid strategy",
                            "validStrategies", new String[]{"round_robin", "weighted_round_robin", "least_conn"}
                    ));
        }

        boolean success = strategyManager.switchStrategy(strategyName);
        
        if (!success) {
            return ResponseEntity.status(HttpStatus.BAD_REQUEST)
                    .body(Map.of("error", "Failed to switch strategy"));
        }

        return ResponseEntity.ok(Map.of(
                "success", true,
                "currentStrategy", strategyManager.getCurrentStrategyName()
        ));
    }

    @GetMapping
    public ResponseEntity<?> getCurrentStrategy() {
        return ResponseEntity.ok(Map.of(
                "currentStrategy", strategyManager.getCurrentStrategyName()
        ));
    }

    private boolean isValidStrategy(String strategy) {
        if (strategy == null) return false;
        String lower = strategy.toLowerCase();
        return "round_robin".equals(lower)
                || "weighted_round_robin".equals(lower)
                || "least_conn".equals(lower);
    }
}

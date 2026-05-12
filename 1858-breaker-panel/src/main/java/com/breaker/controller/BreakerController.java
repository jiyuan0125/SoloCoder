package com.breaker.controller;

import com.breaker.core.BreakerManager;
import com.breaker.core.CircuitBreaker;
import com.breaker.dto.ConfigRequest;
import com.breaker.model.BreakerConfig;
import com.breaker.model.BreakerDetail;
import com.breaker.model.BreakerSummary;
import org.springframework.http.ResponseEntity;
import org.springframework.validation.annotation.Validated;
import org.springframework.web.bind.annotation.*;

import javax.validation.Valid;
import java.util.List;
import java.util.Optional;

@RestController
@RequestMapping("/breakers")
@Validated
public class BreakerController {

    private final BreakerManager breakerManager;

    public BreakerController(BreakerManager breakerManager) {
        this.breakerManager = breakerManager;
    }

    @GetMapping
    public List<BreakerSummary> getAllBreakers() {
        return breakerManager.getAllBreakers();
    }

    @GetMapping("/{service}")
    public ResponseEntity<BreakerDetail> getBreaker(@PathVariable String service) {
        Optional<BreakerDetail> detail = breakerManager.getBreakerDetail(service);
        if (detail.isPresent()) {
            return ResponseEntity.ok(detail.get());
        }
        return ResponseEntity.notFound().build();
    }

    @PutMapping("/{service}/config")
    public ResponseEntity<Void> updateConfig(
            @PathVariable String service,
            @Valid @RequestBody ConfigRequest request) {

        if (request.getFailureThreshold() == null || request.getOpenDurationSeconds() == null) {
            return ResponseEntity.badRequest().build();
        }

        BreakerConfig config = new BreakerConfig(
                request.getFailureThreshold(),
                request.getOpenDurationSeconds()
        );
        breakerManager.createOrUpdateBreaker(service, config);
        return ResponseEntity.ok().build();
    }
}

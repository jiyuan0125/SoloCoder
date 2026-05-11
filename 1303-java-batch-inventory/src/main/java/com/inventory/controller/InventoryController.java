package com.inventory.controller;

import com.inventory.dto.*;
import com.inventory.service.ExpiryAlertService;
import com.inventory.service.InventoryService;
import jakarta.validation.Valid;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;
import java.util.List;

@RestController
@RequestMapping("/api/inventory")
public class InventoryController {

    @Autowired
    private InventoryService inventoryService;

    @Autowired
    private ExpiryAlertService expiryAlertService;

    @PostMapping("/inbound")
    public ResponseEntity<BatchResponse> inbound(@Valid @RequestBody InboundRequest request) {
        BatchResponse response = inventoryService.inbound(request);
        return ResponseEntity.status(HttpStatus.CREATED).body(response);
    }

    @PostMapping("/outbound")
    public ResponseEntity<OutboundResponse> outbound(@Valid @RequestBody OutboundRequest request) {
        OutboundResponse response = inventoryService.outbound(request);
        return ResponseEntity.ok(response);
    }

    @GetMapping
    public ResponseEntity<List<InventoryResponse>> getInventory() {
        return ResponseEntity.ok(inventoryService.getInventory());
    }

    @GetMapping("/product/{productId}")
    public ResponseEntity<List<InventoryResponse>> getInventoryByProductId(@PathVariable Long productId) {
        return ResponseEntity.ok(inventoryService.getInventoryByProductId(productId));
    }

    @GetMapping("/expiry-alerts")
    public ResponseEntity<List<ExpiryAlertResponse>> getExpiryAlerts(
            @RequestParam(name = "days", required = false) Integer daysThreshold) {
        List<ExpiryAlertResponse> alerts;
        if (daysThreshold != null && daysThreshold > 0) {
            alerts = expiryAlertService.getExpiryAlerts(daysThreshold);
        } else {
            alerts = expiryAlertService.getExpiryAlerts();
        }
        return ResponseEntity.ok(alerts);
    }

    @GetMapping("/expired")
    public ResponseEntity<List<ExpiryAlertResponse>> getExpiredProducts() {
        return ResponseEntity.ok(expiryAlertService.getExpiredProducts());
    }
}

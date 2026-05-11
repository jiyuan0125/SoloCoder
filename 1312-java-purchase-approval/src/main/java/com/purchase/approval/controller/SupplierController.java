package com.purchase.approval.controller;

import com.purchase.approval.dto.BlacklistDTO;
import com.purchase.approval.dto.SupplierDTO;
import com.purchase.approval.entity.Supplier;
import com.purchase.approval.exception.BusinessException;
import com.purchase.approval.service.SupplierService;
import lombok.RequiredArgsConstructor;
import org.springframework.http.ResponseEntity;
import org.springframework.validation.annotation.Validated;
import org.springframework.web.bind.annotation.*;

import java.util.HashMap;
import java.util.List;
import java.util.Map;

@RestController
@RequestMapping("/api/suppliers")
@RequiredArgsConstructor
public class SupplierController {

    private final SupplierService supplierService;

    @PostMapping
    public ResponseEntity<Map<String, Object>> createSupplier(@Validated @RequestBody SupplierDTO dto) {
        try {
            Supplier supplier = supplierService.createSupplier(dto);
            Map<String, Object> result = new HashMap<>();
            result.put("success", true);
            result.put("data", supplier);
            result.put("message", "供应商创建成功");
            return ResponseEntity.ok(result);
        } catch (BusinessException e) {
            Map<String, Object> result = new HashMap<>();
            result.put("success", false);
            result.put("message", e.getMessage());
            return ResponseEntity.badRequest().body(result);
        }
    }

    @GetMapping("/{id}")
    public ResponseEntity<Map<String, Object>> getSupplier(@PathVariable Long id) {
        try {
            Supplier supplier = supplierService.getSupplier(id);
            Map<String, Object> result = new HashMap<>();
            result.put("success", true);
            result.put("data", supplier);
            return ResponseEntity.ok(result);
        } catch (BusinessException e) {
            Map<String, Object> result = new HashMap<>();
            result.put("success", false);
            result.put("message", e.getMessage());
            return ResponseEntity.badRequest().body(result);
        }
    }

    @GetMapping
    public ResponseEntity<Map<String, Object>> getAllSuppliers() {
        List<Supplier> suppliers = supplierService.getAllSuppliers();
        Map<String, Object> result = new HashMap<>();
        result.put("success", true);
        result.put("data", suppliers);
        return ResponseEntity.ok(result);
    }

    @GetMapping("/blacklisted")
    public ResponseEntity<Map<String, Object>> getBlacklistedSuppliers() {
        List<Supplier> suppliers = supplierService.getBlacklistedSuppliers();
        Map<String, Object> result = new HashMap<>();
        result.put("success", true);
        result.put("data", suppliers);
        return ResponseEntity.ok(result);
    }

    @GetMapping("/{id}/blacklist-status")
    public ResponseEntity<Map<String, Object>> checkBlacklistStatus(@PathVariable Long id) {
        boolean isBlacklisted = supplierService.isSupplierBlacklisted(id);
        Map<String, Object> result = new HashMap<>();
        result.put("success", true);
        result.put("isBlacklisted", isBlacklisted);
        return ResponseEntity.ok(result);
    }

    @PostMapping("/blacklist")
    public ResponseEntity<Map<String, Object>> addToBlacklist(@Validated @RequestBody BlacklistDTO dto) {
        try {
            Supplier supplier = supplierService.addToBlacklist(dto);
            Map<String, Object> result = new HashMap<>();
            result.put("success", true);
            result.put("data", supplier);
            result.put("message", "供应商已加入黑名单");
            return ResponseEntity.ok(result);
        } catch (BusinessException e) {
            Map<String, Object> result = new HashMap<>();
            result.put("success", false);
            result.put("message", e.getMessage());
            return ResponseEntity.badRequest().body(result);
        }
    }

    @PostMapping("/{id}/remove-blacklist")
    public ResponseEntity<Map<String, Object>> removeFromBlacklist(@PathVariable Long id) {
        try {
            Supplier supplier = supplierService.removeFromBlacklist(id);
            Map<String, Object> result = new HashMap<>();
            result.put("success", true);
            result.put("data", supplier);
            result.put("message", "供应商已从黑名单移除");
            return ResponseEntity.ok(result);
        } catch (BusinessException e) {
            Map<String, Object> result = new HashMap<>();
            result.put("success", false);
            result.put("message", e.getMessage());
            return ResponseEntity.badRequest().body(result);
        }
    }
}

package com.purchase.approval.controller;

import com.purchase.approval.dto.ApprovalDTO;
import com.purchase.approval.dto.ReSubmitDTO;
import com.purchase.approval.entity.ApprovalRecord;
import com.purchase.approval.entity.PurchaseRequest;
import com.purchase.approval.enums.ApprovalLevel;
import com.purchase.approval.exception.BusinessException;
import com.purchase.approval.service.ApprovalService;
import lombok.RequiredArgsConstructor;
import org.springframework.http.ResponseEntity;
import org.springframework.validation.annotation.Validated;
import org.springframework.web.bind.annotation.*;

import java.util.HashMap;
import java.util.List;
import java.util.Map;

@RestController
@RequestMapping("/api/approvals")
@RequiredArgsConstructor
public class ApprovalController {

    private final ApprovalService approvalService;

    @PostMapping("/purchase-request/{requestId}/approve/{level}")
    public ResponseEntity<Map<String, Object>> approvePurchaseRequest(
            @PathVariable Long requestId,
            @PathVariable String level,
            @Validated @RequestBody ApprovalDTO dto) {
        try {
            ApprovalLevel approvalLevel = ApprovalLevel.valueOf(level.toUpperCase());
            PurchaseRequest request = approvalService.approvePurchaseRequest(requestId, approvalLevel, dto);
            
            Map<String, Object> result = new HashMap<>();
            result.put("success", true);
            result.put("data", request);
            result.put("message", dto.getApproved() ? "审批通过" : "审批驳回");
            return ResponseEntity.ok(result);
        } catch (IllegalArgumentException e) {
            Map<String, Object> result = new HashMap<>();
            result.put("success", false);
            result.put("message", "无效的审批级别");
            return ResponseEntity.badRequest().body(result);
        } catch (BusinessException e) {
            Map<String, Object> result = new HashMap<>();
            result.put("success", false);
            result.put("message", e.getMessage());
            return ResponseEntity.badRequest().body(result);
        }
    }

    @PostMapping("/purchase-request/{requestId}/resubmit")
    public ResponseEntity<Map<String, Object>> reSubmitPurchaseRequest(
            @PathVariable Long requestId,
            @Validated @RequestBody ReSubmitDTO dto) {
        try {
            PurchaseRequest request = approvalService.reSubmitPurchaseRequest(requestId, dto);
            Map<String, Object> result = new HashMap<>();
            result.put("success", true);
            result.put("data", request);
            result.put("message", "重新提交成功");
            return ResponseEntity.ok(result);
        } catch (BusinessException e) {
            Map<String, Object> result = new HashMap<>();
            result.put("success", false);
            result.put("message", e.getMessage());
            return ResponseEntity.badRequest().body(result);
        }
    }

    @GetMapping("/purchase-request/{requestId}/records")
    public ResponseEntity<Map<String, Object>> getApprovalRecords(@PathVariable Long requestId) {
        List<ApprovalRecord> records = approvalService.getApprovalRecords(requestId);
        Map<String, Object> result = new HashMap<>();
        result.put("success", true);
        result.put("data", records);
        return ResponseEntity.ok(result);
    }
}

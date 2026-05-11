package com.purchase.approval.controller;

import com.purchase.approval.dto.PurchaseRequestDTO;
import com.purchase.approval.dto.QuoteDTO;
import com.purchase.approval.entity.PurchaseRequest;
import com.purchase.approval.entity.Quote;
import com.purchase.approval.exception.BusinessException;
import com.purchase.approval.service.PurchaseRequestService;
import lombok.RequiredArgsConstructor;
import org.springframework.http.ResponseEntity;
import org.springframework.validation.annotation.Validated;
import org.springframework.web.bind.annotation.*;

import java.util.HashMap;
import java.util.List;
import java.util.Map;

@RestController
@RequestMapping("/api/purchase-requests")
@RequiredArgsConstructor
public class PurchaseRequestController {

    private final PurchaseRequestService purchaseRequestService;

    @PostMapping
    public ResponseEntity<Map<String, Object>> createDraft(@Validated @RequestBody PurchaseRequestDTO dto) {
        try {
            PurchaseRequest request = purchaseRequestService.createDraft(dto);
            Map<String, Object> result = new HashMap<>();
            result.put("success", true);
            result.put("data", request);
            result.put("message", "采购申请草稿创建成功");
            return ResponseEntity.ok(result);
        } catch (BusinessException e) {
            Map<String, Object> result = new HashMap<>();
            result.put("success", false);
            result.put("message", e.getMessage());
            return ResponseEntity.badRequest().body(result);
        }
    }

    @PostMapping("/{id}/submit")
    public ResponseEntity<Map<String, Object>> submitRequest(@PathVariable Long id) {
        try {
            PurchaseRequest request = purchaseRequestService.submitRequest(id);
            Map<String, Object> result = new HashMap<>();
            result.put("success", true);
            result.put("data", request);
            result.put("message", "采购申请提交成功");
            return ResponseEntity.ok(result);
        } catch (BusinessException e) {
            Map<String, Object> result = new HashMap<>();
            result.put("success", false);
            result.put("message", e.getMessage());
            return ResponseEntity.badRequest().body(result);
        }
    }

    @GetMapping("/{id}")
    public ResponseEntity<Map<String, Object>> getRequest(@PathVariable Long id) {
        try {
            PurchaseRequest request = purchaseRequestService.getRequest(id);
            Map<String, Object> result = new HashMap<>();
            result.put("success", true);
            result.put("data", request);
            return ResponseEntity.ok(result);
        } catch (BusinessException e) {
            Map<String, Object> result = new HashMap<>();
            result.put("success", false);
            result.put("message", e.getMessage());
            return ResponseEntity.badRequest().body(result);
        }
    }

    @PostMapping("/{id}/quotes")
    public ResponseEntity<Map<String, Object>> addQuote(
            @PathVariable Long id,
            @Validated @RequestBody QuoteDTO dto) {
        try {
            Quote quote = purchaseRequestService.addQuote(id, dto);
            Map<String, Object> result = new HashMap<>();
            result.put("success", true);
            result.put("data", quote);
            result.put("message", "报价添加成功");
            return ResponseEntity.ok(result);
        } catch (BusinessException e) {
            Map<String, Object> result = new HashMap<>();
            result.put("success", false);
            result.put("message", e.getMessage());
            return ResponseEntity.badRequest().body(result);
        }
    }

    @PutMapping("/{requestId}/quotes/{quoteId}")
    public ResponseEntity<Map<String, Object>> updateQuote(
            @PathVariable Long requestId,
            @PathVariable Long quoteId,
            @Validated @RequestBody QuoteDTO dto) {
        try {
            Quote quote = purchaseRequestService.updateQuote(requestId, quoteId, dto);
            Map<String, Object> result = new HashMap<>();
            result.put("success", true);
            result.put("data", quote);
            result.put("message", "报价更新成功");
            return ResponseEntity.ok(result);
        } catch (BusinessException e) {
            Map<String, Object> result = new HashMap<>();
            result.put("success", false);
            result.put("message", e.getMessage());
            return ResponseEntity.badRequest().body(result);
        }
    }

    @DeleteMapping("/{requestId}/quotes/{quoteId}")
    public ResponseEntity<Map<String, Object>> invalidateQuote(
            @PathVariable Long requestId,
            @PathVariable Long quoteId) {
        try {
            purchaseRequestService.invalidateQuote(requestId, quoteId);
            Map<String, Object> result = new HashMap<>();
            result.put("success", true);
            result.put("message", "报价已作废");
            return ResponseEntity.ok(result);
        } catch (BusinessException e) {
            Map<String, Object> result = new HashMap<>();
            result.put("success", false);
            result.put("message", e.getMessage());
            return ResponseEntity.badRequest().body(result);
        }
    }

    @GetMapping("/{id}/quotes")
    public ResponseEntity<Map<String, Object>> getValidQuotes(@PathVariable Long id) {
        List<Quote> quotes = purchaseRequestService.getValidQuotes(id);
        Map<String, Object> result = new HashMap<>();
        result.put("success", true);
        result.put("data", quotes);
        return ResponseEntity.ok(result);
    }

    @PostMapping("/{id}/select-supplier/{quoteId}")
    public ResponseEntity<Map<String, Object>> selectSupplier(
            @PathVariable Long id,
            @PathVariable Long quoteId) {
        try {
            PurchaseRequest request = purchaseRequestService.selectSupplier(id, quoteId);
            Map<String, Object> result = new HashMap<>();
            result.put("success", true);
            result.put("data", request);
            result.put("message", "供应商选择成功，进入审批流程");
            return ResponseEntity.ok(result);
        } catch (BusinessException e) {
            Map<String, Object> result = new HashMap<>();
            result.put("success", false);
            result.put("message", e.getMessage());
            return ResponseEntity.badRequest().body(result);
        }
    }

    @GetMapping("/submitter/{submitterId}")
    public ResponseEntity<Map<String, Object>> getRequestsBySubmitter(@PathVariable Long submitterId) {
        List<PurchaseRequest> requests = purchaseRequestService.getRequestsBySubmitter(submitterId);
        Map<String, Object> result = new HashMap<>();
        result.put("success", true);
        result.put("data", requests);
        return ResponseEntity.ok(result);
    }
}

package com.purchase.approval.controller;

import com.purchase.approval.dto.*;
import com.purchase.approval.entity.PurchaseOrder;
import com.purchase.approval.entity.ReturnRequest;
import com.purchase.approval.enums.ApprovalLevel;
import com.purchase.approval.enums.OrderStatus;
import com.purchase.approval.exception.BusinessException;
import com.purchase.approval.service.PurchaseOrderService;
import lombok.RequiredArgsConstructor;
import org.springframework.http.ResponseEntity;
import org.springframework.validation.annotation.Validated;
import org.springframework.web.bind.annotation.*;

import java.util.HashMap;
import java.util.List;
import java.util.Map;

@RestController
@RequestMapping("/api/orders")
@RequiredArgsConstructor
public class PurchaseOrderController {

    private final PurchaseOrderService purchaseOrderService;

    @PostMapping("/from-request/{requestId}")
    public ResponseEntity<Map<String, Object>> createOrder(@PathVariable Long requestId) {
        try {
            PurchaseOrder order = purchaseOrderService.createOrderFromApprovedRequest(requestId);
            Map<String, Object> result = new HashMap<>();
            result.put("success", true);
            result.put("data", order);
            result.put("message", "采购订单创建成功");
            return ResponseEntity.ok(result);
        } catch (BusinessException e) {
            Map<String, Object> result = new HashMap<>();
            result.put("success", false);
            result.put("message", e.getMessage());
            return ResponseEntity.badRequest().body(result);
        }
    }

    @GetMapping("/{orderId}")
    public ResponseEntity<Map<String, Object>> getOrder(@PathVariable Long orderId) {
        try {
            PurchaseOrder order = purchaseOrderService.getOrder(orderId);
            Map<String, Object> result = new HashMap<>();
            result.put("success", true);
            result.put("data", order);
            return ResponseEntity.ok(result);
        } catch (BusinessException e) {
            Map<String, Object> result = new HashMap<>();
            result.put("success", false);
            result.put("message", e.getMessage());
            return ResponseEntity.badRequest().body(result);
        }
    }

    @PostMapping("/{orderId}/confirm")
    public ResponseEntity<Map<String, Object>> confirmOrder(@PathVariable Long orderId) {
        try {
            PurchaseOrder order = purchaseOrderService.confirmOrder(orderId);
            Map<String, Object> result = new HashMap<>();
            result.put("success", true);
            result.put("data", order);
            result.put("message", "订单确认成功");
            return ResponseEntity.ok(result);
        } catch (BusinessException e) {
            Map<String, Object> result = new HashMap<>();
            result.put("success", false);
            result.put("message", e.getMessage());
            return ResponseEntity.badRequest().body(result);
        }
    }

    @PostMapping("/{orderId}/ship")
    public ResponseEntity<Map<String, Object>> shipOrder(@PathVariable Long orderId) {
        try {
            PurchaseOrder order = purchaseOrderService.shipOrder(orderId);
            Map<String, Object> result = new HashMap<>();
            result.put("success", true);
            result.put("data", order);
            result.put("message", "订单已发货");
            return ResponseEntity.ok(result);
        } catch (BusinessException e) {
            Map<String, Object> result = new HashMap<>();
            result.put("success", false);
            result.put("message", e.getMessage());
            return ResponseEntity.badRequest().body(result);
        }
    }

    @PostMapping("/{orderId}/receive")
    public ResponseEntity<Map<String, Object>> receivePartial(
            @PathVariable Long orderId,
            @Validated @RequestBody ReceiptDTO dto) {
        try {
            PurchaseOrder order = purchaseOrderService.receivePartial(orderId, dto);
            Map<String, Object> result = new HashMap<>();
            result.put("success", true);
            result.put("data", order);
            result.put("message", "收货成功");
            return ResponseEntity.ok(result);
        } catch (BusinessException e) {
            Map<String, Object> result = new HashMap<>();
            result.put("success", false);
            result.put("message", e.getMessage());
            return ResponseEntity.badRequest().body(result);
        }
    }

    @PostMapping("/{orderId}/inspect")
    public ResponseEntity<Map<String, Object>> inspectPartial(
            @PathVariable Long orderId,
            @Validated @RequestBody InspectionDTO dto) {
        try {
            PurchaseOrder order = purchaseOrderService.inspectPartial(orderId, dto);
            Map<String, Object> result = new HashMap<>();
            result.put("success", true);
            result.put("data", order);
            result.put("message", "验收成功");
            return ResponseEntity.ok(result);
        } catch (BusinessException e) {
            Map<String, Object> result = new HashMap<>();
            result.put("success", false);
            result.put("message", e.getMessage());
            return ResponseEntity.badRequest().body(result);
        }
    }

    @PostMapping("/{orderId}/complete")
    public ResponseEntity<Map<String, Object>> completeOrder(@PathVariable Long orderId) {
        try {
            PurchaseOrder order = purchaseOrderService.completeOrder(orderId);
            Map<String, Object> result = new HashMap<>();
            result.put("success", true);
            result.put("data", order);
            result.put("message", "订单完成");
            return ResponseEntity.ok(result);
        } catch (BusinessException e) {
            Map<String, Object> result = new HashMap<>();
            result.put("success", false);
            result.put("message", e.getMessage());
            return ResponseEntity.badRequest().body(result);
        }
    }

    @GetMapping("/status/{status}")
    public ResponseEntity<Map<String, Object>> getOrdersByStatus(@PathVariable String status) {
        try {
            OrderStatus orderStatus = OrderStatus.valueOf(status.toUpperCase());
            List<PurchaseOrder> orders = purchaseOrderService.getOrdersByStatus(orderStatus);
            Map<String, Object> result = new HashMap<>();
            result.put("success", true);
            result.put("data", orders);
            return ResponseEntity.ok(result);
        } catch (IllegalArgumentException e) {
            Map<String, Object> result = new HashMap<>();
            result.put("success", false);
            result.put("message", "无效的订单状态");
            return ResponseEntity.badRequest().body(result);
        }
    }

    @PostMapping("/returns")
    public ResponseEntity<Map<String, Object>> createReturnRequest(
            @Validated @RequestBody ReturnRequestDTO dto) {
        try {
            ReturnRequest returnRequest = purchaseOrderService.createReturnRequest(dto);
            Map<String, Object> result = new HashMap<>();
            result.put("success", true);
            result.put("data", returnRequest);
            result.put("message", "退货申请创建成功");
            return ResponseEntity.ok(result);
        } catch (BusinessException e) {
            Map<String, Object> result = new HashMap<>();
            result.put("success", false);
            result.put("message", e.getMessage());
            return ResponseEntity.badRequest().body(result);
        }
    }

    @PostMapping("/returns/{returnId}/submit")
    public ResponseEntity<Map<String, Object>> submitReturnRequest(@PathVariable Long returnId) {
        try {
            ReturnRequest returnRequest = purchaseOrderService.submitReturnRequest(returnId);
            Map<String, Object> result = new HashMap<>();
            result.put("success", true);
            result.put("data", returnRequest);
            result.put("message", "退货申请提交成功");
            return ResponseEntity.ok(result);
        } catch (BusinessException e) {
            Map<String, Object> result = new HashMap<>();
            result.put("success", false);
            result.put("message", e.getMessage());
            return ResponseEntity.badRequest().body(result);
        }
    }

    @PostMapping("/returns/{returnId}/approve/{level}")
    public ResponseEntity<Map<String, Object>> approveReturnRequest(
            @PathVariable Long returnId,
            @PathVariable String level,
            @Validated @RequestBody ApprovalDTO dto) {
        try {
            ApprovalLevel approvalLevel = ApprovalLevel.valueOf(level.toUpperCase());
            ReturnRequest returnRequest = purchaseOrderService.approveReturnRequest(returnId, approvalLevel, dto);
            
            Map<String, Object> result = new HashMap<>();
            result.put("success", true);
            result.put("data", returnRequest);
            result.put("message", dto.getApproved() ? "退货审批通过" : "退货审批驳回");
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
}

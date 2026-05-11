package com.delivery.controller;

import com.delivery.entity.Compensation;
import com.delivery.entity.ExceptionReport;
import com.delivery.enums.ExceptionReason;
import com.delivery.service.CompensationService;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;
import java.util.Optional;

@RestController
@RequestMapping("/api/compensation")
@RequiredArgsConstructor
@Slf4j
public class CompensationController {

    private final CompensationService compensationService;

    public static class ReportExceptionDTO {
        private Long deliveryPersonId;
        private ExceptionReason reason;
        private String description;
        private boolean isTimeoutExempt;

        public Long getDeliveryPersonId() { return deliveryPersonId; }
        public void setDeliveryPersonId(Long deliveryPersonId) { this.deliveryPersonId = deliveryPersonId; }
        public ExceptionReason getReason() { return reason; }
        public void setReason(ExceptionReason reason) { this.reason = reason; }
        public String getDescription() { return description; }
        public void setDescription(String description) { this.description = description; }
        public boolean isTimeoutExempt() { return isTimeoutExempt; }
        public void setTimeoutExempt(boolean timeoutExempt) { isTimeoutExempt = timeoutExempt; }
    }

    public static class ApproveReportDTO {
        private String approver;
        private String comment;

        public String getApprover() { return approver; }
        public void setApprover(String approver) { this.approver = approver; }
        public String getComment() { return comment; }
        public void setComment(String comment) { this.comment = comment; }
    }

    @GetMapping("/orders/{orderId}")
    public ResponseEntity<OrderController.ApiResponse<Compensation>> getCompensationByOrderId(
            @PathVariable Long orderId) {
        Optional<Compensation> compensation = compensationService.getCompensationByOrderId(orderId);
        return compensation
            .map(c -> ResponseEntity.ok(OrderController.ApiResponse.success(c)))
            .orElse(ResponseEntity.notFound().build());
    }

    @GetMapping("/orders/{orderId}/calculate")
    public ResponseEntity<OrderController.ApiResponse<CompensationService.CompensationResult>> calculateCompensation(
            @PathVariable Long orderId,
            @RequestParam(required = false) Long orderEntityId) {
        try {
            CompensationService.CompensationResult result;
            if (orderEntityId != null) {
                com.delivery.entity.DeliveryOrder order = 
                    new com.delivery.entity.DeliveryOrder();
                order.setId(orderEntityId);
                result = compensationService.calculateCompensation(order);
            } else {
                com.delivery.entity.DeliveryOrder order = 
                    new com.delivery.entity.DeliveryOrder();
                order.setId(orderId);
                result = compensationService.calculateCompensation(order);
            }
            return ResponseEntity.ok(OrderController.ApiResponse.success(result));
        } catch (Exception e) {
            log.error("计算赔偿失败", e);
            return ResponseEntity.badRequest().body(OrderController.ApiResponse.error(e.getMessage()));
        }
    }

    @PostMapping("/orders/{orderId}/report-exception")
    public ResponseEntity<OrderController.ApiResponse<ExceptionReport>> reportException(
            @PathVariable Long orderId,
            @RequestBody ReportExceptionDTO request) {
        try {
            ExceptionReport report = compensationService.reportException(
                orderId,
                request.getDeliveryPersonId(),
                request.getReason(),
                request.getDescription(),
                request.isTimeoutExempt()
            );
            return ResponseEntity.ok(OrderController.ApiResponse.success("异常报告已提交", report));
        } catch (Exception e) {
            log.error("提交异常报告失败", e);
            return ResponseEntity.badRequest().body(OrderController.ApiResponse.error(e.getMessage()));
        }
    }

    @PostMapping("/exception-reports/{reportId}/approve")
    public ResponseEntity<OrderController.ApiResponse<ExceptionReport>> approveExceptionReport(
            @PathVariable Long reportId,
            @RequestBody ApproveReportDTO request) {
        try {
            ExceptionReport report = compensationService.approveExceptionReport(
                reportId, request.getApprover(), request.getComment());
            return ResponseEntity.ok(OrderController.ApiResponse.success("审核通过", report));
        } catch (Exception e) {
            log.error("审核异常报告失败", e);
            return ResponseEntity.badRequest().body(OrderController.ApiResponse.error(e.getMessage()));
        }
    }
}

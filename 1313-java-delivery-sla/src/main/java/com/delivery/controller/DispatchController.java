package com.delivery.controller;

import com.delivery.entity.DeliveryPerson;
import com.delivery.entity.DeliveryZone;
import com.delivery.enums.TimeSlot;
import com.delivery.service.DispatchService;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;
import java.util.List;

@RestController
@RequestMapping("/api/dispatch")
@RequiredArgsConstructor
@Slf4j
public class DispatchController {

    private final DispatchService dispatchService;

    public static class AssignRequest {
        private TimeSlot timeSlot;
        
        public TimeSlot getTimeSlot() { return timeSlot; }
        public void setTimeSlot(TimeSlot timeSlot) { this.timeSlot = timeSlot; }
    }

    public static class ReassignRequest {
        private Long newPersonId;
        
        public Long getNewPersonId() { return newPersonId; }
        public void setNewPersonId(Long newPersonId) { this.newPersonId = newPersonId; }
    }

    @PostMapping("/orders/{orderId}/assign")
    public ResponseEntity<OrderController.ApiResponse<DispatchService.AssignmentResult>> assignDeliveryPerson(
            @PathVariable Long orderId,
            @RequestBody AssignRequest request) {
        try {
            DispatchService.AssignmentResult result = dispatchService.assignDeliveryPerson(
                orderId, request.getTimeSlot());
            return ResponseEntity.ok(OrderController.ApiResponse.success(result));
        } catch (Exception e) {
            log.error("分配配送员失败", e);
            return ResponseEntity.badRequest().body(OrderController.ApiResponse.error(e.getMessage()));
        }
    }

    @PostMapping("/orders/{orderId}/reassign")
    public ResponseEntity<OrderController.ApiResponse<String>> reassignDeliveryPerson(
            @PathVariable Long orderId,
            @RequestBody ReassignRequest request) {
        try {
            dispatchService.reassignDeliveryPerson(orderId, request.getNewPersonId());
            return ResponseEntity.ok(OrderController.ApiResponse.success("重新分配成功", null));
        } catch (Exception e) {
            log.error("重新分配配送员失败", e);
            return ResponseEntity.badRequest().body(OrderController.ApiResponse.error(e.getMessage()));
        }
    }

    @GetMapping("/delivery-persons/available")
    public ResponseEntity<OrderController.ApiResponse<List<DeliveryPerson>>> getAvailableDeliveryPersons(
            @RequestParam(required = false) String zoneCode,
            @RequestParam TimeSlot timeSlot) {
        List<DeliveryPerson> available = dispatchService.getAvailableDeliveryPersons(zoneCode, timeSlot);
        return ResponseEntity.ok(OrderController.ApiResponse.success(available));
    }

    @GetMapping("/zones")
    public ResponseEntity<OrderController.ApiResponse<List<DeliveryZone>>> getAllZones() {
        List<DeliveryZone> zones = dispatchService.getAllZones();
        return ResponseEntity.ok(OrderController.ApiResponse.success(zones));
    }
}

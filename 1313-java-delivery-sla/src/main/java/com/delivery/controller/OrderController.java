package com.delivery.controller;

import com.delivery.entity.DeliveryOrder;
import com.delivery.entity.OrderTracking;
import com.delivery.enums.DeliveryType;
import com.delivery.enums.TimeSlot;
import com.delivery.service.OrderService;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;
import java.math.BigDecimal;
import java.util.List;

@RestController
@RequestMapping("/api/orders")
@RequiredArgsConstructor
@Slf4j
public class OrderController {

    private final OrderService orderService;

    public static class CreateOrderDTO {
        private String customerCode;
        private DeliveryType deliveryType;
        private TimeSlot preferredTimeSlot;
        private String senderName;
        private String senderPhone;
        private String senderAddress;
        private Double senderLatitude;
        private Double senderLongitude;
        private String receiverName;
        private String receiverPhone;
        private String receiverAddress;
        private Double receiverLatitude;
        private Double receiverLongitude;
        private BigDecimal weight;
        private Double length;
        private Double width;
        private Double height;
        private BigDecimal distance;

        public String getCustomerCode() { return customerCode; }
        public void setCustomerCode(String customerCode) { this.customerCode = customerCode; }
        public DeliveryType getDeliveryType() { return deliveryType; }
        public void setDeliveryType(DeliveryType deliveryType) { this.deliveryType = deliveryType; }
        public TimeSlot getPreferredTimeSlot() { return preferredTimeSlot; }
        public void setPreferredTimeSlot(TimeSlot preferredTimeSlot) { this.preferredTimeSlot = preferredTimeSlot; }
        public String getSenderName() { return senderName; }
        public void setSenderName(String senderName) { this.senderName = senderName; }
        public String getSenderPhone() { return senderPhone; }
        public void setSenderPhone(String senderPhone) { this.senderPhone = senderPhone; }
        public String getSenderAddress() { return senderAddress; }
        public void setSenderAddress(String senderAddress) { this.senderAddress = senderAddress; }
        public Double getSenderLatitude() { return senderLatitude; }
        public void setSenderLatitude(Double senderLatitude) { this.senderLatitude = senderLatitude; }
        public Double getSenderLongitude() { return senderLongitude; }
        public void setSenderLongitude(Double senderLongitude) { this.senderLongitude = senderLongitude; }
        public String getReceiverName() { return receiverName; }
        public void setReceiverName(String receiverName) { this.receiverName = receiverName; }
        public String getReceiverPhone() { return receiverPhone; }
        public void setReceiverPhone(String receiverPhone) { this.receiverPhone = receiverPhone; }
        public String getReceiverAddress() { return receiverAddress; }
        public void setReceiverAddress(String receiverAddress) { this.receiverAddress = receiverAddress; }
        public Double getReceiverLatitude() { return receiverLatitude; }
        public void setReceiverLatitude(Double receiverLatitude) { this.receiverLatitude = receiverLatitude; }
        public Double getReceiverLongitude() { return receiverLongitude; }
        public void setReceiverLongitude(Double receiverLongitude) { this.receiverLongitude = receiverLongitude; }
        public BigDecimal getWeight() { return weight; }
        public void setWeight(BigDecimal weight) { this.weight = weight; }
        public Double getLength() { return length; }
        public void setLength(Double length) { this.length = length; }
        public Double getWidth() { return width; }
        public void setWidth(Double width) { this.width = width; }
        public Double getHeight() { return height; }
        public void setHeight(Double height) { this.height = height; }
        public BigDecimal getDistance() { return distance; }
        public void setDistance(BigDecimal distance) { this.distance = distance; }
    }

    public static class CancelOrderDTO {
        private String reason;
        
        public String getReason() { return reason; }
        public void setReason(String reason) { this.reason = reason; }
    }

    public static class ApiResponse<T> {
        private boolean success;
        private String message;
        private T data;

        public ApiResponse(boolean success, String message, T data) {
            this.success = success;
            this.message = message;
            this.data = data;
        }

        public static <T> ApiResponse<T> success(T data) {
            return new ApiResponse<>(true, "操作成功", data);
        }

        public static <T> ApiResponse<T> success(String message, T data) {
            return new ApiResponse<>(true, message, data);
        }

        public static <T> ApiResponse<T> error(String message) {
            return new ApiResponse<>(false, message, null);
        }

        public boolean isSuccess() { return success; }
        public String getMessage() { return message; }
        public T getData() { return data; }
    }

    @PostMapping
    public ResponseEntity<ApiResponse<DeliveryOrder>> createOrder(@RequestBody CreateOrderDTO request) {
        try {
            OrderService.CreateOrderRequest serviceRequest = new OrderService.CreateOrderRequest();
            serviceRequest.setCustomerCode(request.getCustomerCode());
            serviceRequest.setDeliveryType(request.getDeliveryType());
            serviceRequest.setPreferredTimeSlot(request.getPreferredTimeSlot());
            serviceRequest.setSenderName(request.getSenderName());
            serviceRequest.setSenderPhone(request.getSenderPhone());
            serviceRequest.setSenderAddress(request.getSenderAddress());
            serviceRequest.setSenderLatitude(request.getSenderLatitude());
            serviceRequest.setSenderLongitude(request.getSenderLongitude());
            serviceRequest.setReceiverName(request.getReceiverName());
            serviceRequest.setReceiverPhone(request.getReceiverPhone());
            serviceRequest.setReceiverAddress(request.getReceiverAddress());
            serviceRequest.setReceiverLatitude(request.getReceiverLatitude());
            serviceRequest.setReceiverLongitude(request.getReceiverLongitude());
            serviceRequest.setWeight(request.getWeight());
            serviceRequest.setLength(request.getLength());
            serviceRequest.setWidth(request.getWidth());
            serviceRequest.setHeight(request.getHeight());
            serviceRequest.setDistance(request.getDistance());

            DeliveryOrder order = orderService.createOrder(serviceRequest);
            return ResponseEntity.ok(ApiResponse.success("订单创建成功", order));
        } catch (Exception e) {
            log.error("创建订单失败", e);
            return ResponseEntity.badRequest().body(ApiResponse.error(e.getMessage()));
        }
    }

    @PostMapping("/{orderId}/confirm")
    public ResponseEntity<ApiResponse<DeliveryOrder>> confirmOrder(@PathVariable Long orderId) {
        try {
            DeliveryOrder order = orderService.confirmOrder(orderId);
            return ResponseEntity.ok(ApiResponse.success("订单确认成功", order));
        } catch (Exception e) {
            log.error("确认订单失败", e);
            return ResponseEntity.badRequest().body(ApiResponse.error(e.getMessage()));
        }
    }

    @PostMapping("/{orderId}/start-delivery")
    public ResponseEntity<ApiResponse<DeliveryOrder>> startDelivery(
            @PathVariable Long orderId, 
            @RequestParam Long deliveryPersonId) {
        try {
            DeliveryOrder order = orderService.startDelivery(orderId, deliveryPersonId);
            return ResponseEntity.ok(ApiResponse.success("开始配送", order));
        } catch (Exception e) {
            log.error("开始配送失败", e);
            return ResponseEntity.badRequest().body(ApiResponse.error(e.getMessage()));
        }
    }

    @PostMapping("/{orderId}/complete")
    public ResponseEntity<ApiResponse<DeliveryOrder>> completeDelivery(@PathVariable Long orderId) {
        try {
            DeliveryOrder order = orderService.completeDelivery(orderId);
            return ResponseEntity.ok(ApiResponse.success("配送完成", order));
        } catch (Exception e) {
            log.error("完成配送失败", e);
            return ResponseEntity.badRequest().body(ApiResponse.error(e.getMessage()));
        }
    }

    @PostMapping("/{orderId}/cancel")
    public ResponseEntity<ApiResponse<DeliveryOrder>> cancelOrder(
            @PathVariable Long orderId, 
            @RequestBody CancelOrderDTO request) {
        try {
            DeliveryOrder order = orderService.cancelOrder(orderId, request.getReason());
            return ResponseEntity.ok(ApiResponse.success("订单取消成功", order));
        } catch (Exception e) {
            log.error("取消订单失败", e);
            return ResponseEntity.badRequest().body(ApiResponse.error(e.getMessage()));
        }
    }

    @GetMapping("/{orderNo}")
    public ResponseEntity<ApiResponse<DeliveryOrder>> getOrderByNo(@PathVariable String orderNo) {
        return orderService.getOrderByNo(orderNo)
            .map(order -> ResponseEntity.ok(ApiResponse.success(order)))
            .orElse(ResponseEntity.notFound().build());
    }

    @GetMapping("/{orderId}/tracking")
    public ResponseEntity<ApiResponse<List<OrderTracking>>> getOrderTracking(@PathVariable Long orderId) {
        List<OrderTracking> trackingList = orderService.getOrderTracking(orderId);
        return ResponseEntity.ok(ApiResponse.success(trackingList));
    }

    @GetMapping("/customer/{customerId}")
    public ResponseEntity<ApiResponse<List<DeliveryOrder>>> getOrdersByCustomer(
            @PathVariable Long customerId) {
        List<DeliveryOrder> orders = orderService.getOrdersByCustomer(customerId);
        return ResponseEntity.ok(ApiResponse.success(orders));
    }
}

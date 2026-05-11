package com.delivery.service;

import com.delivery.entity.*;
import com.delivery.enums.DeliveryType;
import com.delivery.enums.OrderStatus;
import com.delivery.enums.TimeSlot;
import com.delivery.repository.*;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;
import java.math.BigDecimal;
import java.time.LocalDateTime;
import java.time.format.DateTimeFormatter;
import java.util.List;
import java.util.Optional;
import java.util.Random;

@Service
@RequiredArgsConstructor
@Slf4j
public class OrderService {

    private final DeliveryOrderRepository orderRepository;
    private final CustomerRepository customerRepository;
    private final OrderTrackingRepository trackingRepository;
    private final DeliveryTimeService deliveryTimeService;
    private final FeeCalculationService feeCalculationService;
    private final CompensationService compensationService;

    public static class CreateOrderRequest {
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

    @Transactional
    public DeliveryOrder createOrder(CreateOrderRequest request) {
        Customer customer = customerRepository.findByCustomerCode(request.getCustomerCode())
            .orElseThrow(() -> new RuntimeException("客户不存在: " + request.getCustomerCode()));
        
        LocalDateTime orderTime = LocalDateTime.now();
        
        FeeCalculationService.FeeBreakdown feeBreakdown = feeCalculationService.calculateFee(
            customer,
            request.getDeliveryType(),
            request.getWeight(),
            request.getDistance(),
            request.getLength(),
            request.getWidth(),
            request.getHeight()
        );
        
        LocalDateTime effectiveOrderTime = deliveryTimeService.calculateEffectiveOrderTime(
            request.getDeliveryType(), orderTime);
        
        LocalDateTime promisedDeliveryTime = deliveryTimeService.calculatePromisedDeliveryTime(
            request.getDeliveryType(), effectiveOrderTime);
        
        String orderNo = generateOrderNo();
        
        DeliveryOrder order = DeliveryOrder.builder()
            .orderNo(orderNo)
            .customer(customer)
            .deliveryType(request.getDeliveryType())
            .status(OrderStatus.PENDING)
            .preferredTimeSlot(request.getPreferredTimeSlot())
            .senderName(request.getSenderName())
            .senderPhone(request.getSenderPhone())
            .senderAddress(request.getSenderAddress())
            .senderLatitude(request.getSenderLatitude())
            .senderLongitude(request.getSenderLongitude())
            .receiverName(request.getReceiverName())
            .receiverPhone(request.getReceiverPhone())
            .receiverAddress(request.getReceiverAddress())
            .receiverLatitude(request.getReceiverLatitude())
            .receiverLongitude(request.getReceiverLongitude())
            .weight(request.getWeight())
            .length(request.getLength())
            .width(request.getWidth())
            .height(request.getHeight())
            .distance(request.getDistance())
            .baseFee(feeBreakdown.getBaseFee())
            .deliveryTypeSurcharge(feeBreakdown.getDeliveryTypeSurcharge())
            .weightSurcharge(feeBreakdown.getWeightSurcharge())
            .oversizedFee(feeBreakdown.getOversizedFee())
            .discount(feeBreakdown.getDiscount())
            .totalFee(feeBreakdown.getTotalFee())
            .monthlyOrderCount(feeBreakdown.getMonthlyOrderCount())
            .orderTime(orderTime)
            .effectiveOrderTime(effectiveOrderTime)
            .promisedDeliveryTime(promisedDeliveryTime)
            .build();
        
        order = orderRepository.save(order);
        
        createTracking(order, OrderStatus.PENDING, "订单已创建，等待处理", "SYSTEM", "系统");
        
        log.info("订单创建成功 - 订单号: {}, 客户: {}", orderNo, customer.getCustomerCode());
        return order;
    }

    @Transactional
    public DeliveryOrder confirmOrder(Long orderId) {
        DeliveryOrder order = orderRepository.findById(orderId)
            .orElseThrow(() -> new RuntimeException("订单不存在: " + orderId));
        
        if (order.getStatus() != OrderStatus.PENDING) {
            throw new RuntimeException("订单状态不正确，无法确认");
        }
        
        order.setStatus(OrderStatus.CONFIRMED);
        order = orderRepository.save(order);
        
        createTracking(order, OrderStatus.CONFIRMED, "订单已确认", "SYSTEM", "系统");
        
        log.info("订单确认成功 - 订单号: {}", order.getOrderNo());
        return order;
    }

    @Transactional
    public DeliveryOrder dispatchOrder(Long orderId, Long deliveryPersonId, String zoneCode) {
        DeliveryOrder order = orderRepository.findById(orderId)
            .orElseThrow(() -> new RuntimeException("订单不存在: " + orderId));
        
        if (order.getStatus() != OrderStatus.CONFIRMED) {
            throw new RuntimeException("订单状态不正确，无法出库");
        }
        
        order.setStatus(OrderStatus.DISPATCHED);
        order.setAssignedZoneCode(zoneCode);
        order = orderRepository.save(order);
        
        createTracking(order, OrderStatus.DISPATCHED, "订单已出库", "SYSTEM", "系统");
        
        log.info("订单出库成功 - 订单号: {}, 配送区域: {}", order.getOrderNo(), zoneCode);
        return order;
    }

    @Transactional
    public DeliveryOrder startDelivery(Long orderId, Long deliveryPersonId) {
        DeliveryOrder order = orderRepository.findById(orderId)
            .orElseThrow(() -> new RuntimeException("订单不存在: " + orderId));
        
        if (order.getStatus() != OrderStatus.DISPATCHED) {
            throw new RuntimeException("订单状态不正确，无法开始配送");
        }
        
        order.setStatus(OrderStatus.IN_TRANSIT);
        order = orderRepository.save(order);
        
        createTracking(order, OrderStatus.IN_TRANSIT, "配送员开始配送", "DELIVERY_PERSON", 
                      order.getDeliveryPerson() != null ? order.getDeliveryPerson().getName() : "配送员");
        
        log.info("开始配送 - 订单号: {}", order.getOrderNo());
        return order;
    }

    @Transactional
    public DeliveryOrder completeDelivery(Long orderId) {
        DeliveryOrder order = orderRepository.findById(orderId)
            .orElseThrow(() -> new RuntimeException("订单不存在: " + orderId));
        
        if (order.getStatus() != OrderStatus.IN_TRANSIT) {
            throw new RuntimeException("订单状态不正确，无法完成配送");
        }
        
        LocalDateTime actualDeliveryTime = LocalDateTime.now();
        order.setStatus(OrderStatus.DELIVERED);
        order.setActualDeliveryTime(actualDeliveryTime);
        order = orderRepository.save(order);
        
        createTracking(order, OrderStatus.DELIVERED, "订单已送达", "DELIVERY_PERSON",
                      order.getDeliveryPerson() != null ? order.getDeliveryPerson().getName() : "配送员");
        
        compensationService.checkAndCreateCompensation(order);
        
        log.info("配送完成 - 订单号: {}", order.getOrderNo());
        return order;
    }

    @Transactional
    public DeliveryOrder cancelOrder(Long orderId, String reason) {
        DeliveryOrder order = orderRepository.findById(orderId)
            .orElseThrow(() -> new RuntimeException("订单不存在: " + orderId));
        
        if (order.getStatus() == OrderStatus.DELIVERED || order.getStatus() == OrderStatus.CANCELLED) {
            throw new RuntimeException("订单状态不正确，无法取消");
        }
        
        BigDecimal cancellationFee = feeCalculationService.calculateCancellationFee(order);
        
        order.setStatus(OrderStatus.CANCELLED);
        order.setCancellationReason(reason);
        order.setCancelledAt(LocalDateTime.now());
        order = orderRepository.save(order);
        
        createTracking(order, OrderStatus.CANCELLED, 
                      "订单已取消，原因: " + reason + "，取消费用: " + cancellationFee, 
                      "SYSTEM", "系统");
        
        log.info("订单取消 - 订单号: {}, 费用: {}", order.getOrderNo(), cancellationFee);
        return order;
    }

    public Optional<DeliveryOrder> getOrderByNo(String orderNo) {
        return orderRepository.findByOrderNo(orderNo);
    }

    public List<DeliveryOrder> getOrdersByCustomer(Long customerId) {
        return orderRepository.findByCustomerId(customerId);
    }

    public List<OrderTracking> getOrderTracking(Long orderId) {
        return trackingRepository.findByOrderIdOrderByTrackingTimeDesc(orderId);
    }

    private String generateOrderNo() {
        String dateStr = LocalDateTime.now().format(DateTimeFormatter.ofPattern("yyyyMMddHHmmss"));
        String randomStr = String.format("%04d", new Random().nextInt(10000));
        return "DL" + dateStr + randomStr;
    }

    private void createTracking(DeliveryOrder order, OrderStatus status, String description, 
                                String operatorType, String operatorName) {
        OrderTracking tracking = OrderTracking.builder()
            .order(order)
            .status(status)
            .description(description)
            .operatorType(operatorType)
            .operatorName(operatorName)
            .build();
        trackingRepository.save(tracking);
    }
}

package com.foodordering.service;

import com.foodordering.entity.Order;
import com.foodordering.enums.OrderStatus;
import com.foodordering.exception.BusinessException;
import lombok.RequiredArgsConstructor;
import org.springframework.stereotype.Service;

@Service
@RequiredArgsConstructor
public class PaymentService {
    
    private final OrderService orderService;
    
    public Order payOrder(Long orderId) {
        Order order = orderService.getOrderById(orderId);
        
        if (order.getStatus() != OrderStatus.PENDING_PAYMENT) {
            throw new BusinessException("只有待支付状态的订单才能支付");
        }
        
        return orderService.updateOrderStatus(orderId, OrderStatus.PAID);
    }
    
    public Order payOrderByOrderNo(String orderNo) {
        Order order = orderService.getOrderByOrderNo(orderNo);
        
        if (order.getStatus() != OrderStatus.PENDING_PAYMENT) {
            throw new BusinessException("只有待支付状态的订单才能支付");
        }
        
        return orderService.updateOrderStatus(order.getId(), OrderStatus.PAID);
    }
}

package com.foodordering.controller;

import com.foodordering.dto.request.CreateOrderRequest;
import com.foodordering.dto.response.ApiResponse;
import com.foodordering.entity.Order;
import com.foodordering.service.OrderService;
import com.foodordering.service.PaymentService;
import jakarta.validation.Valid;
import lombok.RequiredArgsConstructor;
import org.springframework.web.bind.annotation.*;

import java.util.List;

@RestController
@RequestMapping("/api/orders")
@RequiredArgsConstructor
public class OrderController {
    
    private final OrderService orderService;
    private final PaymentService paymentService;
    
    @PostMapping
    public ApiResponse<Order> createOrder(@Valid @RequestBody CreateOrderRequest request) {
        return ApiResponse.success("订单创建成功", orderService.createOrder(request));
    }
    
    @GetMapping("/{id}")
    public ApiResponse<Order> getOrderById(@PathVariable Long id) {
        return ApiResponse.success(orderService.getOrderById(id));
    }
    
    @GetMapping("/order-no/{orderNo}")
    public ApiResponse<Order> getOrderByOrderNo(@PathVariable String orderNo) {
        return ApiResponse.success(orderService.getOrderByOrderNo(orderNo));
    }
    
    @GetMapping("/user/{userId}")
    public ApiResponse<List<Order>> getOrdersByUserId(@PathVariable Long userId) {
        return ApiResponse.success(orderService.getOrdersByUserId(userId));
    }
    
    @PostMapping("/{id}/cancel")
    public ApiResponse<Order> cancelOrder(@PathVariable Long id) {
        return ApiResponse.success("订单已取消", orderService.cancelOrder(id));
    }
    
    @PostMapping("/{id}/pay")
    public ApiResponse<Order> payOrder(@PathVariable Long id) {
        return ApiResponse.success("支付成功", paymentService.payOrder(id));
    }
}

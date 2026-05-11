package com.foodordering.controller;

import com.foodordering.dto.response.ApiResponse;
import com.foodordering.entity.Order;
import com.foodordering.enums.OrderStatus;
import com.foodordering.service.OrderService;
import lombok.RequiredArgsConstructor;
import org.springframework.web.bind.annotation.*;

import java.util.List;

@RestController
@RequestMapping("/api/admin/orders")
@RequiredArgsConstructor
public class AdminOrderController {
    
    private final OrderService orderService;
    
    @GetMapping
    public ApiResponse<List<Order>> getAllOrders() {
        return ApiResponse.success(orderService.getAllOrders());
    }
    
    @GetMapping("/{id}")
    public ApiResponse<Order> getOrderById(@PathVariable Long id) {
        return ApiResponse.success(orderService.getOrderById(id));
    }
    
    @PostMapping("/{id}/prepare")
    public ApiResponse<Order> startPreparing(@PathVariable Long id) {
        return ApiResponse.success("订单开始制作", 
                orderService.updateOrderStatus(id, OrderStatus.PREPARING));
    }
    
    @PostMapping("/{id}/complete")
    public ApiResponse<Order> completeOrder(@PathVariable Long id) {
        return ApiResponse.success("订单已完成", 
                orderService.updateOrderStatus(id, OrderStatus.COMPLETED));
    }
}

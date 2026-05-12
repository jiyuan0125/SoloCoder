package com.foodordering.service;

import com.foodordering.dto.request.CreateOrderRequest;
import com.foodordering.dto.request.OrderItemRequest;
import com.foodordering.entity.Dish;
import com.foodordering.entity.Order;
import com.foodordering.entity.OrderItem;
import com.foodordering.enums.OrderStatus;
import com.foodordering.enums.TasteOption;
import com.foodordering.exception.BusinessException;
import com.foodordering.repository.OrderRepository;
import lombok.RequiredArgsConstructor;
import org.springframework.stereotype.Service;

import java.time.LocalDateTime;
import java.time.format.DateTimeFormatter;
import java.util.ArrayList;
import java.util.List;
import java.util.UUID;

@Service
@RequiredArgsConstructor
public class OrderService {
    
    private static final long PAYMENT_TIMEOUT_MINUTES = 15;
    
    private final OrderRepository orderRepository;
    private final DishService dishService;
    
    public Order createOrder(CreateOrderRequest request) {
        validateOrderItems(request.getItems());
        
        List<OrderItem> orderItems = new ArrayList<>();
        List<Long> deductedDishIds = new ArrayList<>();
        List<Integer> deductedQuantities = new ArrayList<>();
        double totalPrice = 0.0;
        
        try {
            for (OrderItemRequest itemRequest : request.getItems()) {
                Dish dish = dishService.getDishById(itemRequest.getDishId());
                validateDishAvailability(dish, itemRequest.getQuantity());
                validateTasteOption(dish, itemRequest.getTaste());
                
                dishService.decreaseStock(dish.getId(), itemRequest.getQuantity());
                deductedDishIds.add(dish.getId());
                deductedQuantities.add(itemRequest.getQuantity());
                
                OrderItem orderItem = new OrderItem();
                orderItem.setDishId(dish.getId());
                orderItem.setDishName(dish.getName());
                orderItem.setDishPrice(dish.getPrice());
                orderItem.setQuantity(itemRequest.getQuantity());
                orderItem.setTaste(itemRequest.getTaste() != null ? 
                        itemRequest.getTaste() : TasteOption.NONE);
                orderItem.setSubtotal(dish.getPrice() * itemRequest.getQuantity());
                
                orderItems.add(orderItem);
                totalPrice += orderItem.getSubtotal();
            }
        } catch (Exception e) {
            for (int i = 0; i < deductedDishIds.size(); i++) {
                dishService.increaseStock(deductedDishIds.get(i), deductedQuantities.get(i));
            }
            throw e;
        }
        
        Order order = new Order();
        order.setOrderNo(generateOrderNo());
        order.setUserId(request.getUserId());
        order.setItems(orderItems);
        order.setTotalPrice(totalPrice);
        order.setStatus(OrderStatus.PENDING_PAYMENT);
        order.setCreatedAt(LocalDateTime.now());
        order.setUpdatedAt(LocalDateTime.now());
        order.setRemark(request.getRemark());
        
        return orderRepository.save(order);
    }
    
    public Order getOrderById(Long id) {
        Order order = orderRepository.findById(id)
                .orElseThrow(() -> new BusinessException("订单不存在"));
        checkAndCancelExpiredOrder(order);
        return order;
    }
    
    public Order getOrderByOrderNo(String orderNo) {
        Order order = orderRepository.findByOrderNo(orderNo)
                .orElseThrow(() -> new BusinessException("订单不存在"));
        checkAndCancelExpiredOrder(order);
        return order;
    }
    
    public List<Order> getOrdersByUserId(Long userId) {
        List<Order> orders = orderRepository.findByUserId(userId);
        orders.forEach(this::checkAndCancelExpiredOrder);
        return orders;
    }
    
    public List<Order> getAllOrders() {
        List<Order> orders = orderRepository.findAll();
        orders.forEach(this::checkAndCancelExpiredOrder);
        return orders;
    }
    
    public Order cancelOrder(Long id) {
        Order order = orderRepository.findById(id)
                .orElseThrow(() -> new BusinessException("订单不存在"));
        
        if (order.getStatus() != OrderStatus.PENDING_PAYMENT) {
            throw new BusinessException("只有待支付状态的订单才能取消");
        }
        
        return cancelOrderInternal(order);
    }
    
    public void checkAndCancelExpiredOrder(Order order) {
        if (order.getStatus() == OrderStatus.PENDING_PAYMENT) {
            LocalDateTime now = LocalDateTime.now();
            LocalDateTime expireTime = order.getCreatedAt().plusMinutes(PAYMENT_TIMEOUT_MINUTES);
            
            if (now.isAfter(expireTime)) {
                cancelOrderInternal(order);
            }
        }
    }
    
    private Order cancelOrderInternal(Order order) {
        for (OrderItem item : order.getItems()) {
            dishService.increaseStock(item.getDishId(), item.getQuantity());
        }
        
        order.setStatus(OrderStatus.CANCELLED);
        order.setUpdatedAt(LocalDateTime.now());
        return orderRepository.save(order);
    }
    
    public Order updateOrderStatus(Long id, OrderStatus status) {
        Order order = getOrderById(id);
        
        validateStatusTransition(order.getStatus(), status);
        
        order.setStatus(status);
        order.setUpdatedAt(LocalDateTime.now());
        return orderRepository.save(order);
    }
    
    private void validateOrderItems(List<OrderItemRequest> items) {
        if (items == null || items.isEmpty()) {
            throw new BusinessException("订单不能为空");
        }
    }
    
    private void validateDishAvailability(Dish dish, Integer quantity) {
        if (!dish.getIsActive()) {
            throw new BusinessException("菜品【" + dish.getName() + "】已下架");
        }
        if (dish.getStock() < quantity) {
            throw new BusinessException("菜品【" + dish.getName() + "】库存不足，当前库存: " + dish.getStock());
        }
    }
    
    private void validateTasteOption(Dish dish, TasteOption taste) {
        if (taste != null && !dish.getAvailableTastes().contains(taste)) {
            throw new BusinessException("菜品【" + dish.getName() + "】不支持该口味: " + taste.getDisplayName());
        }
    }
    
    private void validateStatusTransition(OrderStatus currentStatus, OrderStatus newStatus) {
        switch (newStatus) {
            case PAID:
                if (currentStatus != OrderStatus.PENDING_PAYMENT) {
                    throw new BusinessException("订单状态不允许支付");
                }
                break;
            case PREPARING:
                if (currentStatus != OrderStatus.PAID) {
                    throw new BusinessException("订单状态不允许开始制作");
                }
                break;
            case COMPLETED:
                if (currentStatus != OrderStatus.PREPARING) {
                    throw new BusinessException("订单状态不允许完成");
                }
                break;
            case CANCELLED:
                if (currentStatus != OrderStatus.PENDING_PAYMENT) {
                    throw new BusinessException("只有待支付状态的订单才能取消");
                }
                break;
            default:
                throw new BusinessException("不支持的订单状态变更");
        }
    }
    
    private String generateOrderNo() {
        String datePart = LocalDateTime.now().format(DateTimeFormatter.ofPattern("yyyyMMddHHmmss"));
        String randomPart = UUID.randomUUID().toString().substring(0, 4).toUpperCase();
        return "ORD" + datePart + randomPart;
    }
}

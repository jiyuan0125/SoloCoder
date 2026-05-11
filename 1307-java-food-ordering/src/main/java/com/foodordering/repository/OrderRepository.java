package com.foodordering.repository;

import com.foodordering.entity.Order;
import com.foodordering.enums.OrderStatus;
import org.springframework.stereotype.Repository;

import java.util.*;
import java.util.concurrent.ConcurrentHashMap;
import java.util.concurrent.atomic.AtomicLong;
import java.util.stream.Collectors;

@Repository
public class OrderRepository {
    private final Map<Long, Order> orders = new ConcurrentHashMap<>();
    private final AtomicLong idGenerator = new AtomicLong(0);
    
    public Order save(Order order) {
        if (order.getId() == null) {
            order.setId(idGenerator.incrementAndGet());
        }
        orders.put(order.getId(), order);
        return order;
    }
    
    public Optional<Order> findById(Long id) {
        return Optional.ofNullable(orders.get(id));
    }
    
    public Optional<Order> findByOrderNo(String orderNo) {
        return orders.values().stream()
                .filter(order -> order.getOrderNo().equals(orderNo))
                .findFirst();
    }
    
    public List<Order> findAll() {
        return new ArrayList<>(orders.values());
    }
    
    public List<Order> findByUserId(Long userId) {
        return orders.values().stream()
                .filter(order -> order.getUserId().equals(userId))
                .sorted(Comparator.comparing(Order::getCreatedAt).reversed())
                .collect(Collectors.toList());
    }
    
    public List<Order> findByStatus(OrderStatus status) {
        return orders.values().stream()
                .filter(order -> order.getStatus().equals(status))
                .toList();
    }
    
    public void deleteById(Long id) {
        orders.remove(id);
    }
    
    public boolean existsById(Long id) {
        return orders.containsKey(id);
    }
}

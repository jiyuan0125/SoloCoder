package com.foodordering.repository;

import com.foodordering.entity.Dish;
import org.springframework.stereotype.Repository;

import java.util.*;
import java.util.concurrent.ConcurrentHashMap;
import java.util.concurrent.atomic.AtomicLong;

@Repository
public class DishRepository {
    private final Map<Long, Dish> dishes = new ConcurrentHashMap<>();
    private final AtomicLong idGenerator = new AtomicLong(0);
    
    public Dish save(Dish dish) {
        if (dish.getId() == null) {
            dish.setId(idGenerator.incrementAndGet());
        }
        dishes.put(dish.getId(), dish);
        return dish;
    }
    
    public Optional<Dish> findById(Long id) {
        return Optional.ofNullable(dishes.get(id));
    }
    
    public List<Dish> findAll() {
        return new ArrayList<>(dishes.values());
    }
    
    public List<Dish> findByIsActiveTrue() {
        return dishes.values().stream()
                .filter(Dish::getIsActive)
                .filter(dish -> dish.getStock() != null && dish.getStock() > 0)
                .toList();
    }
    
    public void deleteById(Long id) {
        dishes.remove(id);
    }
    
    public boolean existsById(Long id) {
        return dishes.containsKey(id);
    }
}

package com.foodordering.service;

import com.foodordering.dto.request.CreateDishRequest;
import com.foodordering.dto.request.UpdateDishRequest;
import com.foodordering.entity.Dish;
import com.foodordering.enums.TasteOption;
import com.foodordering.exception.BusinessException;
import com.foodordering.repository.DishRepository;
import lombok.RequiredArgsConstructor;
import org.springframework.stereotype.Service;

import java.time.LocalDateTime;
import java.util.List;

@Service
@RequiredArgsConstructor
public class DishService {
    
    private final DishRepository dishRepository;
    
    public Dish createDish(CreateDishRequest request) {
        Dish dish = new Dish();
        dish.setName(request.getName());
        dish.setPrice(request.getPrice());
        dish.setCategory(request.getCategory());
        dish.setAvailableTastes(request.getAvailableTastes() != null ? 
                request.getAvailableTastes() : List.of(TasteOption.NONE));
        dish.setStock(request.getStock());
        dish.setIsActive(true);
        dish.setCreatedAt(LocalDateTime.now());
        dish.setUpdatedAt(LocalDateTime.now());
        return dishRepository.save(dish);
    }
    
    public Dish updateDish(Long id, UpdateDishRequest request) {
        Dish dish = dishRepository.findById(id)
                .orElseThrow(() -> new BusinessException("菜品不存在"));
        
        if (request.getName() != null) {
            dish.setName(request.getName());
        }
        if (request.getPrice() != null) {
            dish.setPrice(request.getPrice());
        }
        if (request.getCategory() != null) {
            dish.setCategory(request.getCategory());
        }
        if (request.getAvailableTastes() != null) {
            dish.setAvailableTastes(request.getAvailableTastes());
        }
        if (request.getStock() != null) {
            dish.setStock(request.getStock());
        }
        if (request.getIsActive() != null) {
            dish.setIsActive(request.getIsActive());
        }
        
        dish.setUpdatedAt(LocalDateTime.now());
        return dishRepository.save(dish);
    }
    
    public Dish getDishById(Long id) {
        return dishRepository.findById(id)
                .orElseThrow(() -> new BusinessException("菜品不存在"));
    }
    
    public List<Dish> getAllDishes() {
        return dishRepository.findAll();
    }
    
    public List<Dish> getAvailableDishes() {
        return dishRepository.findByIsActiveTrue();
    }
    
    public void decreaseStock(Long dishId, Integer quantity) {
        Dish dish = getDishById(dishId);
        if (dish.getStock() < quantity) {
            throw new BusinessException("菜品【" + dish.getName() + "】库存不足");
        }
        dish.setStock(dish.getStock() - quantity);
        dish.setUpdatedAt(LocalDateTime.now());
        dishRepository.save(dish);
    }
    
    public void increaseStock(Long dishId, Integer quantity) {
        Dish dish = getDishById(dishId);
        dish.setStock(dish.getStock() + quantity);
        dish.setUpdatedAt(LocalDateTime.now());
        dishRepository.save(dish);
    }
    
    public void setDishActive(Long id, boolean active) {
        Dish dish = getDishById(id);
        dish.setIsActive(active);
        dish.setUpdatedAt(LocalDateTime.now());
        dishRepository.save(dish);
    }
    
    public void deleteDish(Long id) {
        if (!dishRepository.existsById(id)) {
            throw new BusinessException("菜品不存在");
        }
        dishRepository.deleteById(id);
    }
}

package com.foodordering.controller;

import com.foodordering.dto.request.CreateDishRequest;
import com.foodordering.dto.request.UpdateDishRequest;
import com.foodordering.dto.response.ApiResponse;
import com.foodordering.entity.Dish;
import com.foodordering.service.DishService;
import jakarta.validation.Valid;
import lombok.RequiredArgsConstructor;
import org.springframework.web.bind.annotation.*;

import java.util.List;

@RestController
@RequestMapping("/api/dishes")
@RequiredArgsConstructor
public class DishController {
    
    private final DishService dishService;
    
    @GetMapping("/available")
    public ApiResponse<List<Dish>> getAvailableDishes() {
        return ApiResponse.success(dishService.getAvailableDishes());
    }
    
    @GetMapping("/{id}")
    public ApiResponse<Dish> getDishById(@PathVariable Long id) {
        return ApiResponse.success(dishService.getDishById(id));
    }
}

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
@RequestMapping("/api/admin/dishes")
@RequiredArgsConstructor
public class AdminDishController {
    
    private final DishService dishService;
    
    @PostMapping
    public ApiResponse<Dish> createDish(@Valid @RequestBody CreateDishRequest request) {
        return ApiResponse.success("菜品创建成功", dishService.createDish(request));
    }
    
    @PutMapping("/{id}")
    public ApiResponse<Dish> updateDish(
            @PathVariable Long id,
            @Valid @RequestBody UpdateDishRequest request) {
        return ApiResponse.success("菜品更新成功", dishService.updateDish(id, request));
    }
    
    @GetMapping
    public ApiResponse<List<Dish>> getAllDishes() {
        return ApiResponse.success(dishService.getAllDishes());
    }
    
    @GetMapping("/{id}")
    public ApiResponse<Dish> getDishById(@PathVariable Long id) {
        return ApiResponse.success(dishService.getDishById(id));
    }
    
    @PutMapping("/{id}/activate")
    public ApiResponse<Void> activateDish(@PathVariable Long id) {
        dishService.setDishActive(id, true);
        return ApiResponse.success("菜品已上架", null);
    }
    
    @PutMapping("/{id}/deactivate")
    public ApiResponse<Void> deactivateDish(@PathVariable Long id) {
        dishService.setDishActive(id, false);
        return ApiResponse.success("菜品已下架", null);
    }
    
    @DeleteMapping("/{id}")
    public ApiResponse<Void> deleteDish(@PathVariable Long id) {
        dishService.deleteDish(id);
        return ApiResponse.success("菜品删除成功", null);
    }
}

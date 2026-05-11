package com.foodordering.config;

import com.foodordering.dto.request.CreateDishRequest;
import com.foodordering.enums.Category;
import com.foodordering.enums.TasteOption;
import com.foodordering.service.DishService;
import jakarta.annotation.PostConstruct;
import lombok.RequiredArgsConstructor;
import org.springframework.stereotype.Component;

import java.util.List;

@Component
@RequiredArgsConstructor
public class DataInitializer {
    
    private final DishService dishService;
    
    @PostConstruct
    public void initData() {
        if (!dishService.getAllDishes().isEmpty()) {
            return;
        }
        
        CreateDishRequest beefNoodle = new CreateDishRequest();
        beefNoodle.setName("牛肉面");
        beefNoodle.setPrice(28.0);
        beefNoodle.setCategory(Category.STAPLE);
        beefNoodle.setAvailableTastes(List.of(TasteOption.SPICY_LIGHT, TasteOption.SPICY_MEDIUM, TasteOption.SPICY_HEAVY));
        beefNoodle.setStock(10);
        dishService.createDish(beefNoodle);
        
        CreateDishRequest chickenRice = new CreateDishRequest();
        chickenRice.setName("鸡肉饭");
        chickenRice.setPrice(22.0);
        chickenRice.setCategory(Category.STAPLE);
        chickenRice.setAvailableTastes(List.of(TasteOption.NONE));
        chickenRice.setStock(15);
        dishService.createDish(chickenRice);
        
        CreateDishRequest fries = new CreateDishRequest();
        fries.setName("薯条");
        fries.setPrice(12.0);
        fries.setCategory(Category.SNACK);
        fries.setAvailableTastes(List.of(TasteOption.NONE));
        fries.setStock(30);
        dishService.createDish(fries);
        
        CreateDishRequest chickenWings = new CreateDishRequest();
        chickenWings.setName("鸡翅");
        chickenWings.setPrice(15.0);
        chickenWings.setCategory(Category.SNACK);
        chickenWings.setAvailableTastes(List.of(TasteOption.SPICY_LIGHT, TasteOption.SPICY_MEDIUM));
        chickenWings.setStock(20);
        dishService.createDish(chickenWings);
        
        CreateDishRequest cola = new CreateDishRequest();
        cola.setName("可乐");
        cola.setPrice(8.0);
        cola.setCategory(Category.DRINK);
        cola.setAvailableTastes(List.of(TasteOption.NONE));
        cola.setStock(50);
        dishService.createDish(cola);
        
        CreateDishRequest milkTea = new CreateDishRequest();
        milkTea.setName("奶茶");
        milkTea.setPrice(18.0);
        milkTea.setCategory(Category.DRINK);
        milkTea.setAvailableTastes(List.of(TasteOption.SUGAR_LESS, TasteOption.SUGAR_NORMAL, TasteOption.SUGAR_MORE));
        milkTea.setStock(25);
        dishService.createDish(milkTea);
        
        CreateDishRequest cake = new CreateDishRequest();
        cake.setName("蛋糕");
        cake.setPrice(25.0);
        cake.setCategory(Category.DESSERT);
        cake.setAvailableTastes(List.of(TasteOption.NONE));
        cake.setStock(5);
        dishService.createDish(cake);
        
        CreateDishRequest iceCream = new CreateDishRequest();
        iceCream.setName("冰淇淋");
        iceCream.setPrice(10.0);
        iceCream.setCategory(Category.DESSERT);
        iceCream.setAvailableTastes(List.of(TasteOption.NONE));
        iceCream.setStock(0);
        dishService.createDish(iceCream);
    }
}

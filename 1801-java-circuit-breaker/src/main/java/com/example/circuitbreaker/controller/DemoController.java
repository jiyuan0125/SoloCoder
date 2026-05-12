package com.example.circuitbreaker.controller;

import com.example.circuitbreaker.service.DownstreamServiceCaller;
import lombok.RequiredArgsConstructor;
import org.springframework.web.bind.annotation.*;

@RestController
@RequestMapping("/demo")
@RequiredArgsConstructor
public class DemoController {
    private final DownstreamServiceCaller serviceCaller;

    @GetMapping("/payment/{orderId}")
    public String callPayment(@PathVariable String orderId) {
        return serviceCaller.callPaymentService(orderId);
    }

    @GetMapping("/inventory/{productId}")
    public String callInventory(@PathVariable String productId) {
        return serviceCaller.callInventoryService(productId);
    }
}

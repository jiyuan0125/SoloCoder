package com.example.memorymq.controller;

import com.example.memorymq.dto.ApiResponse;
import com.example.memorymq.dto.PublishRequest;
import com.example.memorymq.model.Message;
import com.example.memorymq.service.MessageEngine;
import jakarta.validation.Valid;
import lombok.RequiredArgsConstructor;
import org.springframework.web.bind.annotation.*;

import java.util.HashMap;
import java.util.Map;

@RestController
@RequestMapping("/api/producer")
@RequiredArgsConstructor
public class ProducerController {
    private final MessageEngine engine;

    @PostMapping("/publish")
    public ApiResponse<Map<String, Object>> publish(@Valid @RequestBody PublishRequest request) {
        if (!engine.topicExists(request.getTopic())) {
            boolean created = engine.createTopic(request.getTopic(), null);
            if (!created) {
                return ApiResponse.error("Failed to create topic: " + request.getTopic());
            }
        }
        Message message = Message.builder()
                .topic(request.getTopic())
                .payload(request.getPayload())
                .build();
        boolean success = engine.publish(request.getTopic(), message);
        if (success) {
            Map<String, Object> data = new HashMap<>();
            data.put("messageId", message.getId());
            data.put("topic", request.getTopic());
            return ApiResponse.success("Message published", data);
        }
        return ApiResponse.error("Failed to publish message (queue full)");
    }
}

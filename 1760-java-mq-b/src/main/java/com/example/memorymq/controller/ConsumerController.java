package com.example.memorymq.controller;

import com.example.memorymq.dto.AckRequest;
import com.example.memorymq.dto.ApiResponse;
import com.example.memorymq.dto.ConsumeRequest;
import com.example.memorymq.dto.NackRequest;
import com.example.memorymq.model.Message;
import com.example.memorymq.service.MessageEngine;
import jakarta.validation.Valid;
import lombok.RequiredArgsConstructor;
import org.springframework.web.bind.annotation.*;

import java.util.List;

@RestController
@RequestMapping("/api/consumer")
@RequiredArgsConstructor
public class ConsumerController {
    private final MessageEngine engine;

    @PostMapping("/consume")
    public ApiResponse<List<Message>> consume(@Valid @RequestBody ConsumeRequest request) {
        if (!engine.topicExists(request.getTopic())) {
            return ApiResponse.error("Topic not found: " + request.getTopic());
        }
        int maxMessages = (request.getMaxMessages() != null && request.getMaxMessages() > 0)
            ? request.getMaxMessages() : 10;
        int timeoutSeconds = (request.getTimeoutSeconds() != null && request.getTimeoutSeconds() > 0)
            ? request.getTimeoutSeconds() : 5;
        List<Message> messages = engine.consume(
            request.getTopic(),
            request.getConsumerGroup(),
            maxMessages,
            timeoutSeconds
        );
        return ApiResponse.success(messages);
    }

    @PostMapping("/ack")
    public ApiResponse<Void> ack(@Valid @RequestBody AckRequest request) {
        boolean success = engine.ack(request.getTopic(), request.getConsumerGroup(), request.getMessageIds());
        if (success) {
            return ApiResponse.success("Messages acknowledged", null);
        }
        return ApiResponse.error("Failed to acknowledge messages");
    }

    @PostMapping("/nack")
    public ApiResponse<Void> nack(@Valid @RequestBody NackRequest request) {
        boolean requeue = request.getRequeue() != null ? request.getRequeue() : true;
        boolean success = engine.nack(request.getTopic(), request.getConsumerGroup(), request.getMessageIds(), requeue);
        if (success) {
            return ApiResponse.success("Messages negatively acknowledged", null);
        }
        return ApiResponse.error("Failed to negatively acknowledge messages");
    }
}

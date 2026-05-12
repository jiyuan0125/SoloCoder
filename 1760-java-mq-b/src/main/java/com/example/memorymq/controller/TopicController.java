package com.example.memorymq.controller;

import com.example.memorymq.dto.ApiResponse;
import com.example.memorymq.dto.CreateTopicRequest;
import com.example.memorymq.model.TopicInfo;
import com.example.memorymq.service.MessageEngine;
import jakarta.validation.Valid;
import lombok.RequiredArgsConstructor;
import org.springframework.web.bind.annotation.*;

import java.util.List;

@RestController
@RequestMapping("/api/topics")
@RequiredArgsConstructor
public class TopicController {
    private final MessageEngine engine;

    @PostMapping
    public ApiResponse<Void> createTopic(@Valid @RequestBody CreateTopicRequest request) {
        boolean created = engine.createTopic(request.getName(), request.getMaxMessages());
        if (created) {
            return ApiResponse.success("Topic created: " + request.getName(), null);
        }
        return ApiResponse.error("Topic already exists: " + request.getName());
    }

    @DeleteMapping("/{name}")
    public ApiResponse<Void> deleteTopic(@PathVariable String name) {
        boolean deleted = engine.deleteTopic(name);
        if (deleted) {
            return ApiResponse.success("Topic deleted: " + name, null);
        }
        return ApiResponse.error("Topic not found: " + name);
    }

    @GetMapping
    public ApiResponse<List<TopicInfo>> listTopics() {
        List<TopicInfo> topics = engine.listTopics();
        return ApiResponse.success(topics);
    }
}

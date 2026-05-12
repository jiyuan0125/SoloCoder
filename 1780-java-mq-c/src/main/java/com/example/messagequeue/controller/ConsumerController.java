package com.example.messagequeue.controller;

import com.example.messagequeue.model.CommitOffsetRequest;
import com.example.messagequeue.model.Message;
import com.example.messagequeue.service.MessageQueueService;
import jakarta.validation.Valid;
import lombok.RequiredArgsConstructor;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import java.util.HashMap;
import java.util.Map;

@RestController
@RequestMapping("/api/topics/{topic}")
@RequiredArgsConstructor
public class ConsumerController {

    private final MessageQueueService queueService;

    @GetMapping("/messages")
    public ResponseEntity<Message> receive(@PathVariable String topic) {
        Message message = queueService.receive(topic);
        if (message == null) {
            return ResponseEntity.noContent().build();
        }
        return ResponseEntity.ok(message);
    }

    @PostMapping("/offsets")
    public ResponseEntity<Map<String, Object>> commitOffset(
            @PathVariable String topic,
            @Valid @RequestBody CommitOffsetRequest request) {

        queueService.commitOffset(topic, request.getOffset(), request.getConsumerId());

        Map<String, Object> response = new HashMap<>();
        response.put("topic", topic);
        response.put("offset", request.getOffset());
        response.put("consumerId", request.getConsumerId() != null ? request.getConsumerId() : "default");
        response.put("committed", true);

        return ResponseEntity.ok(response);
    }

    @GetMapping("/offsets")
    public ResponseEntity<Map<String, Object>> getLastOffset(
            @PathVariable String topic,
            @RequestParam(required = false) String consumerId) {

        long offset = queueService.getLastCommittedOffset(topic, consumerId);

        Map<String, Object> response = new HashMap<>();
        response.put("topic", topic);
        response.put("consumerId", consumerId != null ? consumerId : "default");
        response.put("offset", offset);

        return ResponseEntity.ok(response);
    }
}

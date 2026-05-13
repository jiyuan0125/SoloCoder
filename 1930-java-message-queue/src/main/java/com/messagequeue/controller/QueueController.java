package com.messagequeue.controller;

import com.messagequeue.dto.*;
import com.messagequeue.model.Message;
import com.messagequeue.model.Subscription;
import com.messagequeue.service.QueueService;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import java.util.List;

@Slf4j
@RestController
@RequestMapping("/queues")
@RequiredArgsConstructor
public class QueueController {

    private final QueueService queueService;

    @PostMapping("/{name}/subscribe")
    public ResponseEntity<ApiResponse<Subscription>> subscribe(
            @PathVariable String name,
            @RequestBody SubscribeRequest request) {
        try {
            if (request.getGroupId() == null || request.getGroupId().trim().isEmpty()) {
                return ResponseEntity.badRequest()
                        .body(ApiResponse.error("groupId is required"));
            }
            Subscription subscription = queueService.subscribe(name, request.getGroupId(), request.getFilter());
            return ResponseEntity.ok(ApiResponse.success("Subscription created/updated", subscription));
        } catch (IllegalArgumentException e) {
            return ResponseEntity.badRequest().body(ApiResponse.error(e.getMessage()));
        } catch (Exception e) {
            log.error("Failed to subscribe", e);
            return ResponseEntity.internalServerError().body(ApiResponse.error("Internal server error"));
        }
    }

    @DeleteMapping("/{name}/subscribe")
    public ResponseEntity<ApiResponse<Void>> unsubscribe(
            @PathVariable String name,
            @RequestParam String groupId) {
        try {
            boolean removed = queueService.unsubscribe(name, groupId);
            if (removed) {
                return ResponseEntity.ok(ApiResponse.success("Subscription removed", null));
            } else {
                return ResponseEntity.notFound().build();
            }
        } catch (Exception e) {
            log.error("Failed to unsubscribe", e);
            return ResponseEntity.internalServerError().body(ApiResponse.error("Internal server error"));
        }
    }

    @GetMapping("/{name}/subscriptions")
    public ResponseEntity<ApiResponse<List<Subscription>>> listSubscriptions(@PathVariable String name) {
        try {
            List<Subscription> subscriptions = queueService.listSubscriptions(name);
            return ResponseEntity.ok(ApiResponse.success(subscriptions));
        } catch (Exception e) {
            log.error("Failed to list subscriptions", e);
            return ResponseEntity.internalServerError().body(ApiResponse.error("Internal server error"));
        }
    }

    @PostMapping("/{name}/messages")
    public ResponseEntity<ApiResponse<Void>> publish(
            @PathVariable String name,
            @RequestBody PublishRequest request) {
        try {
            if (request.getBody() == null) {
                return ResponseEntity.badRequest().body(ApiResponse.error("body is required"));
            }
            queueService.publish(name, request.getHeaders(), request.getBody());
            return ResponseEntity.ok(ApiResponse.success("Message published", null));
        } catch (Exception e) {
            log.error("Failed to publish message", e);
            return ResponseEntity.internalServerError().body(ApiResponse.error("Internal server error"));
        }
    }

    @PostMapping("/{name}/consume")
    public ResponseEntity<ApiResponse<Message>> consume(
            @PathVariable String name,
            @RequestBody ConsumeRequest request) {
        try {
            if (request.getGroupId() == null || request.getGroupId().trim().isEmpty()) {
                return ResponseEntity.badRequest().body(ApiResponse.error("groupId is required"));
            }
            Message message = queueService.consume(name, request.getGroupId());
            if (message == null) {
                return ResponseEntity.ok(ApiResponse.success("No messages available", null));
            }
            return ResponseEntity.ok(ApiResponse.success(message));
        } catch (IllegalArgumentException e) {
            return ResponseEntity.badRequest().body(ApiResponse.error(e.getMessage()));
        } catch (Exception e) {
            log.error("Failed to consume message", e);
            return ResponseEntity.internalServerError().body(ApiResponse.error("Internal server error"));
        }
    }

    @PostMapping("/{name}/ack")
    public ResponseEntity<ApiResponse<Void>> acknowledge(
            @PathVariable String name,
            @RequestBody AckRequest request) {
        try {
            if (request.getGroupId() == null || request.getMessageId() == null) {
                return ResponseEntity.badRequest().body(ApiResponse.error("groupId and messageId are required"));
            }
            boolean acknowledged = queueService.acknowledge(name, request.getGroupId(), request.getMessageId());
            if (acknowledged) {
                return ResponseEntity.ok(ApiResponse.success("Message acknowledged", null));
            } else {
                return ResponseEntity.status(404).body(ApiResponse.error("Message not found"));
            }
        } catch (Exception e) {
            log.error("Failed to acknowledge message", e);
            return ResponseEntity.internalServerError().body(ApiResponse.error("Internal server error"));
        }
    }

    @PostMapping("/{name}/nack")
    public ResponseEntity<ApiResponse<Void>> nack(
            @PathVariable String name,
            @RequestBody AckRequest request) {
        try {
            if (request.getGroupId() == null || request.getMessageId() == null) {
                return ResponseEntity.badRequest().body(ApiResponse.error("groupId and messageId are required"));
            }
            boolean nacked = queueService.nack(name, request.getGroupId(), request.getMessageId());
            if (nacked) {
                return ResponseEntity.ok(ApiResponse.success("Message rejected", null));
            } else {
                return ResponseEntity.status(404).body(ApiResponse.error("Message not found"));
            }
        } catch (Exception e) {
            log.error("Failed to nack message", e);
            return ResponseEntity.internalServerError().body(ApiResponse.error("Internal server error"));
        }
    }

    @GetMapping("/{name}/dead-letters")
    public ResponseEntity<ApiResponse<List<Message>>> getDeadLetters(@PathVariable String name) {
        try {
            List<Message> deadLetters = queueService.getDeadLetters(name);
            return ResponseEntity.ok(ApiResponse.success(deadLetters));
        } catch (Exception e) {
            log.error("Failed to get dead letters", e);
            return ResponseEntity.internalServerError().body(ApiResponse.error("Internal server error"));
        }
    }

    @PostMapping("/{name}/dead-letters/consume")
    public ResponseEntity<ApiResponse<Message>> consumeDeadLetter(
            @PathVariable String name,
            @RequestBody ConsumeRequest request) {
        try {
            if (request.getGroupId() == null || request.getGroupId().trim().isEmpty()) {
                return ResponseEntity.badRequest().body(ApiResponse.error("groupId is required"));
            }
            Message message = queueService.consumeDeadLetter(name, request.getGroupId());
            if (message == null) {
                return ResponseEntity.ok(ApiResponse.success("No dead letters available", null));
            }
            return ResponseEntity.ok(ApiResponse.success(message));
        } catch (IllegalArgumentException e) {
            return ResponseEntity.badRequest().body(ApiResponse.error(e.getMessage()));
        } catch (Exception e) {
            log.error("Failed to consume dead letter", e);
            return ResponseEntity.internalServerError().body(ApiResponse.error("Internal server error"));
        }
    }
}

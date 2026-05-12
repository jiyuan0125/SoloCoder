package com.example.messagequeue.controller;

import com.example.messagequeue.model.Message;
import com.example.messagequeue.model.SendMessageRequest;
import com.example.messagequeue.service.MessageQueueService;
import com.example.messagequeue.service.MessageValidationService;
import jakarta.validation.Valid;
import lombok.RequiredArgsConstructor;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;
import reactor.core.publisher.Mono;

@RestController
@RequestMapping("/api/topics/{topic}/messages")
@RequiredArgsConstructor
public class ProducerController {

    private final MessageValidationService validationService;
    private final MessageQueueService queueService;

    @PostMapping
    public Mono<ResponseEntity<Message>> send(
            @PathVariable String topic,
            @Valid @RequestBody SendMessageRequest request) {

        validationService.validateMessage(request.getContent(), request.getContentType());

        return queueService.send(topic, request.getContent(), request.getContentType())
                .map(message -> ResponseEntity.ok(message));
    }
}

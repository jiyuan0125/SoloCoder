package com.solocoder.mq.controller;

import com.fasterxml.jackson.annotation.JsonProperty;
import com.solocoder.mq.model.Message;
import com.solocoder.mq.service.QueueService;
import lombok.Data;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import javax.validation.Valid;
import javax.validation.constraints.NotBlank;
import javax.validation.constraints.Min;

@RestController
@RequestMapping("/topics")
public class ProducerController {

    private final QueueService queueService;

    public ProducerController(QueueService queueService) {
        this.queueService = queueService;
    }

    @PostMapping("/{name}/messages")
    public ResponseEntity<MessageResponse> sendMessage(
            @PathVariable("name") String topicName,
            @Valid @RequestBody SendMessageRequest request) {
        
        Message msg = queueService.sendMessage(
                topicName, 
                request.getContent(), 
                request.getDelayMs() != null ? request.getDelayMs() : 0);
        
        return ResponseEntity.ok(MessageResponse.fromMessage(msg));
    }

    @Data
    public static class SendMessageRequest {
        @NotBlank(message = "content is required")
        @JsonProperty("content")
        private String content;

        @Min(value = 0, message = "delay_ms must be >= 0")
        @JsonProperty("delay_ms")
        private Long delayMs = 0L;
    }

    @Data
    public static class MessageResponse {
        private String id;
        private long offset;
        private String topic;
        private String status;
        private long delay_ms;
        private long created_at;
        private long ready_at;

        public static MessageResponse fromMessage(Message msg) {
            MessageResponse resp = new MessageResponse();
            resp.setId(msg.getId());
            resp.setOffset(msg.getOffset());
            resp.setTopic(msg.getTopic());
            resp.setStatus(msg.getStatus().name());
            resp.setDelay_ms(msg.getDelayMs());
            resp.setCreated_at(msg.getCreatedAt());
            resp.setReady_at(msg.getReadyAt());
            return resp;
        }
    }
}

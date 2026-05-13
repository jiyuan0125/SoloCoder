package com.solocoder.mq.controller;

import com.fasterxml.jackson.annotation.JsonProperty;
import com.solocoder.mq.model.Message;
import com.solocoder.mq.service.QueueService;
import lombok.Data;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

@RestController
@RequestMapping("/topics/{name}/consumers")
public class ConsumerController {

    private final QueueService queueService;

    public ConsumerController(QueueService queueService) {
        this.queueService = queueService;
    }

    @PostMapping("/groups/{groupId}/poll")
    public ResponseEntity<PollResponse> poll(
            @PathVariable("name") String topicName,
            @PathVariable("groupId") String groupId,
            @RequestParam(value = "consumer_id", defaultValue = "default") String consumerId) {
        
        Message msg = queueService.poll(topicName, groupId, consumerId);
        
        PollResponse resp = new PollResponse();
        resp.setTopic(topicName);
        resp.setGroup_id(groupId);
        
        if (msg != null) {
            resp.setHas_message(true);
            resp.setMessage(ConsumerMessage.fromMessage(msg));
        } else {
            resp.setHas_message(false);
        }
        
        return ResponseEntity.ok(resp);
    }

    @PostMapping("/groups/{groupId}/ack")
    public ResponseEntity<AckResponse> ack(
            @PathVariable("name") String topicName,
            @PathVariable("groupId") String groupId,
            @RequestParam(value = "consumer_id", defaultValue = "default") String consumerId,
            @RequestBody AckRequest request) {
        
        boolean success = queueService.ack(topicName, groupId, consumerId, request.getMessage_id());
        
        AckResponse resp = new AckResponse();
        resp.setSuccess(success);
        if (success) {
            resp.setMessage("ACK successful");
        } else {
            resp.setMessage("ACK failed - message not found or not in consuming state");
        }
        
        return ResponseEntity.ok(resp);
    }

    @Data
    public static class PollResponse {
        @JsonProperty("topic")
        private String topic;
        
        @JsonProperty("group_id")
        private String group_id;
        
        @JsonProperty("has_message")
        private boolean has_message;
        
        @JsonProperty("message")
        private ConsumerMessage message;
    }

    @Data
    public static class ConsumerMessage {
        @JsonProperty("id")
        private String id;
        
        @JsonProperty("offset")
        private long offset;
        
        @JsonProperty("content")
        private String content;
        
        @JsonProperty("delay_ms")
        private long delay_ms;
        
        @JsonProperty("created_at")
        private long created_at;

        public static ConsumerMessage fromMessage(Message msg) {
            ConsumerMessage cm = new ConsumerMessage();
            cm.setId(msg.getId());
            cm.setOffset(msg.getOffset());
            cm.setContent(msg.getContent());
            cm.setDelay_ms(msg.getDelayMs());
            cm.setCreated_at(msg.getCreatedAt());
            return cm;
        }
    }

    @Data
    public static class AckRequest {
        @JsonProperty("message_id")
        private String message_id;
    }

    @Data
    public static class AckResponse {
        @JsonProperty("success")
        private boolean success;
        
        @JsonProperty("message")
        private String message;
    }
}

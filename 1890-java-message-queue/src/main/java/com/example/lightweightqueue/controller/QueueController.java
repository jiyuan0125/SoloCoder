package com.example.lightweightqueue.controller;

import com.example.lightweightqueue.dto.AckRequest;
import com.example.lightweightqueue.dto.ConsumeRequest;
import com.example.lightweightqueue.dto.ConsumeResponse;
import com.example.lightweightqueue.dto.CreateTopicRequest;
import com.example.lightweightqueue.dto.ProduceRequest;
import com.example.lightweightqueue.dto.ProduceResponse;
import com.example.lightweightqueue.dto.ResendDeadLetterRequest;
import com.example.lightweightqueue.dto.StatsResponse;
import com.example.lightweightqueue.model.ConsumerGroup;
import com.example.lightweightqueue.model.Message;
import com.example.lightweightqueue.model.MessageStatus;
import com.example.lightweightqueue.model.Topic;
import com.example.lightweightqueue.service.MessageQueueService;
import com.example.lightweightqueue.service.TopicManager;
import jakarta.validation.Valid;
import lombok.RequiredArgsConstructor;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;
import java.util.Map;
import java.util.Optional;

@RestController
@RequestMapping("/api/queue")
@RequiredArgsConstructor
public class QueueController {

    private final MessageQueueService messageQueueService;
    private final TopicManager topicManager;

    @PostMapping("/topic")
    public ResponseEntity<Map<String, String>> createTopic(@Valid @RequestBody CreateTopicRequest request) {
        Topic topic = topicManager.createTopic(request.getName(), request.getMaxCapacity());
        Map<String, String> response = new HashMap<>();
        response.put("name", topic.getName());
        response.put("maxCapacity", String.valueOf(topic.getMaxCapacity()));
        return ResponseEntity.ok(response);
    }

    @PostMapping("/produce")
    public ResponseEntity<ProduceResponse> produce(@Valid @RequestBody ProduceRequest request) {
        try {
            String messageId = messageQueueService.produce(request.getTopic(), request.getBody(), request.getFormat());
            return ResponseEntity.ok(ProduceResponse.builder()
                    .success(true)
                    .messageId(messageId)
                    .build());
        } catch (IllegalArgumentException e) {
            return ResponseEntity.badRequest().body(ProduceResponse.builder()
                    .success(false)
                    .error(e.getMessage())
                    .build());
        }
    }

    @PostMapping("/consume")
    public ResponseEntity<ConsumeResponse> consume(@Valid @RequestBody ConsumeRequest request) {
        Optional<Message> message = messageQueueService.consume(
                request.getTopic(), request.getGroupId(), request.getConsumerId());
        
        if (message.isEmpty()) {
            Topic topic = topicManager.getTopic(request.getTopic());
            if (topic == null) {
                return ResponseEntity.ok(ConsumeResponse.builder()
                        .success(false)
                        .error("Topic does not exist")
                        .build());
            }
            
            ConsumerGroup group = topicManager.getConsumerGroup(topic, request.getGroupId());
            if (group == null || group.getConsumers().isEmpty()) {
                return ResponseEntity.ok(ConsumeResponse.builder()
                        .success(false)
                        .error("CONSUMER_NOT_EXIST")
                        .build());
            }
            
            return ResponseEntity.ok(ConsumeResponse.builder()
                    .success(false)
                    .error("No messages available")
                    .build());
        }

        Message msg = message.get();
        return ResponseEntity.ok(ConsumeResponse.builder()
                .success(true)
                .messageId(msg.getId())
                .body(msg.getBody())
                .format(msg.getFormat().name().toLowerCase())
                .build());
    }

    @PostMapping("/ack")
    public ResponseEntity<Map<String, Object>> acknowledge(@Valid @RequestBody AckRequest request) {
        boolean success = messageQueueService.acknowledge(
                request.getTopic(), request.getGroupId(), request.getConsumerId(), request.getMessageId());
        
        Map<String, Object> response = new HashMap<>();
        response.put("success", success);
        if (!success) {
            response.put("error", "Message not found or already acknowledged");
        }
        return ResponseEntity.ok(response);
    }

    @PostMapping("/dead-letter/resend")
    public ResponseEntity<Map<String, Object>> resendDeadLetter(@Valid @RequestBody ResendDeadLetterRequest request) {
        boolean success = messageQueueService.resendDeadLetter(request.getTopic(), request.getMessageId());
        
        Map<String, Object> response = new HashMap<>();
        response.put("success", success);
        if (!success) {
            response.put("error", "Message not found in dead letter queue");
        }
        return ResponseEntity.ok(response);
    }

    @GetMapping("/stats")
    public ResponseEntity<List<StatsResponse>> getStats() {
        List<StatsResponse> statsList = new ArrayList<>();
        
        for (Topic topic : topicManager.getAllTopics()) {
            long pendingCount = topic.getMessages().stream()
                    .filter(m -> m.getStatus() == MessageStatus.PENDING_DELIVERY ||
                            m.getStatus() == MessageStatus.DELIVERED)
                    .count();
            
            Map<String, StatsResponse.GroupStats> groupStats = new HashMap<>();
            for (ConsumerGroup group : topic.getConsumerGroups().values()) {
                groupStats.put(group.getId(), StatsResponse.GroupStats.builder()
                        .groupId(group.getId())
                        .consumedCount(group.getConsumedCount())
                        .lastConsumedIndex(group.getLastConsumedIndex())
                        .consumerCount(group.getConsumers().size())
                        .inFlightCount(group.getInFlightMessages().size())
                        .build());
            }
            
            statsList.add(StatsResponse.builder()
                    .topic(topic.getName())
                    .pendingCount(pendingCount)
                    .deadLetterCount(topic.getDeadLetterQueue().size())
                    .droppedCount(topic.getDroppedCount())
                    .groups(groupStats)
                    .build());
        }
        
        return ResponseEntity.ok(statsList);
    }

    @GetMapping("/stats/{topic}")
    public ResponseEntity<StatsResponse> getTopicStats(@PathVariable String topic) {
        Topic t = topicManager.getTopic(topic);
        if (t == null) {
            return ResponseEntity.notFound().build();
        }
        
        long pendingCount = t.getMessages().stream()
                .filter(m -> m.getStatus() == MessageStatus.PENDING_DELIVERY ||
                        m.getStatus() == MessageStatus.DELIVERED)
                .count();
        
        Map<String, StatsResponse.GroupStats> groupStats = new HashMap<>();
        for (ConsumerGroup group : t.getConsumerGroups().values()) {
            groupStats.put(group.getId(), StatsResponse.GroupStats.builder()
                    .groupId(group.getId())
                    .consumedCount(group.getConsumedCount())
                    .lastConsumedIndex(group.getLastConsumedIndex())
                    .consumerCount(group.getConsumers().size())
                    .inFlightCount(group.getInFlightMessages().size())
                    .build());
        }
        
        return ResponseEntity.ok(StatsResponse.builder()
                .topic(t.getName())
                .pendingCount(pendingCount)
                .deadLetterCount(t.getDeadLetterQueue().size())
                .droppedCount(t.getDroppedCount())
                .groups(groupStats)
                .build());
    }
}

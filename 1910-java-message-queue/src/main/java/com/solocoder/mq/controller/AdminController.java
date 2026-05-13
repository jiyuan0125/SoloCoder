package com.solocoder.mq.controller;

import com.fasterxml.jackson.annotation.JsonProperty;
import com.solocoder.mq.model.ConsumerGroup;
import com.solocoder.mq.model.Message;
import com.solocoder.mq.model.Topic;
import com.solocoder.mq.service.QueueService;
import lombok.Data;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import java.util.ArrayList;
import java.util.List;
import java.util.stream.Collectors;

@RestController
@RequestMapping("/topics")
public class AdminController {

    private final QueueService queueService;

    public AdminController(QueueService queueService) {
        this.queueService = queueService;
    }

    @GetMapping
    public ResponseEntity<List<TopicSummary>> listTopics() {
        List<TopicSummary> summaries = queueService.listTopics().stream()
                .map(topic -> {
                    TopicSummary summary = new TopicSummary();
                    summary.setName(topic.getName());
                    summary.setCapacity(topic.getCapacity());
                    summary.setTotal_messages(queueService.getTotalMessageCount(topic));
                    summary.setNext_offset(topic.getNextOffset());
                    return summary;
                })
                .collect(Collectors.toList());
        
        return ResponseEntity.ok(summaries);
    }

    @GetMapping("/{name}/groups/{gid}")
    public ResponseEntity<ConsumerProgress> getConsumerProgress(
            @PathVariable("name") String topicName,
            @PathVariable("gid") String groupId) {
        
        Topic topic = queueService.getTopic(topicName);
        if (topic == null) {
            return ResponseEntity.notFound().build();
        }
        
        ConsumerGroup group = queueService.getConsumerGroup(topicName, groupId);
        if (group == null) {
            ConsumerProgress progress = new ConsumerProgress();
            progress.setTopic(topicName);
            progress.setGroup_id(groupId);
            progress.setCurrent_offset(-1);
            progress.setLag(topic.getNextOffset());
            progress.setConsumer_active(false);
            return ResponseEntity.ok(progress);
        }
        
        long lag = queueService.getLag(topicName, groupId);
        
        ConsumerProgress progress = new ConsumerProgress();
        progress.setTopic(topicName);
        progress.setGroup_id(groupId);
        progress.setCurrent_offset(group.getCurrentOffsetValue());
        progress.setLag(lag);
        progress.setConsumer_active(group.isConsumerActive());
        progress.setConsumer_id(group.getConsumerId());
        
        return ResponseEntity.ok(progress);
    }

    @GetMapping("/{name}/dlq")
    public ResponseEntity<List<DlqMessageSummary>> getDlqMessages(
            @PathVariable("name") String topicName) {
        
        List<Message> dlqMessages = queueService.getDlqMessages(topicName);
        
        List<DlqMessageSummary> summaries = dlqMessages.stream()
                .map(msg -> {
                    DlqMessageSummary summary = new DlqMessageSummary();
                    summary.setId(msg.getId());
                    summary.setOffset(msg.getOffset());
                    summary.setContent(msg.getContent());
                    summary.setRetry_count(msg.getRetryCount());
                    return summary;
                })
                .collect(Collectors.toList());
        
        return ResponseEntity.ok(summaries);
    }

    @PostMapping("/{name}/dlq/{messageId}/requeue")
    public ResponseEntity<RequeueResponse> requeueDlqMessage(
            @PathVariable("name") String topicName,
            @PathVariable("messageId") String messageId) {
        
        boolean success = queueService.requeueDlqMessage(topicName, messageId);
        
        RequeueResponse resp = new RequeueResponse();
        resp.setSuccess(success);
        if (success) {
            resp.setMessage("Message requeued successfully");
        } else {
            resp.setMessage("Requeue failed - message not found or not in DLQ");
        }
        
        return ResponseEntity.ok(resp);
    }

    @Data
    public static class TopicSummary {
        @JsonProperty("name")
        private String name;
        
        @JsonProperty("capacity")
        private int capacity;
        
        @JsonProperty("total_messages")
        private int total_messages;
        
        @JsonProperty("next_offset")
        private long next_offset;
    }

    @Data
    public static class ConsumerProgress {
        @JsonProperty("topic")
        private String topic;
        
        @JsonProperty("group_id")
        private String group_id;
        
        @JsonProperty("current_offset")
        private long current_offset;
        
        @JsonProperty("lag")
        private long lag;
        
        @JsonProperty("consumer_active")
        private boolean consumer_active;
        
        @JsonProperty("consumer_id")
        private String consumer_id;
    }

    @Data
    public static class DlqMessageSummary {
        @JsonProperty("id")
        private String id;
        
        @JsonProperty("offset")
        private long offset;
        
        @JsonProperty("content")
        private String content;
        
        @JsonProperty("retry_count")
        private int retry_count;
    }

    @Data
    public static class RequeueResponse {
        @JsonProperty("success")
        private boolean success;
        
        @JsonProperty("message")
        private String message;
    }
}

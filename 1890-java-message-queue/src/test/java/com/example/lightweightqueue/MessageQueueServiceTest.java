package com.example.lightweightqueue;

import com.example.lightweightqueue.model.Message;
import com.example.lightweightqueue.model.MessageFormat;
import com.example.lightweightqueue.service.MessageQueueService;
import com.example.lightweightqueue.service.PersistenceService;
import com.example.lightweightqueue.service.TopicManager;
import com.fasterxml.jackson.databind.ObjectMapper;
import org.junit.jupiter.api.BeforeEach;
import org.junit.jupiter.api.Test;

import java.util.Optional;

import static org.junit.jupiter.api.Assertions.*;

class MessageQueueServiceTest {

    private MessageQueueService service;
    private TopicManager topicManager;
    private PersistenceService persistenceService;

    @BeforeEach
    void setUp() {
        topicManager = new TopicManager();
        persistenceService = new PersistenceService(new ObjectMapper());
        service = new MessageQueueService(topicManager, new ObjectMapper(), persistenceService);
    }

    @Test
    void testProduceAndConsume() {
        String messageId = service.produce("test-topic", "Hello, World!", "STRING");
        assertNotNull(messageId);

        Optional<Message> msg = service.consume("test-topic", "group1", "consumer1");
        assertTrue(msg.isPresent());
        assertEquals("Hello, World!", msg.get().getBody());
        assertEquals(MessageFormat.STRING, msg.get().getFormat());
        assertEquals(messageId, msg.get().getId());
    }

    @Test
    void testJsonFormat() {
        String jsonBody = "{\"name\":\"test\",\"value\":123}";
        String messageId = service.produce("test-json", jsonBody, "JSON");
        assertNotNull(messageId);

        Optional<Message> msg = service.consume("test-json", "group1", "consumer1");
        assertTrue(msg.isPresent());
        assertEquals(MessageFormat.JSON, msg.get().getFormat());
    }

    @Test
    void testInvalidJsonFormat() {
        String invalidJson = "{not valid json";
        assertThrows(IllegalArgumentException.class, () -> {
            service.produce("test-json", invalidJson, "JSON");
        });
    }

    @Test
    void testAcknowledge() {
        String messageId = service.produce("test-ack", "message", "STRING");
        
        Optional<Message> msg = service.consume("test-ack", "group1", "consumer1");
        assertTrue(msg.isPresent());
        
        boolean acknowledged = service.acknowledge("test-ack", "group1", "consumer1", messageId);
        assertTrue(acknowledged);
    }

    @Test
    void testConsumerGroupRoundRobin() {
        service.produce("test-rr", "msg1", "STRING");
        service.produce("test-rr", "msg2", "STRING");
        service.produce("test-rr", "msg3", "STRING");

        Optional<Message> msg1 = service.consume("test-rr", "group1", "consumer1");
        assertTrue(msg1.isPresent());

        Optional<Message> msg2 = service.consume("test-rr", "group1", "consumer2");
        assertTrue(msg2.isPresent());

        Optional<Message> msg3 = service.consume("test-rr", "group1", "consumer1");
        assertTrue(msg3.isPresent());
    }

    @Test
    void testTopicCapacity() {
        topicManager.createTopic("small-topic", 5);
        
        for (int i = 0; i < 10; i++) {
            service.produce("small-topic", "msg" + i, "STRING");
        }
        
        assertEquals(5, topicManager.getTopic("small-topic").getMessages().size());
        assertEquals(5, topicManager.getTopic("small-topic").getDroppedCount());
    }

    @Test
    void testNoExistingTopicReturnsEmpty() {
        Optional<Message> msg = service.consume("non-existent-topic", "group1", "consumer1");
        assertFalse(msg.isPresent());
    }
}

package com.example.messagequeue.exception;

import lombok.Getter;

@Getter
public class TopicNotFoundException extends RuntimeException {

    private final String topic;

    public TopicNotFoundException(String topic) {
        super(String.format("Topic '%s' not found", topic));
        this.topic = topic;
    }
}

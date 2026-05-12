package com.example.messagequeue.model;

import lombok.AllArgsConstructor;
import lombok.Data;
import lombok.NoArgsConstructor;

@Data
@NoArgsConstructor
@AllArgsConstructor
public class Message {

    private String id;

    private String topic;

    private String content;

    private String contentType;

    private long timestamp;

    private long offset;
}

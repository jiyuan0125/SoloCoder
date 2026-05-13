package com.messagequeue.dto;

import lombok.Data;

@Data
public class SubscribeRequest {
    private String groupId;
    private String filter;
}

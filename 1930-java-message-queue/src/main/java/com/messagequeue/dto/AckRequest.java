package com.messagequeue.dto;

import lombok.Data;

@Data
public class AckRequest {
    private String groupId;
    private String messageId;
}

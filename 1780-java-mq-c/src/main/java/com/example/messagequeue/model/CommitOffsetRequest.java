package com.example.messagequeue.model;

import jakarta.validation.constraints.NotNull;
import lombok.Data;

@Data
public class CommitOffsetRequest {

    @NotNull(message = "Offset must not be null")
    private Long offset;

    private String consumerId;
}

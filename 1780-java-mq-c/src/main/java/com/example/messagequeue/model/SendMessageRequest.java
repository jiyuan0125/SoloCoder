package com.example.messagequeue.model;

import jakarta.validation.constraints.NotBlank;
import lombok.Data;

@Data
public class SendMessageRequest {

    @NotBlank(message = "Content must not be blank")
    private String content;

    private String contentType = "text/plain";
}

package com.example.cachemiddleware.model;

import lombok.Data;

import javax.validation.constraints.NotBlank;

@Data
public class SubscribeRequest {
    @NotBlank(message = "callbackUrl cannot be blank")
    private String callbackUrl;
}

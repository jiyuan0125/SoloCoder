package com.example.configcenter.dto;

import lombok.Data;

@Data
public class SubscriberRequest {
    private String key;
    private String callbackUrl;
}

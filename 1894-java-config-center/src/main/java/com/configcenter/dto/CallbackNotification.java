package com.configcenter.dto;

import lombok.Data;
import lombok.AllArgsConstructor;
import lombok.NoArgsConstructor;

@Data
@NoArgsConstructor
@AllArgsConstructor
public class CallbackNotification {
    private String application;
    private String environment;
    private String key;
    private String oldValue;
    private String newValue;
}

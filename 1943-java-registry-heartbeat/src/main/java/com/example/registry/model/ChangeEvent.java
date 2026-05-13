package com.example.registry.model;

import lombok.AllArgsConstructor;
import lombok.Builder;
import lombok.Data;
import lombok.NoArgsConstructor;

@Data
@Builder
@NoArgsConstructor
@AllArgsConstructor
public class ChangeEvent {
    private ChangeType type;
    private ServiceInstance instance;

    public enum ChangeType {
        REGISTER,
        DEREGISTER,
        STATUS_CHANGE
    }
}

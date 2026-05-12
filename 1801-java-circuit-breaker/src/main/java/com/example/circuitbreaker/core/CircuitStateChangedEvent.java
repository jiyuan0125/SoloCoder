package com.example.circuitbreaker.core;

import lombok.Getter;
import org.springframework.context.ApplicationEvent;

@Getter
public class CircuitStateChangedEvent extends ApplicationEvent {
    private final String serviceName;
    private final CircuitState fromState;
    private final CircuitState toState;
    private final String reason;

    public CircuitStateChangedEvent(Object source, String serviceName,
                                    CircuitState fromState, CircuitState toState, String reason) {
        super(source);
        this.serviceName = serviceName;
        this.fromState = fromState;
        this.toState = toState;
        this.reason = reason;
    }
}

package com.example.serviceregistry.event;

import com.example.serviceregistry.model.ServiceInstance;
import lombok.Getter;
import org.springframework.context.ApplicationEvent;

import java.util.List;

@Getter
public class ServiceChangedEvent extends ApplicationEvent {

    private final String serviceName;
    private final List<ServiceInstance> healthyInstances;

    public ServiceChangedEvent(Object source, String serviceName, List<ServiceInstance> healthyInstances) {
        super(source);
        this.serviceName = serviceName;
        this.healthyInstances = healthyInstances;
    }
}

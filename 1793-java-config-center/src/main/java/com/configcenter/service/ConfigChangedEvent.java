package com.configcenter.service;

import lombok.Getter;
import org.springframework.context.ApplicationEvent;

@Getter
public class ConfigChangedEvent extends ApplicationEvent {
    
    private final String namespace;
    private final String group;
    private final String key;
    
    public ConfigChangedEvent(Object source, String namespace, String group, String key) {
        super(source);
        this.namespace = namespace;
        this.group = group;
        this.key = key;
    }
}

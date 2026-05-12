package com.configcenter.service;

import lombok.extern.slf4j.Slf4j;
import org.springframework.stereotype.Service;

import java.util.Set;
import java.util.concurrent.ConcurrentHashMap;
import java.util.concurrent.CopyOnWriteArraySet;

@Service
@Slf4j
public class ConfigChangeNotifier {
    
    private final ConcurrentHashMap<String, Set<LongPollSubscriber>> subscribers = new ConcurrentHashMap<>();
    
    public void notifyChange(String namespace, String group, String key) {
        String configKey = buildKey(namespace, group, key);
        log.debug("Notifying change for: {}", configKey);
        
        Set<LongPollSubscriber> subs = subscribers.get(configKey);
        if (subs != null) {
            for (LongPollSubscriber sub : subs) {
                sub.onChange(namespace, group, key);
            }
        }
    }
    
    public LongPollSubscriber subscribe(String namespace, String group, String key) {
        String configKey = buildKey(namespace, group, key);
        LongPollSubscriber subscriber = new LongPollSubscriber(namespace, group, key);
        
        subscribers.compute(configKey, (k, existing) -> {
            Set<LongPollSubscriber> set = existing != null ? existing : new CopyOnWriteArraySet<>();
            set.add(subscriber);
            return set;
        });
        
        return subscriber;
    }
    
    public void unsubscribe(String namespace, String group, String key, LongPollSubscriber subscriber) {
        String configKey = buildKey(namespace, group, key);
        Set<LongPollSubscriber> subs = subscribers.get(configKey);
        if (subs != null) {
            subs.remove(subscriber);
            if (subs.isEmpty()) {
                subscribers.remove(configKey);
            }
        }
    }
    
    private String buildKey(String namespace, String group, String key) {
        return namespace + ":" + group + ":" + key;
    }
}

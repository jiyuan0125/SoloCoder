package com.configcenter.service;

import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.context.event.EventListener;
import org.springframework.scheduling.annotation.Async;
import org.springframework.stereotype.Component;
import org.springframework.transaction.event.TransactionPhase;
import org.springframework.transaction.event.TransactionalEventListener;

@Component
@RequiredArgsConstructor
@Slf4j
public class ConfigChangeEventListener {
    
    private final ConfigChangeNotifier configChangeNotifier;
    
    @TransactionalEventListener(phase = TransactionPhase.AFTER_COMMIT)
    @Async
    public void handleConfigChangedEvent(ConfigChangedEvent event) {
        log.debug("Handling config changed event after commit: {}:{}:{}", 
                event.getNamespace(), event.getGroup(), event.getKey());
        configChangeNotifier.notifyChange(
                event.getNamespace(), 
                event.getGroup(), 
                event.getKey()
        );
    }
    
    @EventListener
    public void handleNonTransactionalConfigChanged(ConfigChangedEvent event) {
        log.debug("Handling config changed event (non-transactional): {}:{}:{}",
                event.getNamespace(), event.getGroup(), event.getKey());
    }
}

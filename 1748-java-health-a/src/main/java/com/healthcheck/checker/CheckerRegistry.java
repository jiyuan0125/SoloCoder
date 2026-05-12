package com.healthcheck.checker;

import com.healthcheck.model.CheckItemConfig;
import org.springframework.stereotype.Component;

import java.util.List;
import java.util.Optional;

@Component
public class CheckerRegistry {
    
    private final List<Checker> checkers;
    
    public CheckerRegistry(List<Checker> checkers) {
        this.checkers = checkers;
    }
    
    public Checker getChecker(CheckItemConfig config) {
        Optional<Checker> checker = checkers.stream()
                .filter(c -> c.supports(config))
                .findFirst();
        
        return checker.orElseThrow(() -> new IllegalArgumentException(
                "No checker found for config: " + config.getName()));
    }
}

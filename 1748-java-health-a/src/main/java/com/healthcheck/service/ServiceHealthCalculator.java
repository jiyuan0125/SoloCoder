package com.healthcheck.service;

import com.healthcheck.model.CheckItemState;
import com.healthcheck.model.CheckStatus;
import com.healthcheck.model.ServiceHealthStatus;
import org.springframework.stereotype.Service;

import java.util.Collection;

@Service
public class ServiceHealthCalculator {
    
    public ServiceHealthStatus calculate(Collection<CheckItemState> checkItemStates) {
        if (checkItemStates == null || checkItemStates.isEmpty()) {
            return ServiceHealthStatus.UNKNOWN;
        }
        
        boolean hasError = false;
        boolean hasWarning = false;
        
        for (CheckItemState state : checkItemStates) {
            CheckStatus status = state.getCurrentStatus();
            
            if (status == CheckStatus.ERROR) {
                hasError = true;
                break;
            } else if (status == CheckStatus.WARNING) {
                hasWarning = true;
            }
        }
        
        if (hasError) {
            return ServiceHealthStatus.UNHEALTHY;
        } else if (hasWarning) {
            return ServiceHealthStatus.SUB_HEALTHY;
        } else {
            return ServiceHealthStatus.HEALTHY;
        }
    }
}

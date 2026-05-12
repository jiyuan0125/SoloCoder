package com.healthcheck.service;

import com.healthcheck.model.CheckStatus;
import com.healthcheck.model.CheckItemState;
import com.healthcheck.model.CheckResult;
import org.springframework.stereotype.Service;

import java.time.Instant;

@Service
public class StateMachineService {

    public CheckItemState transitionState(CheckItemState currentState, CheckResult newResult) {
        CheckStatus newStatus = determineNewStatus(currentState, newResult);
        
        boolean statusChanged = newStatus != currentState.getCurrentStatus();
        
        int newConsecutiveFailures = newStatus == CheckStatus.NORMAL ? 0 :
                (newResult.isSuccess() ? currentState.getConsecutiveFailures() :
                 currentState.getConsecutiveFailures() + 1);
        
        return CheckItemState.builder()
                .serviceId(currentState.getServiceId())
                .checkItemName(currentState.getCheckItemName())
                .currentStatus(newStatus)
                .previousStatus(statusChanged ? currentState.getCurrentStatus() : currentState.getPreviousStatus())
                .lastStateChangeTime(statusChanged ? Instant.now() : currentState.getLastStateChangeTime())
                .consecutiveFailures(newConsecutiveFailures)
                .lastResult(newResult)
                .build();
    }

    private CheckStatus determineNewStatus(CheckItemState currentState, CheckResult newResult) {
        CheckStatus currentStatus = currentState.getCurrentStatus();
        boolean success = newResult.isSuccess();
        
        if (success) {
            if (currentStatus == CheckStatus.ERROR) {
                return CheckStatus.NORMAL;
            }
            return CheckStatus.NORMAL;
        }
        
        return CheckStatus.nextWorse(currentStatus);
    }

    public CheckItemState createInitialState(String serviceId, String checkItemName) {
        return CheckItemState.builder()
                .serviceId(serviceId)
                .checkItemName(checkItemName)
                .currentStatus(CheckStatus.UNKNOWN)
                .previousStatus(CheckStatus.UNKNOWN)
                .lastStateChangeTime(Instant.now())
                .consecutiveFailures(0)
                .build();
    }
}

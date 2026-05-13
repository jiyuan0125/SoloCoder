package com.healthchecker.engine;

import com.healthchecker.entity.ServiceStatus;
import com.healthchecker.entity.ServiceConfig;
import org.springframework.stereotype.Component;

@Component
public class StateTransitionEngine {

    private static final int UNKNOWN_TO_HEALTHY_SUCCESS = 2;
    private static final int HEALTHY_TO_WARNING_FAILURE = 3;
    private static final int WARNING_TO_FAULT_FAILURE = 3;
    private static final int FAULT_TO_HEALTHY_SUCCESS = 5;

    public TransitionResult processCheckResult(ServiceConfig config, boolean checkSuccess) {
        ServiceStatus currentStatus = config.getCurrentStatus();
        if (currentStatus == ServiceStatus.MAINTENANCE) {
            return new TransitionResult(currentStatus, currentStatus, false, 0);
        }

        if (checkSuccess) {
            config.incrementConsecutiveSuccess();
            return handleSuccess(config);
        } else {
            config.incrementConsecutiveFailure();
            return handleFailure(config);
        }
    }

    private TransitionResult handleSuccess(ServiceConfig config) {
        ServiceStatus oldStatus = config.getCurrentStatus();
        int consecutiveSuccess = config.getConsecutiveSuccessCount();

        ServiceStatus newStatus = switch (oldStatus) {
            case UNKNOWN -> consecutiveSuccess >= UNKNOWN_TO_HEALTHY_SUCCESS ? ServiceStatus.HEALTHY : oldStatus;
            case HEALTHY -> oldStatus;
            case WARNING -> oldStatus;
            case FAULT -> consecutiveSuccess >= FAULT_TO_HEALTHY_SUCCESS ? ServiceStatus.HEALTHY : oldStatus;
            default -> oldStatus;
        };

        boolean changed = oldStatus != newStatus;
        if (changed) {
            config.setCurrentStatus(newStatus);
        }

        return new TransitionResult(oldStatus, newStatus, changed, consecutiveSuccess);
    }

    private TransitionResult handleFailure(ServiceConfig config) {
        ServiceStatus oldStatus = config.getCurrentStatus();
        int consecutiveFailure = config.getConsecutiveFailureCount();

        ServiceStatus newStatus = switch (oldStatus) {
            case UNKNOWN -> oldStatus;
            case HEALTHY -> consecutiveFailure >= HEALTHY_TO_WARNING_FAILURE ? ServiceStatus.WARNING : oldStatus;
            case WARNING -> consecutiveFailure >= WARNING_TO_FAULT_FAILURE ? ServiceStatus.FAULT : oldStatus;
            case FAULT -> oldStatus;
            default -> oldStatus;
        };

        boolean changed = oldStatus != newStatus;
        if (changed) {
            config.setCurrentStatus(newStatus);
        }

        return new TransitionResult(oldStatus, newStatus, changed, consecutiveFailure);
    }

    public void resetToUnknown(ServiceConfig config) {
        config.setCurrentStatus(ServiceStatus.UNKNOWN);
        config.resetCounters();
    }

    public record TransitionResult(
            ServiceStatus oldStatus,
            ServiceStatus newStatus,
            boolean statusChanged,
            int consecutiveCount
    ) {}
}
package com.health.checker;

import com.health.model.CheckResult;
import com.health.model.ServiceRegistration;

public interface HealthChecker {
    CheckResult check(ServiceRegistration service);
    boolean supports(ServiceRegistration service);
}
package com.health.service;

import com.health.model.ServiceRegistration;

public interface NotificationService {
    void notifyUnhealthy(ServiceRegistration service);
    void notifyRecovered(ServiceRegistration service);
}
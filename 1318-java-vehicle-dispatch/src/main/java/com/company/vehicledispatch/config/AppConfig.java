package com.company.vehicledispatch.config;

import lombok.Data;
import org.springframework.boot.context.properties.ConfigurationProperties;
import org.springframework.stereotype.Component;

@Data
@Component
@ConfigurationProperties(prefix = "app")
public class AppConfig {
    private Vehicle vehicle = new Vehicle();
    private Dispatch dispatch = new Dispatch();
    private LongDistance longDistance = new LongDistance();
    private Double fuelDeviationThreshold;
    private Integer insuranceReminderDays;

    @Data
    public static class Vehicle {
        private Integer maintenanceIntervalKm;
        private Integer fuelWarningThreshold;
        private Integer fuelUnavailableThreshold;
    }

    @Data
    public static class Dispatch {
        private Integer bufferMinutes;
    }

    @Data
    public static class LongDistance {
        private Integer thresholdKm;
    }
}

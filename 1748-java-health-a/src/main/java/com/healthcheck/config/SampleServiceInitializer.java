package com.healthcheck.config;

import com.healthcheck.model.CheckItemConfig;
import com.healthcheck.model.CheckType;
import com.healthcheck.model.ServiceConfig;
import com.healthcheck.scheduler.HealthCheckScheduler;
import com.healthcheck.service.ServiceConfigStore;
import jakarta.annotation.PostConstruct;
import org.springframework.stereotype.Component;

import java.util.Arrays;

@Component
public class SampleServiceInitializer {

    private final ServiceConfigStore configStore;
    private final HealthCheckScheduler scheduler;

    public SampleServiceInitializer(ServiceConfigStore configStore,
                                    HealthCheckScheduler scheduler) {
        this.configStore = configStore;
        this.scheduler = scheduler;
    }

    @PostConstruct
    public void init() {
        ServiceConfig apiService = new ServiceConfig();
        apiService.setServiceId("api-gateway");
        apiService.setServiceName("API Gateway Service");
        apiService.setDescription("Main API Gateway");
        apiService.setCheckInterval(10);
        apiService.setWebhookUrl("http://localhost:9000/webhook");
        apiService.setEnabled(true);

        CheckItemConfig httpCheck = new CheckItemConfig();
        httpCheck.setName("health-http");
        httpCheck.setType(CheckType.HTTP);
        httpCheck.setTarget("http://localhost:8081/actuator/health");
        httpCheck.setTimeout(5000);
        httpCheck.setMethod("GET");
        httpCheck.setExpectedStatusCode(200);

        CheckItemConfig tcpCheck = new CheckItemConfig();
        tcpCheck.setName("health-tcp");
        tcpCheck.setType(CheckType.TCP);
        tcpCheck.setHost("localhost");
        tcpCheck.setPort(8081);
        tcpCheck.setTimeout(5000);

        apiService.setCheckItems(Arrays.asList(httpCheck, tcpCheck));

        configStore.registerService(apiService);
        scheduler.scheduleService(apiService);
    }
}

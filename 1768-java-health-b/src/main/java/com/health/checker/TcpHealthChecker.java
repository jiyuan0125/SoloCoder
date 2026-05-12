package com.health.checker;

import com.health.model.CheckResult;
import com.health.model.CheckType;
import com.health.model.ServiceRegistration;
import org.springframework.stereotype.Component;

import java.io.IOException;
import java.net.InetSocketAddress;
import java.net.Socket;
import java.time.LocalDateTime;

@Component
public class TcpHealthChecker implements HealthChecker {

    @Override
    public CheckResult check(ServiceRegistration service) {
        LocalDateTime start = LocalDateTime.now();
        long startTime = System.currentTimeMillis();

        String host = service.getEndpoint();
        Integer port = service.getPort();
        
        if (port == null) {
            return CheckResult.builder()
                    .healthy(false)
                    .responseTimeMs(0)
                    .message("TCP check requires port configuration")
                    .checkedAt(start)
                    .build();
        }

        int timeoutMs = service.getTimeoutMilliseconds() != null ? 
                service.getTimeoutMilliseconds() : 5000;

        try (Socket socket = new Socket()) {
            socket.connect(new InetSocketAddress(host, port), timeoutMs);
            int responseTime = (int) (System.currentTimeMillis() - startTime);
            
            return CheckResult.builder()
                    .healthy(true)
                    .responseTimeMs(responseTime)
                    .message("OK - Connected to " + host + ":" + port)
                    .checkedAt(start)
                    .build();

        } catch (IOException e) {
            int responseTime = (int) (System.currentTimeMillis() - startTime);
            return CheckResult.builder()
                    .healthy(false)
                    .responseTimeMs(responseTime)
                    .message("Connection failed: " + e.getMessage())
                    .checkedAt(start)
                    .build();
        }
    }

    @Override
    public boolean supports(ServiceRegistration service) {
        return CheckType.TCP.equals(service.getCheckType());
    }
}
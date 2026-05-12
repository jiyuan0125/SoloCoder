package com.healthcheck.checker;

import com.healthcheck.model.CheckItemConfig;
import com.healthcheck.model.CheckResult;
import com.healthcheck.model.CheckStatus;
import com.healthcheck.model.CheckType;
import org.springframework.stereotype.Component;

import java.io.IOException;
import java.net.InetSocketAddress;
import java.net.Socket;
import java.time.Instant;

@Component
public class TcpChecker implements Checker {

    @Override
    public CheckResult execute(CheckItemConfig config, String serviceId) {
        long startTime = System.currentTimeMillis();
        Instant now = Instant.now();
        
        String host = config.getHost();
        Integer port = config.getPort();
        int timeout = config.getTimeout() != null ? config.getTimeout() : 5000;
        
        if (host == null || port == null) {
            return CheckResult.builder()
                    .serviceId(serviceId)
                    .checkItemName(config.getName())
                    .status(CheckStatus.ERROR)
                    .message("TCP check failed: host or port not configured")
                    .responseTime(System.currentTimeMillis() - startTime)
                    .timestamp(now)
                    .success(false)
                    .build();
        }

        try (Socket socket = new Socket()) {
            socket.connect(new InetSocketAddress(host, port), timeout);
            
            long responseTime = System.currentTimeMillis() - startTime;
            return CheckResult.builder()
                    .serviceId(serviceId)
                    .checkItemName(config.getName())
                    .status(CheckStatus.NORMAL)
                    .message("TCP connection successful to " + host + ":" + port)
                    .responseTime(responseTime)
                    .timestamp(now)
                    .success(true)
                    .build();

        } catch (IOException e) {
            long responseTime = System.currentTimeMillis() - startTime;
            return CheckResult.builder()
                    .serviceId(serviceId)
                    .checkItemName(config.getName())
                    .status(CheckStatus.ERROR)
                    .message("TCP check failed: " + e.getMessage())
                    .responseTime(responseTime)
                    .timestamp(now)
                    .success(false)
                    .build();
        }
    }

    @Override
    public boolean supports(CheckItemConfig config) {
        return CheckType.TCP.equals(config.getType());
    }
}

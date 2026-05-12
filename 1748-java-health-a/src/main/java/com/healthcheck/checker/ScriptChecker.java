package com.healthcheck.checker;

import com.healthcheck.model.CheckItemConfig;
import com.healthcheck.model.CheckResult;
import com.healthcheck.model.CheckStatus;
import com.healthcheck.model.CheckType;
import org.springframework.stereotype.Component;

import java.io.BufferedReader;
import java.io.InputStreamReader;
import java.nio.charset.StandardCharsets;
import java.time.Instant;
import java.util.concurrent.TimeUnit;

@Component
public class ScriptChecker implements Checker {

    @Override
    public CheckResult execute(CheckItemConfig config, String serviceId) {
        long startTime = System.currentTimeMillis();
        Instant now = Instant.now();
        
        String scriptPath = config.getScriptPath();
        if (scriptPath == null || scriptPath.isEmpty()) {
            return CheckResult.builder()
                    .serviceId(serviceId)
                    .checkItemName(config.getName())
                    .status(CheckStatus.ERROR)
                    .message("Script check failed: script path not configured")
                    .responseTime(System.currentTimeMillis() - startTime)
                    .timestamp(now)
                    .success(false)
                    .build();
        }

        int timeout = config.getTimeout() != null ? config.getTimeout() : 30000;
        
        try {
            ProcessBuilder processBuilder = new ProcessBuilder(scriptPath);
            processBuilder.redirectErrorStream(true);
            
            Process process = processBuilder.start();
            
            StringBuilder output = new StringBuilder();
            try (BufferedReader reader = new BufferedReader(
                    new InputStreamReader(process.getInputStream(), StandardCharsets.UTF_8))) {
                String line;
                while ((line = reader.readLine()) != null) {
                    output.append(line).append("\n");
                }
            }
            
            boolean completed = process.waitFor(timeout, TimeUnit.MILLISECONDS);
            
            if (!completed) {
                process.destroyForcibly();
                long responseTime = System.currentTimeMillis() - startTime;
                return CheckResult.builder()
                        .serviceId(serviceId)
                        .checkItemName(config.getName())
                        .status(CheckStatus.ERROR)
                        .message("Script check timeout after " + timeout + "ms")
                        .responseTime(responseTime)
                        .timestamp(now)
                        .success(false)
                        .build();
            }
            
            int exitCode = process.exitValue();
            long responseTime = System.currentTimeMillis() - startTime;
            boolean success = exitCode == 0;
            
            return CheckResult.builder()
                    .serviceId(serviceId)
                    .checkItemName(config.getName())
                    .status(success ? CheckStatus.NORMAL : CheckStatus.ERROR)
                    .message("Script exit code: " + exitCode + "\n" + output.toString().trim())
                    .responseTime(responseTime)
                    .timestamp(now)
                    .success(success)
                    .build();

        } catch (Exception e) {
            long responseTime = System.currentTimeMillis() - startTime;
            return CheckResult.builder()
                    .serviceId(serviceId)
                    .checkItemName(config.getName())
                    .status(CheckStatus.ERROR)
                    .message("Script check failed: " + e.getMessage())
                    .responseTime(responseTime)
                    .timestamp(now)
                    .success(false)
                    .build();
        }
    }

    @Override
    public boolean supports(CheckItemConfig config) {
        return CheckType.SCRIPT.equals(config.getType());
    }
}

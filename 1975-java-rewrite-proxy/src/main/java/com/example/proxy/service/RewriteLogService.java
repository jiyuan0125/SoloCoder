package com.example.proxy.service;

import com.example.proxy.model.RewriteLog;
import org.springframework.stereotype.Service;

import java.util.ArrayList;
import java.util.Collections;
import java.util.List;

@Service
public class RewriteLogService {

    private static final int MAX_LOGS = 5000;
    private final List<RewriteLog> logs = Collections.synchronizedList(new ArrayList<>());

    public void log(Long ruleId, String originalPath, String rewrittenPath) {
        RewriteLog log = new RewriteLog(ruleId, originalPath, rewrittenPath);
        synchronized (logs) {
            logs.add(log);
            while (logs.size() > MAX_LOGS) {
                logs.remove(0);
            }
        }
    }

    public List<RewriteLog> getAllLogs() {
        return new ArrayList<>(logs);
    }
}

package com.healthcheck.service;

import com.healthcheck.model.CheckResult;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.stereotype.Service;

import java.util.ArrayList;
import java.util.Collections;
import java.util.LinkedList;
import java.util.List;
import java.util.Map;
import java.util.concurrent.ConcurrentHashMap;

@Service
public class HistoryService {
    
    @Value("${healthcheck.history-size:100}")
    private int maxHistorySize;
    
    private final Map<String, List<CheckResult>> histories = new ConcurrentHashMap<>();
    
    public void addResult(CheckResult result) {
        String key = result.getServiceId() + ":" + result.getCheckItemName();
        histories.compute(key, (k, list) -> {
            List<CheckResult> newList = list != null ? list : new LinkedList<>();
            newList.add(result);
            while (newList.size() > maxHistorySize) {
                newList.remove(0);
            }
            return newList;
        });
    }
    
    public List<CheckResult> getHistory(String serviceId, String checkItemName, int n) {
        String key = serviceId + ":" + checkItemName;
        List<CheckResult> history = histories.get(key);
        
        if (history == null || history.isEmpty()) {
            return Collections.emptyList();
        }
        
        List<CheckResult> result = new ArrayList<>(history);
        int start = Math.max(0, result.size() - n);
        return result.subList(start, result.size());
    }
    
    public List<CheckResult> getServiceHistory(String serviceId, int n) {
        List<CheckResult> allResults = new ArrayList<>();
        
        for (Map.Entry<String, List<CheckResult>> entry : histories.entrySet()) {
            if (entry.getKey().startsWith(serviceId + ":")) {
                allResults.addAll(entry.getValue());
            }
        }
        
        allResults.sort((a, b) -> b.getTimestamp().compareTo(a.getTimestamp()));
        
        return allResults.size() > n ? allResults.subList(0, n) : allResults;
    }
}

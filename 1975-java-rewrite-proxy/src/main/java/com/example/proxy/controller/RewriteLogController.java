package com.example.proxy.controller;

import com.example.proxy.model.RewriteLog;
import com.example.proxy.service.RewriteLogService;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;

import java.util.List;

@RestController
@RequestMapping("/rewrite-logs")
public class RewriteLogController {

    private final RewriteLogService logService;

    public RewriteLogController(RewriteLogService logService) {
        this.logService = logService;
    }

    @GetMapping
    public List<RewriteLog> getAllLogs() {
        return logService.getAllLogs();
    }
}

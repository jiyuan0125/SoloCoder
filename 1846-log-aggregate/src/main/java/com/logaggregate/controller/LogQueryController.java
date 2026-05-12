package com.logaggregate.controller;

import com.logaggregate.model.LogPageResult;
import com.logaggregate.service.ingest.TimestampParser;
import com.logaggregate.service.query.LogQueryService;
import org.springframework.web.bind.annotation.*;

import java.util.List;

@RestController
@RequestMapping("/logs")
public class LogQueryController {

    private final LogQueryService logQueryService;

    public LogQueryController(LogQueryService logQueryService) {
        this.logQueryService = logQueryService;
    }

    @GetMapping
    public LogPageResult queryLogs(
            @RequestParam(required = false) String service,
            @RequestParam(required = false) List<String> level,
            @RequestParam(required = false, name = "start_time") String startTimeStr,
            @RequestParam(required = false, name = "end_time") String endTimeStr,
            @RequestParam(required = false) String keyword,
            @RequestParam(required = false) Integer page,
            @RequestParam(required = false, name = "page_size") Integer pageSize
    ) {
        Long startTimeUtcMs = parseTime(startTimeStr);
        Long endTimeUtcMs = parseTime(endTimeStr);

        return logQueryService.queryLogs(service, level, startTimeUtcMs, endTimeUtcMs, keyword, page, pageSize);
    }

    private Long parseTime(String timeStr) {
        if (timeStr == null || timeStr.isBlank()) {
            return null;
        }
        try {
            return TimestampParser.parseToUtcMillis(timeStr);
        } catch (Exception e) {
            return null;
        }
    }
}

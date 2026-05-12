package com.logaggregate.service.query;

import com.logaggregate.model.LogPageResult;
import com.logaggregate.model.StoredLogEntry;
import com.logaggregate.service.storage.LogStorageService;
import org.springframework.stereotype.Service;

import java.util.List;

@Service
public class LogQueryService {

    private static final int DEFAULT_PAGE = 1;
    private static final int DEFAULT_PAGE_SIZE = 20;
    private static final int MAX_PAGE_SIZE = 100;

    private final LogStorageService logStorageService;

    public LogQueryService(LogStorageService logStorageService) {
        this.logStorageService = logStorageService;
    }

    public LogPageResult queryLogs(String service, List<String> levels, Long startTimeUtcMs, Long endTimeUtcMs, String keyword, Integer page, Integer pageSize) {
        int actualPage = (page == null || page < 1) ? DEFAULT_PAGE : page;
        int actualPageSize = (pageSize == null || pageSize < 1) ? DEFAULT_PAGE_SIZE : Math.min(pageSize, MAX_PAGE_SIZE);

        List<StoredLogEntry> allResults = logStorageService.queryLogs(service, levels, startTimeUtcMs, endTimeUtcMs, keyword);

        long total = allResults.size();
        int startIndex = (actualPage - 1) * actualPageSize;
        int endIndex = Math.min(startIndex + actualPageSize, allResults.size());

        List<StoredLogEntry> pageData;
        if (startIndex >= allResults.size()) {
            pageData = List.of();
        } else {
            pageData = allResults.subList(startIndex, endIndex);
        }

        return new LogPageResult(pageData, total, actualPage, actualPageSize);
    }
}

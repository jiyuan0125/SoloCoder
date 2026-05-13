package com.example.lightweightqueue.dto;

import lombok.Data;
import lombok.Builder;
import lombok.AllArgsConstructor;
import lombok.NoArgsConstructor;

import java.util.Map;

@Data
@Builder
@NoArgsConstructor
@AllArgsConstructor
public class StatsResponse {
    private String topic;
    private long pendingCount;
    private long deadLetterCount;
    private long droppedCount;
    private Map<String, GroupStats> groups;

    @Data
    @Builder
    @NoArgsConstructor
    @AllArgsConstructor
    public static class GroupStats {
        private String groupId;
        private long consumedCount;
        private long lastConsumedIndex;
        private int consumerCount;
        private int inFlightCount;
    }
}

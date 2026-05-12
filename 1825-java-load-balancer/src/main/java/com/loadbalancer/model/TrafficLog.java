package com.loadbalancer.model;

import lombok.Data;
import lombok.Builder;
import lombok.AllArgsConstructor;
import lombok.NoArgsConstructor;

import java.time.LocalDateTime;

@Data
@Builder
@AllArgsConstructor
@NoArgsConstructor
public class TrafficLog {
    private String id;
    private String nodeId;
    private SelectionReason reason;
    private String requestPath;
    private LocalDateTime timestamp;

    public enum SelectionReason {
        WEIGHTED_ROUND_ROBIN,
        PATH_MATCHED,
        TAG_FILTERED
    }
}

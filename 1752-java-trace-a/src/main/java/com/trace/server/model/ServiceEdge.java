package com.trace.server.model;

import lombok.AllArgsConstructor;
import lombok.Data;

@Data
@AllArgsConstructor
public class ServiceEdge {
    private String fromServiceId;
    private String toServiceId;
    private Long callCount;
}

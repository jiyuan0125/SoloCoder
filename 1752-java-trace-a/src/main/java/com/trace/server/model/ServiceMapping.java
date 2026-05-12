package com.trace.server.model;

import lombok.Data;
import java.time.Instant;

@Data
public class ServiceMapping {
    private String serviceId;
    private String currentName;
    private String previousName;
    private Instant updatedAt;
}

package com.trace.server.model;

import lombok.AllArgsConstructor;
import lombok.Data;

@Data
@AllArgsConstructor
public class ServiceNode {
    private String serviceId;
    private String serviceName;
}

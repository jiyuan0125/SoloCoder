package com.depgraph.model;

import lombok.AllArgsConstructor;
import lombok.Data;
import lombok.NoArgsConstructor;

import java.util.List;

@Data
@NoArgsConstructor
@AllArgsConstructor
public class ImpactResult {
    private String sourceService;
    private List<AffectedService> affectedServices;

    @Data
    @NoArgsConstructor
    @AllArgsConstructor
    public static class AffectedService {
        private String serviceName;
        private int dependencyDepth;
        private List<String> impactPath;
        private boolean direct;
    }
}

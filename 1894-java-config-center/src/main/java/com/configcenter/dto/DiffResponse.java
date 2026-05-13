package com.configcenter.dto;

import lombok.Data;
import lombok.AllArgsConstructor;
import lombok.NoArgsConstructor;
import java.util.List;

@Data
@NoArgsConstructor
@AllArgsConstructor
public class DiffResponse {
    private Long fromVersion;
    private Long toVersion;
    private List<ConfigChange> changes;

    @Data
    @NoArgsConstructor
    @AllArgsConstructor
    public static class ConfigChange {
        private String key;
        private String oldValue;
        private String newValue;
        private ChangeType changeType;
    }

    public enum ChangeType {
        ADDED, MODIFIED, DELETED
    }
}

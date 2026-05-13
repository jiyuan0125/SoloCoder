package com.edi.converter.mapping;

import lombok.AllArgsConstructor;
import lombok.Data;
import lombok.NoArgsConstructor;

import java.util.ArrayList;
import java.util.List;

@Data
@NoArgsConstructor
@AllArgsConstructor
public class MappingRule {
    private String partnerId;
    private String version;
    private String documentType;
    private List<SegmentMapping> segmentMappings = new ArrayList<>();
    private List<RequiredField> requiredFields = new ArrayList<>();
    private DelimiterConfig delimiters = new DelimiterConfig();

    @Data
    @NoArgsConstructor
    @AllArgsConstructor
    public static class SegmentMapping {
        private String sourceSegmentTag;
        private String targetSegmentTag;
        private List<DataElementMapping> dataElementMappings = new ArrayList<>();
        private int occurrence = 1;
    }

    @Data
    @NoArgsConstructor
    @AllArgsConstructor
    public static class DataElementMapping {
        private int sourceElementPosition;
        private int targetElementPosition;
        private int sourceComponentPosition;
        private int targetComponentPosition;
        private String fieldName;
        private String defaultValue;
    }

    @Data
    @NoArgsConstructor
    @AllArgsConstructor
    public static class RequiredField {
        private String segmentTag;
        private int elementPosition;
        private int componentPosition;
        private String fieldName;
        private String expectedValue;
        private boolean required = true;
    }

    @Data
    @NoArgsConstructor
    @AllArgsConstructor
    public static class DelimiterConfig {
        private char segmentTerminator = '\'';
        private char dataElementSeparator = '+';
        private char componentSeparator = ':';
        private char releaseCharacter = '?';
    }
}

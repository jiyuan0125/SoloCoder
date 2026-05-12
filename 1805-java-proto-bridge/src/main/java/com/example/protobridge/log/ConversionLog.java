package com.example.protobridge.log;

import lombok.AllArgsConstructor;
import lombok.Builder;
import lombok.Data;
import lombok.NoArgsConstructor;

import java.time.LocalDateTime;

@Data
@Builder
@NoArgsConstructor
@AllArgsConstructor
public class ConversionLog {

    private String id;

    private String direction;

    private String path;

    private String method;

    private Long durationMs;

    private Boolean success;

    private String errorReason;

    private LocalDateTime timestamp;

    private String ruleId;
}

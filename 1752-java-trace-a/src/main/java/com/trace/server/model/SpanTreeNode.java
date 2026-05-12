package com.trace.server.model;

import lombok.Data;
import java.util.List;

@Data
public class SpanTreeNode {
    private Span span;
    private Long duration;
    private Boolean isBottleneck;
    private List<SpanTreeNode> children;
}

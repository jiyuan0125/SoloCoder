package com.depgraph.model;

import lombok.AllArgsConstructor;
import lombok.Data;
import lombok.NoArgsConstructor;

@Data
@NoArgsConstructor
@AllArgsConstructor
public class DependencyEdge {
    private String from;
    private String to;
    private long callCount = 0;

    public DependencyEdge(String from, String to) {
        this.from = from;
        this.to = to;
    }

    public void incrementCallCount() {
        this.callCount++;
    }
}

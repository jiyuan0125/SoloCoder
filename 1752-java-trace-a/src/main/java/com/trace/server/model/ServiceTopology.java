package com.trace.server.model;

import lombok.Data;
import java.util.List;

@Data
public class ServiceTopology {
    private List<ServiceNode> nodes;
    private List<ServiceEdge> edges;
}

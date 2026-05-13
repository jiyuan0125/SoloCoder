package com.servicemesh.controlplane.controller;

import com.servicemesh.common.model.TopologyData;
import com.servicemesh.controlplane.service.TopologyService;
import org.springframework.stereotype.Controller;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.ResponseBody;

@Controller
public class DashboardController {

    private final TopologyService topologyService;

    public DashboardController(TopologyService topologyService) {
        this.topologyService = topologyService;
    }

    @GetMapping("/")
    public String dashboard() {
        return "forward:/dashboard.html";
    }

    @GetMapping("/api/topology/json")
    @ResponseBody
    public TopologyData getTopologyJson() {
        return topologyService.getTopology();
    }
}

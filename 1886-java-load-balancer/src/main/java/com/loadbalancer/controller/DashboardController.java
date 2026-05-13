package com.loadbalancer.controller;

import com.loadbalancer.dto.DashboardResponse;
import com.loadbalancer.model.HealthCheckResult;
import com.loadbalancer.model.Node;
import com.loadbalancer.service.NodeRegistryService;
import lombok.RequiredArgsConstructor;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;

import java.util.ArrayList;
import java.util.List;
import java.util.stream.Collectors;

@RestController
@RequestMapping("/dashboard")
@RequiredArgsConstructor
public class DashboardController {
    private final NodeRegistryService nodeRegistryService;

    @GetMapping
    public DashboardResponse getDashboard() {
        List<DashboardResponse.NodeDashboard> nodeDashboards = new ArrayList<>();
        
        for (Node node : nodeRegistryService.getAllNodes()) {
            node.getLock().readLock().lock();
            try {
                List<DashboardResponse.HealthCheckRecord> recentChecks = new ArrayList<>();
                for (HealthCheckResult result : node.getRecentChecks()) {
                    recentChecks.add(new DashboardResponse.HealthCheckRecord(
                            result.getTimestamp().toString(),
                            result.isSuccess() ? "成功" : "失败"
                    ));
                }

                nodeDashboards.add(new DashboardResponse.NodeDashboard(
                        node.getId(),
                        node.getAddress(),
                        node.getWeight(),
                        statusToChinese(node.getStatus().name()),
                        node.getActiveConnectionsCount(),
                        recentChecks
                ));
            } finally {
                node.getLock().readLock().unlock();
            }
        }
        
        return new DashboardResponse(nodeDashboards);
    }

    private String statusToChinese(String status) {
        return switch (status) {
            case "NEW_REGISTERED" -> "新注册";
            case "NORMAL_SERVICE" -> "正常服务";
            case "SUSPECTED_FAILURE" -> "疑似故障";
            case "OFFLINE" -> "已下线";
            default -> status;
        };
    }
}

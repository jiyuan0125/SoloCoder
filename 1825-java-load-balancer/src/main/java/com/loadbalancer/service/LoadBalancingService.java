package com.loadbalancer.service;

import com.loadbalancer.model.BackendNode;
import com.loadbalancer.model.RoutingRule;
import com.loadbalancer.model.TrafficLog.SelectionReason;
import org.springframework.stereotype.Service;

import java.util.List;
import java.util.Optional;

@Service
public class LoadBalancingService {
    private final NodeManager nodeManager;
    private final RoutingRuleManager ruleManager;
    private final WeightedRoundRobinBalancer balancer;
    private final TrafficLogService trafficLogService;

    public LoadBalancingService(NodeManager nodeManager,
                                 RoutingRuleManager ruleManager,
                                 WeightedRoundRobinBalancer balancer,
                                 TrafficLogService trafficLogService) {
        this.nodeManager = nodeManager;
        this.ruleManager = ruleManager;
        this.balancer = balancer;
        this.trafficLogService = trafficLogService;
    }

    public BackendNode selectNode(String requestPath) {
        Optional<RoutingRule> matchedRule = ruleManager.findMatchingRule(requestPath);

        List<BackendNode> candidates;
        SelectionReason reason;

        if (matchedRule.isPresent()) {
            RoutingRule rule = matchedRule.get();
            candidates = nodeManager.getNodesWithTags(rule.getRequiredTags());
            reason = SelectionReason.TAG_FILTERED;
            if (!rule.getRequiredTags().isEmpty()) {
                reason = SelectionReason.TAG_FILTERED;
            } else {
                reason = SelectionReason.PATH_MATCHED;
            }
        } else {
            candidates = nodeManager.getAvailableNodes();
            reason = SelectionReason.WEIGHTED_ROUND_ROBIN;
        }

        if (candidates.isEmpty()) {
            return null;
        }

        BackendNode selected = balancer.select(candidates);
        if (selected != null) {
            trafficLogService.record(selected.getId(), reason, requestPath);
        }
        return selected;
    }
}

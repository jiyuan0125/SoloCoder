package com.loadbalancer.service;

import com.loadbalancer.model.BackendNode;
import com.loadbalancer.model.RoutingRule;
import org.springframework.stereotype.Service;

import java.util.Optional;

@Service
public class LoadBalancerService {
    private final WeightedRoundRobinService roundRobinService;
    private final RoutingRuleService routingRuleService;
    private final DistributionHistoryService historyService;

    public LoadBalancerService(WeightedRoundRobinService roundRobinService,
                              RoutingRuleService routingRuleService,
                              DistributionHistoryService historyService) {
        this.roundRobinService = roundRobinService;
        this.routingRuleService = routingRuleService;
        this.historyService = historyService;
    }

    public Optional<BackendNode> selectBackend(String requestPath) {
        Optional<RoutingRule> matchedRule = routingRuleService.matchRule(requestPath);
        
        final Optional<BackendNode> selectedNode;
        final String reason;
        final String ruleId;

        if (matchedRule.isPresent()) {
            RoutingRule rule = matchedRule.get();
            ruleId = rule.getId();
            
            if (!rule.getRequiredTags().isEmpty()) {
                selectedNode = roundRobinService.selectNodeWithTags(rule.getRequiredTags());
                reason = "path_match_with_tag_filter";
            } else {
                selectedNode = roundRobinService.selectNode();
                reason = "path_match";
            }
        } else {
            selectedNode = roundRobinService.selectNode();
            reason = "round_robin";
            ruleId = null;
        }

        if (selectedNode.isPresent()) {
            historyService.record(selectedNode.get().getId(), reason, requestPath, ruleId);
        }

        return selectedNode;
    }
}

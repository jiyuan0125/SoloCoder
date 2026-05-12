package com.gateway.filter;

import com.gateway.model.FilterChainDefinition;
import com.gateway.model.FilterResult;
import com.gateway.model.RequestContext;
import lombok.extern.slf4j.Slf4j;
import org.springframework.stereotype.Component;

import java.util.*;
import java.util.concurrent.atomic.AtomicReference;

@Slf4j
@Component
public class FilterChainManager {

    private final AtomicReference<Map<String, List<GatewayFilter>>> filterChains = new AtomicReference<>(new HashMap<>());
    private final AtomicReference<Map<String, FilterChainDefinition>> chainDefinitions = new AtomicReference<>(new HashMap<>());
    private final Map<String, GatewayFilter> filterRegistry;

    public FilterChainManager(List<GatewayFilter> filters) {
        this.filterRegistry = new HashMap<>();
        for (GatewayFilter filter : filters) {
            this.filterRegistry.put(filter.getName(), filter);
        }
        log.info("已注册过滤器: {}", filterRegistry.keySet());
    }

    public void updateFilterChain(FilterChainDefinition definition) {
        List<GatewayFilter> chain = new ArrayList<>();
        for (String filterName : definition.getFilters()) {
            GatewayFilter filter = filterRegistry.get(filterName);
            if (filter == null) {
                log.warn("未找到过滤器: {}", filterName);
                continue;
            }
            chain.add(filter);
        }

        Map<String, List<GatewayFilter>> newChains = new HashMap<>(filterChains.get());
        newChains.put(definition.getId(), Collections.unmodifiableList(chain));
        filterChains.set(newChains);

        Map<String, FilterChainDefinition> newDefinitions = new HashMap<>(chainDefinitions.get());
        newDefinitions.put(definition.getId(), definition);
        chainDefinitions.set(newDefinitions);

        log.info("已更新过滤器链: {}, 包含 {} 个过滤器", definition.getId(), chain.size());
    }

    public boolean deleteFilterChain(String chainId) {
        Map<String, List<GatewayFilter>> newChains = new HashMap<>(filterChains.get());
        boolean removed = newChains.remove(chainId) != null;
        if (removed) {
            filterChains.set(newChains);

            Map<String, FilterChainDefinition> newDefinitions = new HashMap<>(chainDefinitions.get());
            newDefinitions.remove(chainId);
            chainDefinitions.set(newDefinitions);

            log.info("已删除过滤器链: {}", chainId);
        }
        return removed;
    }

    public List<GatewayFilter> getFilterChain(String chainId) {
        List<GatewayFilter> chain = filterChains.get().get(chainId);
        return chain != null ? chain : Collections.emptyList();
    }

    public FilterChainDefinition getFilterChainDefinition(String chainId) {
        return chainDefinitions.get().get(chainId);
    }

    public Collection<FilterChainDefinition> getAllFilterChainDefinitions() {
        return chainDefinitions.get().values();
    }

    public Set<String> getAvailableFilterNames() {
        return filterRegistry.keySet();
    }

    public FilterResult executeFilterChain(String chainId, RequestContext context) {
        List<GatewayFilter> chain = getFilterChain(chainId);

        for (GatewayFilter filter : chain) {
            FilterResult result = filter.filter(context);
            if (result.getStatus() == FilterResult.Status.REJECT) {
                log.debug("过滤器 {} 拒绝请求", filter.getName());
                return result;
            }
        }

        return FilterResult.pass();
    }
}

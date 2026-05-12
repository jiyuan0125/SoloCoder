package com.gateway.controller;

import com.gateway.filter.FilterChainManager;
import com.gateway.model.FilterChainDefinition;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import javax.validation.Valid;
import java.util.Collection;
import java.util.HashMap;
import java.util.Map;
import java.util.Set;

@RestController
@RequestMapping("${gateway.admin.prefix:/admin}/filters")
public class FilterChainAdminController {

    private final FilterChainManager filterChainManager;

    public FilterChainAdminController(FilterChainManager filterChainManager) {
        this.filterChainManager = filterChainManager;
    }

    @GetMapping("/names")
    public ResponseEntity<Set<String>> listAvailableFilters() {
        return ResponseEntity.ok(filterChainManager.getAvailableFilterNames());
    }

    @GetMapping("/chains")
    public ResponseEntity<Collection<FilterChainDefinition>> listFilterChains() {
        return ResponseEntity.ok(filterChainManager.getAllFilterChainDefinitions());
    }

    @GetMapping("/chains/{id}")
    public ResponseEntity<FilterChainDefinition> getFilterChain(@PathVariable String id) {
        FilterChainDefinition definition = filterChainManager.getFilterChainDefinition(id);
        if (definition == null) {
            return ResponseEntity.notFound().build();
        }
        return ResponseEntity.ok(definition);
    }

    @PostMapping("/chains")
    public ResponseEntity<Map<String, Object>> createFilterChain(@Valid @RequestBody FilterChainDefinition definition) {
        filterChainManager.updateFilterChain(definition);
        
        Map<String, Object> result = new HashMap<>();
        result.put("message", "过滤器链已创建");
        result.put("chainId", definition.getId());
        
        return ResponseEntity.ok(result);
    }

    @PutMapping("/chains/{id}")
    public ResponseEntity<Map<String, Object>> updateFilterChain(
            @PathVariable String id,
            @Valid @RequestBody FilterChainDefinition definition) {
        if (!id.equals(definition.getId())) {
            return ResponseEntity.badRequest().build();
        }
        
        filterChainManager.updateFilterChain(definition);
        
        Map<String, Object> result = new HashMap<>();
        result.put("message", "过滤器链已更新");
        result.put("chainId", definition.getId());
        
        return ResponseEntity.ok(result);
    }

    @DeleteMapping("/chains/{id}")
    public ResponseEntity<Map<String, Object>> deleteFilterChain(@PathVariable String id) {
        boolean deleted = filterChainManager.deleteFilterChain(id);
        
        Map<String, Object> result = new HashMap<>();
        if (deleted) {
            result.put("message", "过滤器链已删除");
            return ResponseEntity.ok(result);
        } else {
            result.put("message", "过滤器链不存在");
            return ResponseEntity.notFound().build();
        }
    }
}

package com.example.converter.service;

import com.example.converter.exception.MappingNotFoundException;
import com.example.converter.model.MappingConfig;
import org.springframework.stereotype.Service;

import java.util.ArrayList;
import java.util.List;
import java.util.Map;
import java.util.concurrent.ConcurrentHashMap;

@Service
public class MappingService {

    private final Map<String, MappingConfig> mappings = new ConcurrentHashMap<>();

    public MappingConfig createMapping(MappingConfig config) {
        mappings.put(config.getName(), config);
        return config;
    }

    public List<MappingConfig> getAllMappings() {
        return new ArrayList<>(mappings.values());
    }

    public MappingConfig getMapping(String name) {
        MappingConfig config = mappings.get(name);
        if (config == null) {
            throw new MappingNotFoundException("Mapping not found: " + name);
        }
        return config;
    }

    public boolean exists(String name) {
        return mappings.containsKey(name);
    }
}

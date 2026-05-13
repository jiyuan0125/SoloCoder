package com.example.converter.controller;

import com.example.converter.model.MappingConfig;
import com.example.converter.service.MappingService;
import lombok.RequiredArgsConstructor;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import java.util.List;

@RestController
@RequestMapping("/mappings")
@RequiredArgsConstructor
public class MappingController {

    private final MappingService mappingService;

    @PostMapping
    public ResponseEntity<MappingConfig> createMapping(@RequestBody MappingConfig config) {
        MappingConfig created = mappingService.createMapping(config);
        return ResponseEntity.status(HttpStatus.CREATED).body(created);
    }

    @GetMapping
    public ResponseEntity<List<MappingConfig>> getAllMappings() {
        return ResponseEntity.ok(mappingService.getAllMappings());
    }

    @GetMapping("/{name}")
    public ResponseEntity<MappingConfig> getMapping(@PathVariable String name) {
        return ResponseEntity.ok(mappingService.getMapping(name));
    }
}

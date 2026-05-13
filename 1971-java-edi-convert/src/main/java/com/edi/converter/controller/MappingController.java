package com.edi.converter.controller;

import com.edi.converter.mapping.MappingRule;
import com.edi.converter.mapping.MappingService;
import org.springframework.beans.factory.annotation.Autowired;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import java.util.List;
import java.util.Optional;

@RestController
@RequestMapping("/mappings")
public class MappingController {

    private final MappingService mappingService;

    @Autowired
    public MappingController(MappingService mappingService) {
        this.mappingService = mappingService;
    }

    @PostMapping
    public ResponseEntity<MappingRule> createMapping(@RequestBody MappingRule mapping) {
        MappingRule saved = mappingService.saveMapping(mapping);
        return ResponseEntity.ok(saved);
    }

    @GetMapping
    public ResponseEntity<List<MappingRule>> getAllMappings() {
        return ResponseEntity.ok(mappingService.getAllMappings());
    }

    @GetMapping("/{partnerId}")
    public ResponseEntity<MappingRule> getMapping(@PathVariable String partnerId) {
        Optional<MappingRule> mapping = mappingService.getMapping(partnerId);
        return mapping.map(ResponseEntity::ok).orElse(ResponseEntity.notFound().build());
    }

    @PutMapping("/{partnerId}")
    public ResponseEntity<MappingRule> updateMapping(@PathVariable String partnerId, @RequestBody MappingRule mapping) {
        if (!partnerId.equals(mapping.getPartnerId())) {
            mapping.setPartnerId(partnerId);
        }
        MappingRule saved = mappingService.saveMapping(mapping);
        return ResponseEntity.ok(saved);
    }

    @DeleteMapping("/{partnerId}")
    public ResponseEntity<Void> deleteMapping(@PathVariable String partnerId) {
        mappingService.deleteMapping(partnerId);
        return ResponseEntity.noContent().build();
    }
}

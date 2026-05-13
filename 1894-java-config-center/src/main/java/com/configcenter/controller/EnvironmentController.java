package com.configcenter.controller;

import com.configcenter.dto.EnvironmentRequest;
import com.configcenter.entity.Environment;
import com.configcenter.service.EnvironmentService;
import jakarta.validation.Valid;
import lombok.RequiredArgsConstructor;
import org.springframework.http.HttpStatus;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import java.util.List;

@RestController
@RequestMapping("/api/environments")
@RequiredArgsConstructor
public class EnvironmentController {

    private final EnvironmentService environmentService;

    @GetMapping
    public ResponseEntity<List<Environment>> getAllEnvironments() {
        return ResponseEntity.ok(environmentService.getAllEnvironments());
    }

    @GetMapping("/{id}")
    public ResponseEntity<Environment> getEnvironmentById(@PathVariable Long id) {
        return ResponseEntity.ok(environmentService.getEnvironmentById(id));
    }

    @GetMapping("/name/{name}")
    public ResponseEntity<Environment> getEnvironmentByName(@PathVariable String name) {
        return ResponseEntity.ok(environmentService.getEnvironmentByName(name));
    }

    @PostMapping
    public ResponseEntity<Environment> createEnvironment(@Valid @RequestBody EnvironmentRequest request) {
        Environment created = environmentService.createEnvironment(request);
        return new ResponseEntity<>(created, HttpStatus.CREATED);
    }

    @PutMapping("/{id}")
    public ResponseEntity<Environment> updateEnvironment(@PathVariable Long id, @Valid @RequestBody EnvironmentRequest request) {
        return ResponseEntity.ok(environmentService.updateEnvironment(id, request));
    }

    @DeleteMapping("/{id}")
    public ResponseEntity<Void> deleteEnvironment(@PathVariable Long id) {
        environmentService.deleteEnvironment(id);
        return ResponseEntity.noContent().build();
    }
}

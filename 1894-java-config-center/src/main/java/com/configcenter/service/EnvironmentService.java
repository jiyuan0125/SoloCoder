package com.configcenter.service;

import com.configcenter.dto.EnvironmentRequest;
import com.configcenter.entity.Environment;
import com.configcenter.exception.ResourceNotFoundException;
import com.configcenter.exception.ValidationException;
import com.configcenter.repository.EnvironmentRepository;
import lombok.RequiredArgsConstructor;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.util.Arrays;
import java.util.List;

@Service
@RequiredArgsConstructor
public class EnvironmentService {

    private static final List<String> ALLOWED_ENVIRONMENTS = Arrays.asList("dev", "staging", "prod");

    private final EnvironmentRepository environmentRepository;

    public List<Environment> getAllEnvironments() {
        return environmentRepository.findAll();
    }

    public Environment getEnvironmentById(Long id) {
        return environmentRepository.findById(id)
                .orElseThrow(() -> new ResourceNotFoundException("Environment not found with id: " + id));
    }

    public Environment getEnvironmentByName(String name) {
        return environmentRepository.findByName(name)
                .orElseThrow(() -> new ResourceNotFoundException("Environment not found with name: " + name));
    }

    @Transactional
    public Environment createEnvironment(EnvironmentRequest request) {
        String name = request.getName();
        validateEnvironmentName(name);
        if (environmentRepository.existsByName(name)) {
            throw new ValidationException("Environment with name '" + name + "' already exists");
        }
        Environment env = new Environment();
        env.setName(name);
        env.setDescription(request.getDescription());
        return environmentRepository.save(env);
    }

    @Transactional
    public Environment updateEnvironment(Long id, EnvironmentRequest request) {
        Environment env = getEnvironmentById(id);
        String name = request.getName();
        validateEnvironmentName(name);
        if (!env.getName().equals(name) && environmentRepository.existsByName(name)) {
            throw new ValidationException("Environment with name '" + name + "' already exists");
        }
        env.setName(name);
        env.setDescription(request.getDescription());
        return environmentRepository.save(env);
    }

    @Transactional
    public void deleteEnvironment(Long id) {
        if (!environmentRepository.existsById(id)) {
            throw new ResourceNotFoundException("Environment not found with id: " + id);
        }
        environmentRepository.deleteById(id);
    }

    public boolean existsById(Long id) {
        return environmentRepository.existsById(id);
    }

    private void validateEnvironmentName(String name) {
        if (!ALLOWED_ENVIRONMENTS.contains(name)) {
            throw new ValidationException("Environment name must be one of: " + ALLOWED_ENVIRONMENTS);
        }
    }
}

package com.configcenter.service;

import com.configcenter.dto.ApplicationRequest;
import com.configcenter.entity.Application;
import com.configcenter.exception.ResourceNotFoundException;
import com.configcenter.exception.ValidationException;
import com.configcenter.repository.ApplicationRepository;
import lombok.RequiredArgsConstructor;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.util.List;

@Service
@RequiredArgsConstructor
public class ApplicationService {

    private final ApplicationRepository applicationRepository;

    public List<Application> getAllApplications() {
        return applicationRepository.findAll();
    }

    public Application getApplicationById(Long id) {
        return applicationRepository.findById(id)
                .orElseThrow(() -> new ResourceNotFoundException("Application not found with id: " + id));
    }

    public Application getApplicationByName(String name) {
        return applicationRepository.findByName(name)
                .orElseThrow(() -> new ResourceNotFoundException("Application not found with name: " + name));
    }

    @Transactional
    public Application createApplication(ApplicationRequest request) {
        if (applicationRepository.existsByName(request.getName())) {
            throw new ValidationException("Application with name '" + request.getName() + "' already exists");
        }
        Application app = new Application();
        app.setName(request.getName());
        app.setDescription(request.getDescription());
        return applicationRepository.save(app);
    }

    @Transactional
    public Application updateApplication(Long id, ApplicationRequest request) {
        Application app = getApplicationById(id);
        if (!app.getName().equals(request.getName()) && 
            applicationRepository.existsByName(request.getName())) {
            throw new ValidationException("Application with name '" + request.getName() + "' already exists");
        }
        app.setName(request.getName());
        app.setDescription(request.getDescription());
        return applicationRepository.save(app);
    }

    @Transactional
    public void deleteApplication(Long id) {
        if (!applicationRepository.existsById(id)) {
            throw new ResourceNotFoundException("Application not found with id: " + id);
        }
        applicationRepository.deleteById(id);
    }

    public boolean existsById(Long id) {
        return applicationRepository.existsById(id);
    }
}

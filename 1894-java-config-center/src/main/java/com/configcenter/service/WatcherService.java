package com.configcenter.service;

import com.configcenter.dto.WatcherRequest;
import com.configcenter.entity.Watcher;
import com.configcenter.exception.ResourceNotFoundException;
import com.configcenter.exception.ValidationException;
import com.configcenter.repository.WatcherRepository;
import lombok.RequiredArgsConstructor;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.util.List;

@Service
@RequiredArgsConstructor
public class WatcherService {

    private final WatcherRepository watcherRepository;
    private final ApplicationService applicationService;

    public List<Watcher> getWatchersByApplication(Long applicationId) {
        if (!applicationService.existsById(applicationId)) {
            throw new ResourceNotFoundException("Application not found with id: " + applicationId);
        }
        return watcherRepository.findByApplicationIdAndIsActiveTrue(applicationId);
    }

    @Transactional
    public Watcher registerWatcher(Long applicationId, WatcherRequest request) {
        if (!applicationService.existsById(applicationId)) {
            throw new ResourceNotFoundException("Application not found with id: " + applicationId);
        }
        
        validateCallbackUrl(request.getCallbackUrl());
        
        return watcherRepository.findByApplicationIdAndCallbackUrl(applicationId, request.getCallbackUrl())
                .map(existing -> {
                    existing.setIsActive(true);
                    existing.setConsecutiveFailures(0);
                    return watcherRepository.save(existing);
                })
                .orElseGet(() -> {
                    Watcher watcher = new Watcher();
                    watcher.setApplicationId(applicationId);
                    watcher.setCallbackUrl(request.getCallbackUrl());
                    watcher.setConsecutiveFailures(0);
                    watcher.setIsActive(true);
                    return watcherRepository.save(watcher);
                });
    }

    @Transactional
    public void unregisterWatcher(Long applicationId, WatcherRequest request) {
        watcherRepository.findByApplicationIdAndCallbackUrl(applicationId, request.getCallbackUrl())
                .ifPresent(watcher -> {
                    watcher.setIsActive(false);
                    watcherRepository.save(watcher);
                });
    }

    @Transactional
    public void markWatcherSuccess(Watcher watcher) {
        watcher.setConsecutiveFailures(0);
        watcher.setIsActive(true);
        watcherRepository.save(watcher);
    }

    @Transactional
    public void markWatcherFailure(Watcher watcher, int maxConsecutiveFailures) {
        watcher.setConsecutiveFailures(watcher.getConsecutiveFailures() + 1);
        if (watcher.getConsecutiveFailures() >= maxConsecutiveFailures) {
            watcher.setIsActive(false);
        }
        watcherRepository.save(watcher);
    }

    private void validateCallbackUrl(String url) {
        if (url == null || url.isEmpty()) {
            throw new ValidationException("Callback URL cannot be empty");
        }
        if (!url.startsWith("http://") && !url.startsWith("https://")) {
            throw new ValidationException("Callback URL must start with http:// or https://");
        }
    }
}

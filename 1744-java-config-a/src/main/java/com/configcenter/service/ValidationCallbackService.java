package com.configcenter.service;

import com.configcenter.model.Release;
import com.configcenter.model.ReleaseConfigChange;
import com.configcenter.repository.ReleaseConfigChangeRepository;
import com.fasterxml.jackson.databind.ObjectMapper;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.http.ResponseEntity;
import org.springframework.stereotype.Service;
import org.springframework.web.client.RestTemplate;

import java.util.HashMap;
import java.util.List;
import java.util.Map;

@Slf4j
@Service
@RequiredArgsConstructor
public class ValidationCallbackService {

    private final ReleaseConfigChangeRepository configChangeRepository;
    private final RestTemplate restTemplate;
    private final ObjectMapper objectMapper;

    public String validateRelease(Release release) {
        String callbackUrl = release.getProject().getValidationCallbackUrl();
        
        if (callbackUrl == null || callbackUrl.trim().isEmpty()) {
            log.info("No validation callback URL configured for project {}. Skipping validation.", 
                    release.getProject().getName());
            return "SKIPPED - No callback URL configured";
        }
        
        try {
            List<ReleaseConfigChange> configChanges = configChangeRepository.findByReleaseId(release.getId());
            
            Map<String, Object> payload = buildValidationPayload(release, configChanges);
            
            log.info("Calling validation callback for release {}: {}", release.getId(), callbackUrl);
            
            ResponseEntity<String> response = restTemplate.postForEntity(callbackUrl, payload, String.class);
            
            boolean success = response.getStatusCode().is2xxSuccessful();
            
            String result = success ? "SUCCESS" : "FAILED";
            log.info("Validation callback result for release {}: {} - Status: {}, Body: {}", 
                    release.getId(), result, response.getStatusCode(), response.getBody());
            
            if (!success) {
                log.error("Validation callback failed for release {}. Status: {}, But continuing with release process.", 
                        release.getId(), response.getStatusCode());
            }
            
            return result + " - Status: " + response.getStatusCode();
            
        } catch (Exception e) {
            log.error("Validation callback failed for release {}. But continuing with release process.", 
                    release.getId(), e);
            return "FAILED - Exception: " + e.getMessage();
        }
    }

    private Map<String, Object> buildValidationPayload(Release release, List<ReleaseConfigChange> configChanges) {
        Map<String, Object> payload = new HashMap<>();
        payload.put("releaseId", release.getId());
        payload.put("projectId", release.getProject().getId());
        payload.put("projectName", release.getProject().getName());
        payload.put("environment", release.getEnvironment());
        payload.put("releaseType", release.getType().name());
        
        Map<String, String> changes = new HashMap<>();
        for (ReleaseConfigChange change : configChanges) {
            changes.put(change.getConfigKey(), change.getNewValue());
        }
        payload.put("configChanges", changes);
        
        Map<String, String> oldValues = new HashMap<>();
        for (ReleaseConfigChange change : configChanges) {
            oldValues.put(change.getConfigKey(), change.getOldValue());
        }
        payload.put("oldValues", oldValues);
        
        return payload;
    }
}

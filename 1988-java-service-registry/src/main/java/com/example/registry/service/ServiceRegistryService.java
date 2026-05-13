package com.example.registry.service;

import com.example.registry.config.RegistryProperties;
import com.example.registry.dto.DiscoveryResponse;
import com.example.registry.model.ServiceInstance;
import com.example.registry.model.Subscriber;
import com.example.registry.repository.ServiceRegistryRepository;
import org.springframework.stereotype.Service;

import java.util.List;
import java.util.Map;

@Service
public class ServiceRegistryService {

    private final ServiceRegistryRepository repository;
    private final RegistryProperties properties;
    private final NotificationService notificationService;

    public ServiceRegistryService(ServiceRegistryRepository repository,
                                  RegistryProperties properties,
                                  NotificationService notificationService) {
        this.repository = repository;
        this.properties = properties;
        this.notificationService = notificationService;
    }

    public ServiceInstance register(String serviceName, String ip, int port, 
                                    Map<String, String> metadata) {
        return repository.register(serviceName, ip, port, metadata);
    }

    public DiscoveryResponse discover(String serviceName, Map<String, String> metadataFilter, 
                                       Integer page) {
        List<ServiceInstance> instances = metadataFilter != null && !metadataFilter.isEmpty()
            ? repository.findByMetadata(serviceName, metadataFilter)
            : repository.findAllInstancesExcludingGoingDown(serviceName);

        int total = instances.size();
        int pageSize = properties.getDiscovery().getPageSize();
        
        if (total <= properties.getDiscovery().getPaginationThreshold() || page == null) {
            return DiscoveryResponse.builder()
                .instances(instances)
                .total(total)
                .page(1)
                .pageSize(total)
                .totalPages(1)
                .build();
        }

        int pageNum = Math.max(1, page);
        int totalPages = (int) Math.ceil((double) total / pageSize);
        int fromIndex = (pageNum - 1) * pageSize;
        int toIndex = Math.min(fromIndex + pageSize, total);
        
        List<ServiceInstance> pageContent = instances.subList(fromIndex, toIndex);
        
        return DiscoveryResponse.builder()
            .instances(pageContent)
            .total(total)
            .page(pageNum)
            .pageSize(pageSize)
            .totalPages(totalPages)
            .build();
    }

    public boolean updateMetadata(String serviceName, String instanceId, 
                                  Map<String, String> metadata) {
        ServiceInstance instance = repository.findInstance(serviceName, instanceId)
            .orElse(null);
        
        if (instance == null) {
            return false;
        }

        instance.setMetadata(metadata);
        repository.updateInstance(instance);
        
        notificationService.notifyMetadataChange(serviceName, instance);
        return true;
    }

    public Subscriber subscribe(String serviceName, String callbackUrl) {
        return repository.subscribe(serviceName, callbackUrl);
    }

    public void unsubscribe(String serviceName, String callbackUrl) {
        repository.unsubscribe(serviceName, callbackUrl);
    }
}

package com.configcenter.repository;

import com.configcenter.model.ConfigItem;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.data.jpa.repository.JpaSpecificationExecutor;
import org.springframework.stereotype.Repository;

import java.util.Optional;

@Repository
public interface ConfigItemRepository extends JpaRepository<ConfigItem, Long>, JpaSpecificationExecutor<ConfigItem> {
    
    Optional<ConfigItem> findByNamespaceAndGroupAndKey(String namespace, String group, String key);
    
    boolean existsByNamespaceAndGroupAndKey(String namespace, String group, String key);
    
    void deleteByNamespaceAndGroupAndKey(String namespace, String group, String key);
}

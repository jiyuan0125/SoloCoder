package com.configcenter.repository;

import com.configcenter.model.ConfigHistory;
import org.springframework.data.domain.Page;
import org.springframework.data.domain.Pageable;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.data.jpa.repository.JpaSpecificationExecutor;
import org.springframework.stereotype.Repository;

import java.util.Optional;

@Repository
public interface ConfigHistoryRepository extends JpaRepository<ConfigHistory, Long>, JpaSpecificationExecutor<ConfigHistory> {
    
    Page<ConfigHistory> findByNamespaceAndGroupAndKeyOrderByVersionDesc(String namespace, String group, String key, Pageable pageable);
    
    Optional<ConfigHistory> findByNamespaceAndGroupAndKeyAndVersion(String namespace, String group, String key, Long version);
    
    Optional<ConfigHistory> findFirstByNamespaceAndGroupAndKeyOrderByVersionDesc(String namespace, String group, String key);
}

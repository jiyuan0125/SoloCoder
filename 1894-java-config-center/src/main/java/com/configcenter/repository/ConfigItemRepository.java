package com.configcenter.repository;

import com.configcenter.entity.ConfigItem;
import org.springframework.data.domain.Page;
import org.springframework.data.domain.Pageable;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.stereotype.Repository;
import java.util.Optional;

@Repository
public interface ConfigItemRepository extends JpaRepository<ConfigItem, Long> {
    Optional<ConfigItem> findByApplicationIdAndEnvironmentIdAndConfigKey(Long applicationId, Long environmentId, String configKey);
    Page<ConfigItem> findByApplicationIdAndEnvironmentId(Long applicationId, Long environmentId, Pageable pageable);
    boolean existsByApplicationIdAndEnvironmentIdAndConfigKey(Long applicationId, Long environmentId, String configKey);
}

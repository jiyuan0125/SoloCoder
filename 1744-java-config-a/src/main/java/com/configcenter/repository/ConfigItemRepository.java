package com.configcenter.repository;

import com.configcenter.model.ConfigItem;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.stereotype.Repository;

import java.util.List;
import java.util.Optional;

@Repository
public interface ConfigItemRepository extends JpaRepository<ConfigItem, Long> {
    Optional<ConfigItem> findByProjectIdAndEnvironmentAndConfigKey(Long projectId, String environment, String configKey);
    List<ConfigItem> findByProjectIdAndEnvironment(Long projectId, String environment);
    List<ConfigItem> findByProjectIdAndEnvironmentAndStatusIn(Long projectId, String environment, List<ConfigItem.ReleaseStatus> statuses);
    List<ConfigItem> findByProjectIdAndEnvironmentAndStatus(Long projectId, String environment, ConfigItem.ReleaseStatus status);
}

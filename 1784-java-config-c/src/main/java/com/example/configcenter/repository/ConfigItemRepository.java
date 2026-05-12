package com.example.configcenter.repository;

import com.example.configcenter.entity.ConfigItem;
import org.springframework.data.jpa.repository.JpaRepository;

import java.util.Optional;

public interface ConfigItemRepository extends JpaRepository<ConfigItem, Long> {
    Optional<ConfigItem> findByConfigKey(String configKey);
    boolean existsByConfigKey(String configKey);
    void deleteByConfigKey(String configKey);
}

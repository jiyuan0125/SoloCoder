package com.example.configcenter.repository;

import com.example.configcenter.entity.ConfigSubscriber;
import org.springframework.data.jpa.repository.JpaRepository;

import java.util.List;

public interface ConfigSubscriberRepository extends JpaRepository<ConfigSubscriber, Long> {
    List<ConfigSubscriber> findByConfigKey(String configKey);
    void deleteByConfigKey(String configKey);
}

package com.configcenter.repository;

import com.configcenter.entity.Watcher;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.stereotype.Repository;
import java.util.List;
import java.util.Optional;

@Repository
public interface WatcherRepository extends JpaRepository<Watcher, Long> {
    List<Watcher> findByApplicationIdAndIsActiveTrue(Long applicationId);
    Optional<Watcher> findByApplicationIdAndCallbackUrl(Long applicationId, String callbackUrl);
    boolean existsByApplicationIdAndCallbackUrl(Long applicationId, String callbackUrl);
}

package com.configcenter.repository;

import com.configcenter.model.ClientSubscription;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.data.jpa.repository.Modifying;
import org.springframework.data.jpa.repository.Query;
import org.springframework.stereotype.Repository;

import java.time.LocalDateTime;
import java.util.List;
import java.util.Optional;

@Repository
public interface ClientSubscriptionRepository extends JpaRepository<ClientSubscription, Long> {
    Optional<ClientSubscription> findByInstanceIdAndProjectIdAndEnvironment(String instanceId, Long projectId, String environment);
    List<ClientSubscription> findByProjectIdAndEnvironment(Long projectId, String environment);
    
    @Modifying
    @Query("UPDATE ClientSubscription c SET c.lastHeartbeat = :time WHERE c.id = :id")
    void updateLastHeartbeat(Long id, LocalDateTime time);
}

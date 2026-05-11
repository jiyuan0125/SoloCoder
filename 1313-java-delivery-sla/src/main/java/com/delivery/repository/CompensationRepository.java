package com.delivery.repository;

import com.delivery.entity.Compensation;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.stereotype.Repository;
import java.util.Optional;

@Repository
public interface CompensationRepository extends JpaRepository<Compensation, Long> {
    Optional<Compensation> findByOrderId(Long orderId);
}

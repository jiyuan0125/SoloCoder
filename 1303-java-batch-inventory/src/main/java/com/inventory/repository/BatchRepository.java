package com.inventory.repository;

import com.inventory.entity.Batch;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.data.jpa.repository.Query;
import org.springframework.data.repository.query.Param;
import org.springframework.stereotype.Repository;
import java.util.List;
import java.util.Optional;

@Repository
public interface BatchRepository extends JpaRepository<Batch, Long> {
    Optional<Batch> findByProductIdAndBatchNumber(Long productId, String batchNumber);

    @Query("SELECT b FROM Batch b WHERE b.product.id = :productId ORDER BY b.inboundDate ASC, b.id ASC")
    List<Batch> findByProductIdOrderByInboundDateAsc(@Param("productId") Long productId);

    boolean existsByProductIdAndBatchNumber(Long productId, String batchNumber);
}

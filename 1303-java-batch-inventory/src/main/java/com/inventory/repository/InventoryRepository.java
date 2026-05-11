package com.inventory.repository;

import com.inventory.entity.Inventory;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.data.jpa.repository.Query;
import org.springframework.data.repository.query.Param;
import org.springframework.stereotype.Repository;
import java.util.List;
import java.util.Optional;

@Repository
public interface InventoryRepository extends JpaRepository<Inventory, Long> {
    Optional<Inventory> findByBatchId(Long batchId);

    @Query("SELECT i FROM Inventory i JOIN i.batch b WHERE i.product.id = :productId AND i.currentQuantity > 0 ORDER BY b.inboundDate ASC, b.id ASC")
    List<Inventory> findAvailableByProductIdOrderByInboundDateAsc(@Param("productId") Long productId);

    @Query("SELECT i FROM Inventory i WHERE i.currentQuantity > 0 ORDER BY i.product.id ASC, i.batch.inboundDate ASC")
    List<Inventory> findAllAvailable();

    @Query("SELECT SUM(i.currentQuantity) FROM Inventory i WHERE i.product.id = :productId")
    Integer sumCurrentQuantityByProductId(@Param("productId") Long productId);
}

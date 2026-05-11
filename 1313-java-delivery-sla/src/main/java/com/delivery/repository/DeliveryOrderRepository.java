package com.delivery.repository;

import com.delivery.entity.DeliveryOrder;
import com.delivery.enums.OrderStatus;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.data.jpa.repository.Query;
import org.springframework.data.repository.query.Param;
import org.springframework.stereotype.Repository;
import java.time.LocalDateTime;
import java.util.List;
import java.util.Optional;

@Repository
public interface DeliveryOrderRepository extends JpaRepository<DeliveryOrder, Long> {
    Optional<DeliveryOrder> findByOrderNo(String orderNo);
    
    List<DeliveryOrder> findByCustomerId(Long customerId);
    
    List<DeliveryOrder> findByStatus(OrderStatus status);
    
    @Query("SELECT COUNT(o) FROM DeliveryOrder o WHERE o.customer.id = :customerId " +
           "AND o.orderTime >= :startOfMonth AND o.orderTime < :endOfMonth " +
           "AND o.status != 'CANCELLED'")
    int countMonthlyOrders(@Param("customerId") Long customerId,
                           @Param("startOfMonth") LocalDateTime startOfMonth,
                           @Param("endOfMonth") LocalDateTime endOfMonth);
    
    @Query("SELECT o FROM DeliveryOrder o WHERE o.deliveryPerson.id = :deliveryPersonId " +
           "AND o.status IN ('DISPATCHED', 'IN_TRANSIT')")
    List<DeliveryOrder> findActiveOrdersByDeliveryPerson(@Param("deliveryPersonId") Long deliveryPersonId);
    
    @Query("SELECT o FROM DeliveryOrder o WHERE o.status = 'IN_TRANSIT' " +
           "AND o.promisedDeliveryTime < :now")
    List<DeliveryOrder> findDelayedOrders(@Param("now") LocalDateTime now);
}

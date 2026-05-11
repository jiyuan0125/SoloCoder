package com.delivery.repository;

import com.delivery.entity.DeliveryPerson;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.stereotype.Repository;
import java.util.List;
import java.util.Optional;

@Repository
public interface DeliveryPersonRepository extends JpaRepository<DeliveryPerson, Long> {
    Optional<DeliveryPerson> findByEmployeeCode(String employeeCode);
    List<DeliveryPerson> findByActiveTrue();
    List<DeliveryPerson> findByActiveTrueAndCurrentZoneCode(String zoneCode);
}

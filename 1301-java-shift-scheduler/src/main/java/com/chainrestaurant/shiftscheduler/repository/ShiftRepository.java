package com.chainrestaurant.shiftscheduler.repository;

import com.chainrestaurant.shiftscheduler.entity.Shift;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.stereotype.Repository;

import java.util.Optional;

@Repository
public interface ShiftRepository extends JpaRepository<Shift, Long> {
    
    Optional<Shift> findByName(String name);
    
    boolean existsByName(String name);
}

package com.company.vehicledispatch.repository;

import com.company.vehicledispatch.entity.DispatchRequest;
import com.company.vehicledispatch.enums.RequestStatus;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.data.jpa.repository.Query;
import org.springframework.data.repository.query.Param;
import org.springframework.stereotype.Repository;

import java.time.LocalDateTime;
import java.util.List;

@Repository
public interface DispatchRequestRepository extends JpaRepository<DispatchRequest, Long> {
    List<DispatchRequest> findByStatus(RequestStatus status);
    
    List<DispatchRequest> findByAssignedVehicleId(Long vehicleId);
    
    List<DispatchRequest> findByDepartmentId(Long departmentId);
    
    @Query("SELECT dr FROM DispatchRequest dr WHERE dr.assignedVehicle.id = :vehicleId " +
           "AND dr.status IN ('PENDING', 'APPROVED') " +
           "AND ((dr.startDateTime < :endDateTime AND dr.endDateTime > :startDateTime) " +
           "OR (dr.startDateTime = :endDateTime))")
    List<DispatchRequest> findConflictingRequests(@Param("vehicleId") Long vehicleId,
                                                   @Param("startDateTime") LocalDateTime startDateTime,
                                                   @Param("endDateTime") LocalDateTime endDateTime);
    
    @Query("SELECT dr FROM DispatchRequest dr WHERE dr.assignedVehicle.id = :vehicleId " +
           "AND dr.status IN ('PENDING', 'APPROVED') " +
           "AND dr.department.id = :departmentId " +
           "AND dr.endDateTime = :startDateTime")
    List<DispatchRequest> findContinuousSameDepartmentRequests(@Param("vehicleId") Long vehicleId,
                                                                 @Param("departmentId") Long departmentId,
                                                                 @Param("startDateTime") LocalDateTime startDateTime);
    
    @Query("SELECT dr FROM DispatchRequest dr WHERE dr.assignedVehicle.id = :vehicleId " +
           "AND dr.status IN ('PENDING', 'APPROVED') " +
           "AND dr.endDateTime <= :endDateTime " +
           "AND dr.endDateTime > :bufferStart")
    List<DispatchRequest> findRequestsWithinBuffer(@Param("vehicleId") Long vehicleId,
                                                    @Param("endDateTime") LocalDateTime endDateTime,
                                                    @Param("bufferStart") LocalDateTime bufferStart);
}

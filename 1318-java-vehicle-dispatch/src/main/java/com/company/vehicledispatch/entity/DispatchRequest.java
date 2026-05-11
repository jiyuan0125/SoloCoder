package com.company.vehicledispatch.entity;

import com.company.vehicledispatch.enums.RequestStatus;
import lombok.Data;

import javax.persistence.*;
import java.time.LocalDateTime;

@Data
@Entity
@Table(name = "dispatch_requests")
public class DispatchRequest {
    @Id
    @GeneratedValue(strategy = GenerationType.IDENTITY)
    private Long id;

    @ManyToOne(fetch = FetchType.LAZY)
    @JoinColumn(name = "department_id", nullable = false)
    private Department department;

    @Column(nullable = false)
    private LocalDateTime startDateTime;

    @Column(nullable = false)
    private LocalDateTime endDateTime;

    @Column(nullable = false)
    private String purpose;

    @Column(nullable = false)
    private String destination;

    @Column(nullable = false)
    private Integer requiredSeats;

    private Integer estimatedDistance;

    @ManyToOne(fetch = FetchType.LAZY)
    @JoinColumn(name = "vehicle_id")
    private Vehicle assignedVehicle;

    @Enumerated(EnumType.STRING)
    @Column(nullable = false)
    private RequestStatus status;

    private String remarks;

    private Boolean sameDepartmentContinuous;

    private Boolean dispatchConfirmed;

    private LocalDateTime createdAt;

    private LocalDateTime updatedAt;
}

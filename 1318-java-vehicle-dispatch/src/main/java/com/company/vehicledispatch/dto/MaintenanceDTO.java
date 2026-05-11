package com.company.vehicledispatch.dto;

import com.company.vehicledispatch.enums.MaintenanceType;
import lombok.Data;

import javax.validation.constraints.NotNull;
import java.time.LocalDate;

@Data
public class MaintenanceDTO {
    private Long id;

    @NotNull(message = "车辆ID不能为空")
    private Long vehicleId;

    @NotNull(message = "维修类型不能为空")
    private MaintenanceType type;

    private String description;

    @NotNull(message = "计划日期不能为空")
    private LocalDate scheduledDate;

    private LocalDate startDate;

    private LocalDate endDate;

    private Double cost;

    private String status;

    private String remarks;
}

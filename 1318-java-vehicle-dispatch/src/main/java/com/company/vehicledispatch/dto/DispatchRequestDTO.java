package com.company.vehicledispatch.dto;

import lombok.Data;

import javax.validation.constraints.NotBlank;
import javax.validation.constraints.NotNull;
import javax.validation.constraints.Positive;
import java.time.LocalDateTime;

@Data
public class DispatchRequestDTO {
    private Long id;

    @NotNull(message = "部门ID不能为空")
    private Long departmentId;

    @NotNull(message = "开始时间不能为空")
    private LocalDateTime startDateTime;

    @NotNull(message = "结束时间不能为空")
    private LocalDateTime endDateTime;

    @NotBlank(message = "用车事由不能为空")
    private String purpose;

    @NotBlank(message = "目的地不能为空")
    private String destination;

    @NotNull(message = "需要座位数不能为空")
    @Positive(message = "需要座位数必须大于0")
    private Integer requiredSeats;

    private Integer estimatedDistance;

    private Long assignedVehicleId;

    private String remarks;

    private Boolean sameDepartmentContinuous;

    private Boolean dispatchConfirmed;
}

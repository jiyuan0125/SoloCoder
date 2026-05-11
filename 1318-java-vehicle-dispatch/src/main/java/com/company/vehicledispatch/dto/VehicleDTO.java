package com.company.vehicledispatch.dto;

import com.company.vehicledispatch.enums.FuelType;
import com.company.vehicledispatch.enums.VehicleStatus;
import lombok.Data;

import javax.validation.constraints.NotBlank;
import javax.validation.constraints.NotNull;
import javax.validation.constraints.Positive;
import java.time.LocalDate;

@Data
public class VehicleDTO {
    private Long id;

    @NotBlank(message = "车牌号不能为空")
    private String plateNumber;

    @NotBlank(message = "车型不能为空")
    private String model;

    @NotNull(message = "座位数不能为空")
    @Positive(message = "座位数必须大于0")
    private Integer seats;

    @NotNull(message = "燃油类型不能为空")
    private FuelType fuelType;

    @NotNull(message = "当前油量不能为空")
    private Double fuelLevel;

    @NotNull(message = "最大油量不能为空")
    private Double maxFuelLevel;

    @NotNull(message = "当前里程数不能为空")
    private Double currentMileage;

    private Double lastMaintenanceMileage;

    @NotNull(message = "保险到期日不能为空")
    private LocalDate insuranceExpiryDate;

    @NotNull(message = "年检到期日不能为空")
    private LocalDate annualInspectionExpiryDate;

    private VehicleStatus status;

    private Double averageFuelConsumption;
}

package com.company.vehicledispatch.dto;

import com.company.vehicledispatch.entity.DispatchRequest;
import com.company.vehicledispatch.entity.Vehicle;
import lombok.Builder;
import lombok.Data;

import java.util.List;

@Data
@Builder
public class DispatchResultDTO {
    private boolean success;
    private String message;
    private DispatchRequest request;
    private Vehicle assignedVehicle;
    private List<String> warnings;
    private List<Vehicle> availableVehicles;
}

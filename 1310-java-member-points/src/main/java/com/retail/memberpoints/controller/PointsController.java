package com.retail.memberpoints.controller;

import com.retail.memberpoints.dto.ApiResponse;
import com.retail.memberpoints.dto.EarnPointsRequest;
import com.retail.memberpoints.dto.RefundRequest;
import com.retail.memberpoints.dto.UsePointsRequest;
import com.retail.memberpoints.entity.PointsRecord;
import com.retail.memberpoints.service.PointsService;
import jakarta.validation.Valid;
import lombok.RequiredArgsConstructor;
import org.springframework.web.bind.annotation.*;

@RestController
@RequestMapping("/api/points")
@RequiredArgsConstructor
public class PointsController {

    private final PointsService pointsService;

    @PostMapping("/earn")
    public ApiResponse<PointsRecord> earnPoints(@Valid @RequestBody EarnPointsRequest request) {
        PointsRecord record = pointsService.earnPoints(request);
        return ApiResponse.success(record);
    }

    @PostMapping("/use")
    public ApiResponse<PointsRecord> usePoints(@Valid @RequestBody UsePointsRequest request) {
        PointsRecord record = pointsService.usePoints(request);
        return ApiResponse.success(record);
    }

    @PostMapping("/refund")
    public ApiResponse<PointsRecord> refundPoints(@Valid @RequestBody RefundRequest request) {
        PointsRecord record = pointsService.refundPoints(request);
        return ApiResponse.success(record);
    }
}

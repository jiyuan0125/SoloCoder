package com.retail.memberpoints.controller;

import com.retail.memberpoints.dto.ApiResponse;
import com.retail.memberpoints.entity.Member;
import com.retail.memberpoints.entity.PointsRecord;
import com.retail.memberpoints.service.MemberService;
import com.retail.memberpoints.service.PointsExpirationService;
import com.retail.memberpoints.service.PointsService;
import lombok.RequiredArgsConstructor;
import org.springframework.web.bind.annotation.*;

import java.util.List;

@RestController
@RequestMapping("/api/admin")
@RequiredArgsConstructor
public class AdminController {

    private final MemberService memberService;
    private final PointsService pointsService;
    private final PointsExpirationService pointsExpirationService;

    @GetMapping("/members")
    public ApiResponse<List<Member>> getAllMembers() {
        List<Member> members = memberService.getAllMembers();
        return ApiResponse.success(members);
    }

    @GetMapping("/points-records")
    public ApiResponse<List<PointsRecord>> getAllPointsRecords() {
        List<PointsRecord> records = pointsService.getAllPointsRecords();
        return ApiResponse.success(records);
    }

    @PostMapping("/expire-points")
    public ApiResponse<Integer> expirePoints() {
        int expired = pointsExpirationService.expirePoints();
        return ApiResponse.success(expired);
    }
}

package com.retail.memberpoints.controller;

import com.retail.memberpoints.dto.ApiResponse;
import com.retail.memberpoints.dto.MemberPointsSummary;
import com.retail.memberpoints.dto.RegisterMemberRequest;
import com.retail.memberpoints.entity.Member;
import com.retail.memberpoints.entity.PointsRecord;
import com.retail.memberpoints.service.MemberService;
import com.retail.memberpoints.service.PointsService;
import jakarta.validation.Valid;
import lombok.RequiredArgsConstructor;
import org.springframework.web.bind.annotation.*;

import java.util.List;

@RestController
@RequestMapping("/api/members")
@RequiredArgsConstructor
public class MemberController {

    private final MemberService memberService;
    private final PointsService pointsService;

    @PostMapping
    public ApiResponse<Member> register(@Valid @RequestBody RegisterMemberRequest request) {
        Member member = memberService.register(request);
        return ApiResponse.success(member);
    }

    @GetMapping("/{memberId}")
    public ApiResponse<Member> getById(@PathVariable Long memberId) {
        Member member = memberService.getById(memberId);
        return ApiResponse.success(member);
    }

    @GetMapping("/phone/{phone}")
    public ApiResponse<Member> getByPhone(@PathVariable String phone) {
        Member member = memberService.getByPhone(phone);
        return ApiResponse.success(member);
    }

    @GetMapping("/{memberId}/summary")
    public ApiResponse<MemberPointsSummary> getSummary(@PathVariable Long memberId) {
        MemberPointsSummary summary = pointsService.getMemberSummary(memberId);
        return ApiResponse.success(summary);
    }

    @GetMapping("/{memberId}/points/history")
    public ApiResponse<List<PointsRecord>> getPointsHistory(@PathVariable Long memberId) {
        List<PointsRecord> history = pointsService.getPointsHistory(memberId);
        return ApiResponse.success(history);
    }

    @GetMapping("/{memberId}/points/available")
    public ApiResponse<Integer> getAvailablePoints(@PathVariable Long memberId) {
        int available = pointsService.getAvailablePoints(memberId);
        return ApiResponse.success(available);
    }

    @GetMapping("/{memberId}/points/soon-expiring")
    public ApiResponse<Integer> getSoonExpiringPoints(@PathVariable Long memberId) {
        int expiring = pointsService.getSoonExpiringPoints(memberId);
        return ApiResponse.success(expiring);
    }
}

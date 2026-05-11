package com.example.groupbuy.controller;

import com.example.groupbuy.common.Result;
import com.example.groupbuy.dto.PayGroupBuyDTO;
import com.example.groupbuy.entity.GroupBuyParticipant;
import com.example.groupbuy.service.GroupBuyPayService;
import com.example.groupbuy.service.GroupBuyRefundService;
import lombok.RequiredArgsConstructor;
import org.springframework.web.bind.annotation.*;

import javax.validation.Valid;
import java.util.List;

@RestController
@RequestMapping("/api/group-buy/pay")
@RequiredArgsConstructor
public class GroupBuyPayController {
    
    private final GroupBuyPayService payService;
    private final GroupBuyRefundService refundService;
    
    @PostMapping
    public Result<GroupBuyParticipant> pay(@Valid @RequestBody PayGroupBuyDTO dto) {
        GroupBuyParticipant participant = payService.pay(dto);
        return Result.success(participant);
    }
    
    @GetMapping("/participants/{groupOrderId}")
    public Result<List<GroupBuyParticipant>> getPaidParticipants(@PathVariable Long groupOrderId) {
        List<GroupBuyParticipant> participants = payService.getPaidParticipants(groupOrderId);
        return Result.success(participants);
    }
    
    @PostMapping("/refund/{participantId}")
    public Result<Boolean> applyRefund(@PathVariable Long participantId) {
        boolean success = refundService.userApplyRefund(participantId);
        return Result.success(success);
    }
}

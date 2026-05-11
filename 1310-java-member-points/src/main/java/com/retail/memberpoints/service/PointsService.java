package com.retail.memberpoints.service;

import com.retail.memberpoints.dto.EarnPointsRequest;
import com.retail.memberpoints.dto.MemberPointsSummary;
import com.retail.memberpoints.dto.RefundRequest;
import com.retail.memberpoints.dto.UsePointsRequest;
import com.retail.memberpoints.entity.Member;
import com.retail.memberpoints.entity.PointsRecord;
import com.retail.memberpoints.enums.MemberLevel;
import com.retail.memberpoints.enums.PointsStatus;
import com.retail.memberpoints.enums.TransactionType;
import com.retail.memberpoints.exception.BusinessException;
import com.retail.memberpoints.repository.PointsRecordRepository;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.math.BigDecimal;
import java.math.RoundingMode;
import java.time.LocalDateTime;
import java.util.List;

@Service
@Slf4j
@RequiredArgsConstructor
public class PointsService {

    private final MemberService memberService;
    private final PointsRecordRepository pointsRecordRepository;

    @Transactional
    public PointsRecord earnPoints(EarnPointsRequest request) {
        Member member = memberService.getById(request.getMemberId());
        
        MemberLevel currentLevel = member.getLevel();
        int basePoints = request.getOrderAmount().setScale(0, RoundingMode.DOWN).intValue();
        int earnedPoints = (int) Math.floor(basePoints * currentLevel.getMultiplier());
        
        memberService.addSpending(request.getMemberId(), request.getOrderAmount());
        
        PointsRecord record = PointsRecord.builder()
                .memberId(member.getId())
                .orderNo(request.getOrderNo())
                .transactionType(TransactionType.EARN)
                .memberLevelAtTime(currentLevel)
                .points(earnedPoints)
                .orderAmount(request.getOrderAmount())
                .status(PointsStatus.AVAILABLE)
                .remark(String.format("消费%.2f元，%s倍率，获得%d积分",
                        request.getOrderAmount(), currentLevel.getDescription(), earnedPoints))
                .build();
        
        PointsRecord saved = pointsRecordRepository.save(record);
        
        memberService.updatePointsBalance(member.getId(), earnedPoints);
        
        log.info("会员 {} 获得积分: 订单 {}, 金额 {}, 等级 {}, 积分 {}", 
                member.getId(), request.getOrderNo(), request.getOrderAmount(), 
                currentLevel, earnedPoints);
        
        return saved;
    }

    @Transactional
    public PointsRecord refundPoints(RefundRequest request) {
        Member member = memberService.getById(request.getMemberId());
        
        List<PointsRecord> earnRecords = pointsRecordRepository
                .findByMemberIdAndTransactionTypeAndOrderNo(
                        request.getMemberId(), TransactionType.EARN, request.getOrderNo());
        
        if (earnRecords.isEmpty()) {
            throw new BusinessException("未找到该订单的积分记录");
        }
        
        List<PointsRecord> deductRecords = pointsRecordRepository
                .findByMemberIdAndTransactionTypeAndOrderNo(
                        request.getMemberId(), TransactionType.DEDUCT, request.getOrderNo());
        
        if (!deductRecords.isEmpty()) {
            throw BusinessException.orderAlreadyRefunded();
        }
        
        PointsRecord earnRecord = earnRecords.get(0);
        int pointsToDeduct = earnRecord.getPoints();
        
        PointsRecord deductRecord = PointsRecord.builder()
                .memberId(member.getId())
                .orderNo(request.getOrderNo())
                .transactionType(TransactionType.DEDUCT)
                .memberLevelAtTime(member.getLevel())
                .points(-pointsToDeduct)
                .orderAmount(earnRecord.getOrderAmount())
                .status(PointsStatus.AVAILABLE)
                .remark(String.format("订单退款，扣回%d积分", pointsToDeduct))
                .build();
        
        PointsRecord saved = pointsRecordRepository.save(deductRecord);
        
        memberService.updatePointsBalance(member.getId(), -pointsToDeduct);
        
        log.info("会员 {} 退款扣回积分: 订单 {}, 扣回积分 {}", 
                member.getId(), request.getOrderNo(), pointsToDeduct);
        
        return saved;
    }

    @Transactional
    public PointsRecord usePoints(UsePointsRequest request) {
        Member member = memberService.getById(request.getMemberId());
        
        int availableBalance = getAvailablePoints(member.getId());
        
        if (request.getPoints() > availableBalance) {
            throw BusinessException.insufficientPoints();
        }
        
        BigDecimal deductionAmount = new BigDecimal(request.getPoints())
                .divide(new BigDecimal("100"), 2, RoundingMode.DOWN);
        
        if (deductionAmount.compareTo(request.getOrderAmount()) > 0) {
            throw BusinessException.invalidDeductionAmount();
        }
        
        consumePointsFromOldest(member.getId(), request.getPoints());
        
        PointsRecord record = PointsRecord.builder()
                .memberId(member.getId())
                .orderNo(request.getOrderNo())
                .transactionType(TransactionType.USE)
                .memberLevelAtTime(member.getLevel())
                .points(-request.getPoints())
                .orderAmount(request.getOrderAmount())
                .status(PointsStatus.AVAILABLE)
                .remark(String.format("下单使用%d积分，抵扣%.2f元", 
                        request.getPoints(), deductionAmount))
                .build();
        
        PointsRecord saved = pointsRecordRepository.save(record);
        
        memberService.updatePointsBalance(member.getId(), -request.getPoints());
        
        log.info("会员 {} 使用积分: 订单 {}, 使用积分 {}, 抵扣金额 {}", 
                member.getId(), request.getOrderNo(), request.getPoints(), deductionAmount);
        
        return saved;
    }

    private void consumePointsFromOldest(Long memberId, int pointsToConsume) {
        List<PointsRecord> availableRecords = pointsRecordRepository
                .findEarnedAvailablePointsOrderByAcquireTime(memberId);
        
        int remaining = pointsToConsume;
        
        for (PointsRecord record : availableRecords) {
            if (remaining <= 0) break;
            
            int recordPoints = record.getPoints();
            if (recordPoints <= remaining) {
                record.setStatus(PointsStatus.USED);
                remaining -= recordPoints;
            } else {
                record.setPoints(recordPoints - remaining);
                remaining = 0;
            }
            pointsRecordRepository.save(record);
        }
    }

    public int getAvailablePoints(Long memberId) {
        List<PointsRecord> records = pointsRecordRepository
                .findEarnedAvailablePointsOrderByAcquireTime(memberId);
        
        LocalDateTime now = LocalDateTime.now();
        int total = 0;
        
        for (PointsRecord record : records) {
            if (record.getExpireTime() == null || record.getExpireTime().isAfter(now)) {
                total += record.getPoints();
            }
        }
        
        List<PointsRecord> allRecords = pointsRecordRepository
                .findByMemberIdOrderByCreateTimeDesc(memberId);
        
        for (PointsRecord record : allRecords) {
            if (record.getTransactionType() == TransactionType.DEDUCT &&
                record.getStatus() == PointsStatus.AVAILABLE) {
                total += record.getPoints();
            }
        }
        
        return total;
    }

    public int getSoonExpiringPoints(Long memberId) {
        LocalDateTime now = LocalDateTime.now();
        int currentYear = now.getYear();
        LocalDateTime endOfYear = LocalDateTime.of(currentYear, 12, 31, 23, 59, 59);
        
        List<PointsRecord> records = pointsRecordRepository
                .findAvailablePointsWithExpireBefore(memberId, PointsStatus.AVAILABLE, endOfYear);
        
        int total = 0;
        for (PointsRecord record : records) {
            if (record.getTransactionType() == TransactionType.EARN) {
                total += record.getPoints();
            }
        }
        
        return total;
    }

    public List<PointsRecord> getPointsHistory(Long memberId) {
        return pointsRecordRepository.findByMemberIdOrderByCreateTimeDesc(memberId);
    }

    public List<PointsRecord> getAllPointsRecords() {
        return pointsRecordRepository.findAll();
    }

    public MemberPointsSummary getMemberSummary(Long memberId) {
        Member member = memberService.getById(memberId);
        int availablePoints = getAvailablePoints(memberId);
        int soonExpiring = getSoonExpiringPoints(memberId);
        
        return MemberPointsSummary.builder()
                .memberId(member.getId())
                .phone(member.getPhone())
                .name(member.getName())
                .level(member.getLevel())
                .levelDescription(member.getLevel().getDescription())
                .totalSpending(member.getTotalSpending())
                .pointsBalance(availablePoints)
                .soonExpiringPoints(soonExpiring)
                .registerTime(member.getRegisterTime())
                .build();
    }
}

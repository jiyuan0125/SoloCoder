package com.retail.memberpoints.service;

import com.retail.memberpoints.entity.Member;
import com.retail.memberpoints.entity.PointsRecord;
import com.retail.memberpoints.enums.PointsStatus;
import com.retail.memberpoints.enums.TransactionType;
import com.retail.memberpoints.repository.MemberRepository;
import com.retail.memberpoints.repository.PointsRecordRepository;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.scheduling.annotation.Scheduled;
import org.springframework.stereotype.Service;
import org.springframework.transaction.annotation.Transactional;

import java.time.LocalDateTime;
import java.util.List;

@Service
@Slf4j
@RequiredArgsConstructor
public class PointsExpirationService {

    private final PointsRecordRepository pointsRecordRepository;
    private final MemberRepository memberRepository;

    @Scheduled(cron = "59 59 23 31 12 ?")
    @Transactional
    public void expirePointsYearly() {
        log.info("开始执行年度积分过期清理任务");
        expirePoints();
        log.info("年度积分过期清理任务执行完成");
    }

    @Transactional
    public int expirePoints() {
        LocalDateTime now = LocalDateTime.now();
        
        log.info("开始执行积分过期清理");
        
        List<PointsRecord> allRecords = pointsRecordRepository.findAll();
        
        int expiredCount = 0;
        int expiredPoints = 0;
        
        for (PointsRecord record : allRecords) {
            if (record.getTransactionType() == TransactionType.EARN &&
                record.getStatus() == PointsStatus.AVAILABLE &&
                record.getExpireTime() != null &&
                !record.getExpireTime().isAfter(now)) {
                
                Member member = memberRepository.findById(record.getMemberId()).orElse(null);
                if (member == null) {
                    continue;
                }
                
                int expiringYear = record.getExpireTime().getYear();
                
                record.setStatus(PointsStatus.EXPIRED);
                pointsRecordRepository.save(record);
                
                PointsRecord expireRecord = PointsRecord.builder()
                        .memberId(record.getMemberId())
                        .transactionType(TransactionType.EXPIRED)
                        .memberLevelAtTime(member.getLevel())
                        .points(-record.getPoints())
                        .status(PointsStatus.AVAILABLE)
                        .remark(String.format("%d年度积分过期，扣回%d积分", expiringYear, record.getPoints()))
                        .build();
                pointsRecordRepository.save(expireRecord);
                
                member.setPointsBalance(member.getPointsBalance() - record.getPoints());
                memberRepository.save(member);
                
                expiredCount++;
                expiredPoints += record.getPoints();
                
                log.info("会员 {} 积分过期: {} 积分", record.getMemberId(), record.getPoints());
            }
        }
        
        log.info("积分过期清理完成，处理记录数: {}, 过期积分总数: {}", expiredCount, expiredPoints);
        return expiredPoints;
    }

    public void checkAndUpdateExpiredStatus() {
        LocalDateTime now = LocalDateTime.now();
        List<PointsRecord> records = pointsRecordRepository.findAll();
        
        for (PointsRecord record : records) {
            if (record.getTransactionType() == TransactionType.EARN &&
                record.getStatus() == PointsStatus.AVAILABLE &&
                record.getExpireTime() != null &&
                record.getExpireTime().isBefore(now)) {
                
                log.debug("积分记录 {} 已过期但状态未更新", record.getId());
            }
        }
    }
}

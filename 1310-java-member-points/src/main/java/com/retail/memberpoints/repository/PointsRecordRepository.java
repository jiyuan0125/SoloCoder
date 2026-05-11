package com.retail.memberpoints.repository;

import com.retail.memberpoints.entity.PointsRecord;
import com.retail.memberpoints.enums.PointsStatus;
import com.retail.memberpoints.enums.TransactionType;
import org.springframework.data.jpa.repository.JpaRepository;
import org.springframework.data.jpa.repository.Query;
import org.springframework.data.repository.query.Param;
import org.springframework.stereotype.Repository;

import java.time.LocalDateTime;
import java.util.List;

@Repository
public interface PointsRecordRepository extends JpaRepository<PointsRecord, Long> {

    List<PointsRecord> findByMemberIdOrderByCreateTimeDesc(Long memberId);

    List<PointsRecord> findByMemberIdAndTransactionTypeInOrderByCreateTimeDesc(Long memberId, List<TransactionType> types);

    List<PointsRecord> findByMemberIdAndStatusOrderByAcquireTimeAsc(Long memberId, PointsStatus status);

    @Query("SELECT pr FROM PointsRecord pr WHERE pr.memberId = :memberId AND pr.transactionType = :type AND pr.orderNo = :orderNo")
    List<PointsRecord> findByMemberIdAndTransactionTypeAndOrderNo(Long memberId, TransactionType type, String orderNo);

    @Query("SELECT pr FROM PointsRecord pr WHERE pr.memberId = :memberId AND pr.status = :status AND pr.expireTime <= :expireTime ORDER BY pr.acquireTime ASC")
    List<PointsRecord> findAvailablePointsWithExpireBefore(Long memberId, PointsStatus status, LocalDateTime expireTime);

    @Query("SELECT pr FROM PointsRecord pr WHERE pr.memberId = :memberId AND pr.status = 'AVAILABLE' AND pr.expireTime <= :endTime AND pr.acquireTime >= :startTime ORDER BY pr.acquireTime ASC")
    List<PointsRecord> findSoonExpiringPoints(Long memberId, LocalDateTime startTime, LocalDateTime endTime);

    @Query("SELECT pr FROM PointsRecord pr WHERE pr.transactionType = 'EARN' AND pr.status = 'AVAILABLE' AND pr.memberId = :memberId ORDER BY pr.acquireTime ASC")
    List<PointsRecord> findEarnedAvailablePointsOrderByAcquireTime(Long memberId);

    List<PointsRecord> findByOrderNo(String orderNo);
}

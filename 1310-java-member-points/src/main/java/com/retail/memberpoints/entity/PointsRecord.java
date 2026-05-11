package com.retail.memberpoints.entity;

import com.retail.memberpoints.enums.MemberLevel;
import com.retail.memberpoints.enums.PointsStatus;
import com.retail.memberpoints.enums.TransactionType;
import jakarta.persistence.*;
import lombok.AllArgsConstructor;
import lombok.Builder;
import lombok.Data;
import lombok.NoArgsConstructor;

import java.math.BigDecimal;
import java.time.LocalDateTime;

@Entity
@Table(name = "points_record")
@Data
@Builder
@NoArgsConstructor
@AllArgsConstructor
public class PointsRecord {

    @Id
    @GeneratedValue(strategy = GenerationType.IDENTITY)
    private Long id;

    @Column(nullable = false)
    private Long memberId;

    @Column(length = 50)
    private String orderNo;

    @Enumerated(EnumType.STRING)
    @Column(nullable = false, length = 20)
    private TransactionType transactionType;

    @Enumerated(EnumType.STRING)
    @Column(nullable = false, length = 20)
    private MemberLevel memberLevelAtTime;

    @Column(nullable = false)
    private Integer points;

    @Column(precision = 15, scale = 2)
    private BigDecimal orderAmount;

    @Column(nullable = false)
    private LocalDateTime acquireTime;

    @Column
    private LocalDateTime expireTime;

    @Enumerated(EnumType.STRING)
    @Column(nullable = false, length = 20)
    private PointsStatus status;

    @Column(length = 200)
    private String remark;

    @Column(nullable = false)
    private LocalDateTime createTime;

    @PrePersist
    protected void onCreate() {
        if (createTime == null) {
            createTime = LocalDateTime.now();
        }
        if (acquireTime == null) {
            acquireTime = LocalDateTime.now();
        }
        if (status == null) {
            status = PointsStatus.AVAILABLE;
        }
        if (expireTime == null && transactionType == TransactionType.EARN) {
            expireTime = LocalDateTime.of(
                acquireTime.getYear(), 12, 31, 23, 59, 59
            );
        }
    }
}

package com.retail.memberpoints.entity;

import com.retail.memberpoints.enums.MemberLevel;
import jakarta.persistence.*;
import lombok.AllArgsConstructor;
import lombok.Builder;
import lombok.Data;
import lombok.NoArgsConstructor;

import java.math.BigDecimal;
import java.time.LocalDateTime;

@Entity
@Table(name = "member")
@Data
@Builder
@NoArgsConstructor
@AllArgsConstructor
public class Member {

    @Id
    @GeneratedValue(strategy = GenerationType.IDENTITY)
    private Long id;

    @Column(unique = true, nullable = false, length = 20)
    private String phone;

    @Column(nullable = false, length = 50)
    private String name;

    @Column(nullable = false)
    private LocalDateTime registerTime;

    @Column(nullable = false, precision = 15, scale = 2)
    private BigDecimal totalSpending;

    @Column(nullable = false)
    private Integer pointsBalance;

    @Enumerated(EnumType.STRING)
    @Column(nullable = false, length = 20)
    private MemberLevel level;

    @Column(nullable = false)
    private LocalDateTime updateTime;

    @PrePersist
    protected void onCreate() {
        if (registerTime == null) {
            registerTime = LocalDateTime.now();
        }
        updateTime = LocalDateTime.now();
        if (totalSpending == null) {
            totalSpending = BigDecimal.ZERO;
        }
        if (pointsBalance == null) {
            pointsBalance = 0;
        }
        if (level == null) {
            level = MemberLevel.NORMAL;
        }
    }

    @PreUpdate
    protected void onUpdate() {
        updateTime = LocalDateTime.now();
    }
}

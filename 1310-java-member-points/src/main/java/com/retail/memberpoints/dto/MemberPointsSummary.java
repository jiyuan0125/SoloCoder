package com.retail.memberpoints.dto;

import com.retail.memberpoints.enums.MemberLevel;
import lombok.AllArgsConstructor;
import lombok.Builder;
import lombok.Data;
import lombok.NoArgsConstructor;

import java.math.BigDecimal;
import java.time.LocalDateTime;

@Data
@Builder
@NoArgsConstructor
@AllArgsConstructor
public class MemberPointsSummary {

    private Long memberId;
    private String phone;
    private String name;
    private MemberLevel level;
    private String levelDescription;
    private BigDecimal totalSpending;
    private Integer pointsBalance;
    private Integer soonExpiringPoints;
    private LocalDateTime registerTime;
}

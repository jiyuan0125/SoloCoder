package com.performance.common.enums;

import java.math.BigDecimal;
import java.math.RoundingMode;

public enum PerformanceGrade {
    S("S", new BigDecimal("4.5"), null, new BigDecimal("1.5")),
    A("A", new BigDecimal("3.5"), new BigDecimal("4.5"), new BigDecimal("1.2")),
    B("B", new BigDecimal("2.5"), new BigDecimal("3.5"), new BigDecimal("1.0")),
    C("C", new BigDecimal("1.5"), new BigDecimal("2.5"), new BigDecimal("0.5")),
    D("D", null, new BigDecimal("1.5"), new BigDecimal("0"));

    private final String grade;
    private final BigDecimal minScore;
    private final BigDecimal maxScore;
    private final BigDecimal bonusCoefficient;

    PerformanceGrade(String grade, BigDecimal minScore, BigDecimal maxScore, BigDecimal bonusCoefficient) {
        this.grade = grade;
        this.minScore = minScore;
        this.maxScore = maxScore;
        this.bonusCoefficient = bonusCoefficient;
    }

    public static PerformanceGrade fromScore(BigDecimal score) {
        BigDecimal normalized = score.setScale(1, RoundingMode.HALF_UP);
        
        if (normalized.compareTo(new BigDecimal("4.5")) >= 0) {
            return S;
        } else if (normalized.compareTo(new BigDecimal("3.5")) >= 0) {
            return A;
        } else if (normalized.compareTo(new BigDecimal("2.5")) >= 0) {
            return B;
        } else if (normalized.compareTo(new BigDecimal("1.5")) >= 0) {
            return C;
        } else {
            return D;
        }
    }

    public String getGrade() {
        return grade;
    }

    public BigDecimal getMinScore() {
        return minScore;
    }

    public BigDecimal getMaxScore() {
        return maxScore;
    }

    public BigDecimal getBonusCoefficient() {
        return bonusCoefficient;
    }
}
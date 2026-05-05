package com.performance.common.util;

import java.math.BigDecimal;
import java.math.RoundingMode;

public class ScoreValidator {
    private static final BigDecimal MIN_SCORE = new BigDecimal("1.0");
    private static final BigDecimal MAX_SCORE = new BigDecimal("5.0");
    private static final BigDecimal MIN_DIFFERENCE = new BigDecimal("0.5");

    private ScoreValidator() {
    }

    public static boolean isValidScore(BigDecimal score) {
        if (score == null) {
            return false;
        }
        BigDecimal normalized = score.setScale(1, RoundingMode.HALF_UP);
        return normalized.compareTo(MIN_SCORE) >= 0 && normalized.compareTo(MAX_SCORE) <= 0;
    }

    public static boolean hasMinimumDifference(BigDecimal score1, BigDecimal score2) {
        if (score1 == null || score2 == null) {
            return false;
        }
        BigDecimal normalized1 = normalizeScore(score1);
        BigDecimal normalized2 = normalizeScore(score2);
        BigDecimal diff = normalized1.subtract(normalized2).abs();
        return diff.compareTo(MIN_DIFFERENCE) >= 0;
    }

    public static BigDecimal normalizeScore(BigDecimal score) {
        if (score == null) {
            return null;
        }
        return score.setScale(1, RoundingMode.HALF_UP);
    }

    public static boolean isDifferenceZero(BigDecimal score1, BigDecimal score2) {
        if (score1 == null || score2 == null) {
            return false;
        }
        BigDecimal normalized1 = normalizeScore(score1);
        BigDecimal normalized2 = normalizeScore(score2);
        return normalized1.compareTo(normalized2) == 0;
    }
}
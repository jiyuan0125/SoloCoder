package com.payroll.server.util;

import com.payroll.server.entity.DeductionDetail;
import com.payroll.server.entity.PayrollArchive;

import java.math.BigDecimal;
import java.nio.charset.StandardCharsets;
import java.security.MessageDigest;
import java.security.NoSuchAlgorithmException;
import java.util.List;
import java.util.stream.Collectors;

public final class ChecksumUtil {

    private ChecksumUtil() {
    }

    public static String calculateChecksum(PayrollArchive archive) {
        StringBuilder sb = new StringBuilder();
        
        sb.append(archive.getEmployeeId()).append("|");
        sb.append(archive.getYear()).append("|");
        sb.append(archive.getMonth()).append("|");
        sb.append(formatBigDecimal(archive.getBaseSalary())).append("|");
        sb.append(formatDeductionDetails(archive.getDeductionDetails())).append("|");
        sb.append(formatBigDecimal(archive.getTotalDeduction())).append("|");
        sb.append(formatBigDecimal(archive.getNetPay()));

        return md5Hash(sb.toString());
    }

    public static boolean verifyChecksum(PayrollArchive archive, String storedChecksum) {
        if (storedChecksum == null || storedChecksum.isEmpty()) {
            return false;
        }
        String calculatedChecksum = calculateChecksum(archive);
        return storedChecksum.equals(calculatedChecksum);
    }

    private static String formatBigDecimal(BigDecimal value) {
        if (value == null) {
            return "0";
        }
        return value.stripTrailingZeros().toPlainString();
    }

    private static String formatDeductionDetails(List<DeductionDetail> details) {
        if (details == null || details.isEmpty()) {
            return "";
        }
        return details.stream()
                .sorted((d1, d2) -> {
                    if (d1.getDeductionType() == null) return -1;
                    if (d2.getDeductionType() == null) return 1;
                    return d1.getDeductionType().compareTo(d2.getDeductionType());
                })
                .map(d -> (d.getDeductionType() == null ? "" : d.getDeductionType()) 
                        + ":" + formatBigDecimal(d.getAmount()))
                .collect(Collectors.joining(","));
    }

    private static String md5Hash(String input) {
        try {
            MessageDigest md = MessageDigest.getInstance("MD5");
            byte[] hashBytes = md.digest(input.getBytes(StandardCharsets.UTF_8));
            StringBuilder sb = new StringBuilder();
            for (byte b : hashBytes) {
                sb.append(String.format("%02x", b));
            }
            return sb.toString();
        } catch (NoSuchAlgorithmException e) {
            throw new RuntimeException("MD5 algorithm not available", e);
        }
    }
}

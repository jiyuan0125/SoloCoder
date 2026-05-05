package com.payroll.server.util;

public final class MaskUtil {

    private MaskUtil() {
    }

    public static String maskName(String name) {
        if (name == null || name.isEmpty()) {
            return name;
        }
        if (name.length() <= 1) {
            return "*";
        }
        if (name.length() == 2) {
            return name.charAt(0) + "*";
        }
        return name.charAt(0) + "*".repeat(name.length() - 2) + name.charAt(name.length() - 1);
    }

    public static String maskIdCard(String idCard) {
        if (idCard == null || idCard.length() < 8) {
            return idCard;
        }
        int maskLength = idCard.length() - 8;
        if (maskLength <= 0) {
            return idCard;
        }
        return idCard.substring(0, 4) + "*".repeat(maskLength) + idCard.substring(idCard.length() - 4);
    }
}

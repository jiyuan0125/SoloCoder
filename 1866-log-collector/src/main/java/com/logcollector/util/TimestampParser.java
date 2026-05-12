package com.logcollector.util;

import java.time.Instant;
import java.time.LocalDateTime;
import java.time.ZoneId;
import java.time.ZoneOffset;
import java.time.ZonedDateTime;
import java.time.format.DateTimeFormatter;
import java.time.format.DateTimeParseException;
import java.util.List;

public class TimestampParser {

    private static final List<DateTimeFormatter> FORMATTERS = List.of(
            DateTimeFormatter.ISO_ZONED_DATE_TIME,
            DateTimeFormatter.ISO_OFFSET_DATE_TIME,
            DateTimeFormatter.ISO_LOCAL_DATE_TIME,
            DateTimeFormatter.ofPattern("yyyy-MM-dd HH:mm:ss"),
            DateTimeFormatter.ofPattern("yyyy-MM-dd HH:mm:ss.SSS"),
            DateTimeFormatter.ofPattern("yyyy/MM/dd HH:mm:ss"),
            DateTimeFormatter.ofPattern("yyyy/MM/dd HH:mm:ss.SSS"),
            DateTimeFormatter.ofPattern("yyyyMMddHHmmss")
    );

    private static final long MILLIS_LIMIT = 1_000_000_000_000L;

    public static Long parse(Object timestamp) {
        if (timestamp == null) {
            return null;
        }

        if (timestamp instanceof Long) {
            return normalizeEpoch((Long) timestamp);
        }
        if (timestamp instanceof Integer) {
            return normalizeEpoch((Integer) timestamp * 1000L);
        }
        if (timestamp instanceof Number) {
            return normalizeEpoch(((Number) timestamp).longValue());
        }
        if (timestamp instanceof String) {
            return parseString((String) timestamp);
        }

        return null;
    }

    private static Long normalizeEpoch(long value) {
        if (value <= MILLIS_LIMIT) {
            return value * 1000L;
        }
        return value;
    }

    private static Long parseString(String value) {
        String trimmed = value.trim();
        if (trimmed.isEmpty()) {
            return null;
        }

        try {
            long number = Long.parseLong(trimmed);
            return normalizeEpoch(number);
        } catch (NumberFormatException ignored) {
        }

        for (DateTimeFormatter formatter : FORMATTERS) {
            try {
                return parseWithFormatter(trimmed, formatter);
            } catch (DateTimeParseException ignored) {
            }
        }

        return null;
    }

    private static Long parseWithFormatter(String value, DateTimeFormatter formatter) {
        try {
            ZonedDateTime zoned = ZonedDateTime.parse(value, formatter);
            return zoned.toInstant().toEpochMilli();
        } catch (DateTimeParseException e) {
            try {
                Instant instant = Instant.parse(value);
                return instant.toEpochMilli();
            } catch (DateTimeParseException e2) {
                try {
                    LocalDateTime local = LocalDateTime.parse(value, formatter);
                    return local.atZone(ZoneId.systemDefault()).withZoneSameInstant(ZoneOffset.UTC).toInstant().toEpochMilli();
                } catch (DateTimeParseException e3) {
                    throw e;
                }
            }
        }
    }
}

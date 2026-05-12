package com.logaggregate.service.ingest;

import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import java.time.Instant;
import java.time.LocalDateTime;
import java.time.ZoneId;
import java.time.ZonedDateTime;
import java.time.format.DateTimeFormatter;
import java.time.format.DateTimeParseException;
import java.util.List;

public class TimestampParser {

    private static final Logger logger = LoggerFactory.getLogger(TimestampParser.class);

    private static final List<DateTimeFormatter> FORMATTERS = List.of(
            DateTimeFormatter.ISO_INSTANT,
            DateTimeFormatter.ISO_ZONED_DATE_TIME,
            DateTimeFormatter.ISO_OFFSET_DATE_TIME,
            DateTimeFormatter.ISO_LOCAL_DATE_TIME,
            DateTimeFormatter.ofPattern("yyyy-MM-dd HH:mm:ss"),
            DateTimeFormatter.ofPattern("yyyy-MM-dd'T'HH:mm:ss"),
            DateTimeFormatter.ofPattern("yyyy/MM/dd HH:mm:ss")
    );

    public static long parseToUtcMillis(Object timestamp) {
        if (timestamp == null) {
            throw new IllegalArgumentException("timestamp is required");
        }

        if (timestamp instanceof Number) {
            long value = ((Number) timestamp).longValue();
            if (value < 10000000000L) {
                return value * 1000;
            }
            return value;
        }

        if (timestamp instanceof String) {
            String tsStr = ((String) timestamp).trim();
            try {
                long num = Long.parseLong(tsStr);
                if (num < 10000000000L) {
                    return num * 1000;
                }
                return num;
            } catch (NumberFormatException ignored) {
            }

            for (DateTimeFormatter formatter : FORMATTERS) {
                try {
                    Instant instant;
                    if (formatter.equals(DateTimeFormatter.ISO_INSTANT)) {
                        instant = Instant.from(formatter.parse(tsStr));
                    } else {
                        try {
                            ZonedDateTime zdt = ZonedDateTime.parse(tsStr, formatter);
                            instant = zdt.toInstant();
                        } catch (DateTimeParseException e) {
                            LocalDateTime ldt = LocalDateTime.parse(tsStr, formatter);
                            instant = ldt.atZone(ZoneId.systemDefault()).withZoneSameInstant(ZoneId.of("UTC")).toInstant();
                        }
                    }
                    return instant.toEpochMilli();
                } catch (DateTimeParseException ignored) {
                }
            }

            throw new IllegalArgumentException("Unable to parse timestamp: " + tsStr);
        }

        throw new IllegalArgumentException("Unsupported timestamp type: " + timestamp.getClass());
    }
}

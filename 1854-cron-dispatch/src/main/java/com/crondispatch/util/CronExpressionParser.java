package com.crondispatch.util;

import java.time.Instant;
import java.time.LocalDateTime;
import java.time.ZoneId;
import java.util.BitSet;

public class CronExpressionParser {

    private static final int MINUTE = 0;
    private static final int HOUR = 1;
    private static final int DAY_OF_MONTH = 2;
    private static final int MONTH = 3;
    private static final int DAY_OF_WEEK = 4;

    private final BitSet[] fields = new BitSet[5];
    private final String expression;

    public CronExpressionParser(String expression) {
        this.expression = expression;
        parse();
    }

    private void parse() {
        String[] parts = expression.trim().split("\\s+");
        if (parts.length != 5) {
            throw new IllegalArgumentException("Cron表达式必须包含5个字段: " + expression);
        }

        fields[MINUTE] = parseField(parts[0], 0, 59);
        fields[HOUR] = parseField(parts[1], 0, 23);
        fields[DAY_OF_MONTH] = parseField(parts[2], 1, 31);
        fields[MONTH] = parseField(parts[3], 1, 12);
        fields[DAY_OF_WEEK] = parseField(parts[4], 0, 6);
    }

    private BitSet parseField(String field, int min, int max) {
        BitSet bitSet = new BitSet(max - min + 1);

        if (field.equals("*")) {
            bitSet.set(0, max - min + 1);
            return bitSet;
        }

        String[] parts = field.split(",");
        for (String part : parts) {
            if (part.contains("/")) {
                String[] stepParts = part.split("/");
                int step = Integer.parseInt(stepParts[1]);
                String range = stepParts[0].equals("*") ? min + "-" + max : stepParts[0];
                String[] rangeParts = range.split("-");
                int start = Integer.parseInt(rangeParts[0]);
                int end = rangeParts.length > 1 ? Integer.parseInt(rangeParts[1]) : max;
                for (int i = start; i <= end; i += step) {
                    if (i >= min && i <= max) {
                        bitSet.set(i - min);
                    }
                }
            } else if (part.contains("-")) {
                String[] rangeParts = part.split("-");
                int start = Integer.parseInt(rangeParts[0]);
                int end = Integer.parseInt(rangeParts[1]);
                for (int i = start; i <= end; i++) {
                    if (i >= min && i <= max) {
                        bitSet.set(i - min);
                    }
                }
            } else {
                int value = Integer.parseInt(part);
                if (value >= min && value <= max) {
                    bitSet.set(value - min);
                }
            }
        }

        return bitSet;
    }

    public Instant getNextExecutionTime(Instant from) {
        ZoneId zoneId = ZoneId.systemDefault();
        LocalDateTime current = LocalDateTime.ofInstant(from, zoneId).withSecond(0).withNano(0).plusMinutes(1);
        
        for (int i = 0; i < 525600; i++) {
            LocalDateTime candidate = current.plusMinutes(i);
            
            int minute = candidate.getMinute();
            int hour = candidate.getHour();
            int dayOfMonth = candidate.getDayOfMonth();
            int month = candidate.getMonthValue();
            int dayOfWeek = candidate.getDayOfWeek().getValue() % 7;

            if (fields[MINUTE].get(minute) &&
                fields[HOUR].get(hour) &&
                fields[DAY_OF_MONTH].get(dayOfMonth - 1) &&
                fields[MONTH].get(month - 1) &&
                fields[DAY_OF_WEEK].get(dayOfWeek)) {
                return candidate.atZone(zoneId).toInstant();
            }
        }

        return null;
    }

    public String getExpression() {
        return expression;
    }

    public static boolean isValid(String expression) {
        try {
            new CronExpressionParser(expression);
            return true;
        } catch (Exception e) {
            return false;
        }
    }
}

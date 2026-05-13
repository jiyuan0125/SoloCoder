package com.messagequeue.filter;

import com.jayway.jsonpath.JsonPath;
import com.messagequeue.model.Message;
import lombok.AllArgsConstructor;

@AllArgsConstructor
public class JsonPathExpression implements FilterExpression {
    private final String jsonPath;
    private final String operator;
    private final String expectedValue;

    @Override
    public boolean evaluate(Message message) {
        try {
            String body = message.getBody();
            if (body == null) {
                return false;
            }
            Object value = JsonPath.read(body, jsonPath);
            return compare(value, operator, expectedValue);
        } catch (Exception e) {
            return false;
        }
    }

    private boolean compare(Object actual, String op, String expected) {
        if (actual == null) {
            return false;
        }
        String actualStr = actual.toString();
        switch (op) {
            case "=":
            case "==":
                return actualStr.equals(expected);
            case "!=":
                return !actualStr.equals(expected);
            case ">":
                return compareNumbers(actualStr, expected) > 0;
            case ">=":
                return compareNumbers(actualStr, expected) >= 0;
            case "<":
                return compareNumbers(actualStr, expected) < 0;
            case "<=":
                return compareNumbers(actualStr, expected) <= 0;
            case "CONTAINS":
                return actualStr.contains(expected);
            default:
                return actualStr.equals(expected);
        }
    }

    private int compareNumbers(String actual, String expected) {
        try {
            double actualNum = Double.parseDouble(actual);
            double expectedNum = Double.parseDouble(expected);
            return Double.compare(actualNum, expectedNum);
        } catch (NumberFormatException e) {
            return actual.compareTo(expected);
        }
    }
}

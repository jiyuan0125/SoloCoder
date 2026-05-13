package com.messagequeue.filter;

import lombok.extern.slf4j.Slf4j;

import java.util.ArrayList;
import java.util.List;
import java.util.regex.Matcher;
import java.util.regex.Pattern;

@Slf4j
public class FilterParser {

    private static final Pattern TOKEN_PATTERN = Pattern.compile(
            "\\s*(?:(AND|OR)|([()])|([$][a-zA-Z0-9_\\[\\].]*\\s*(?:=|!=|>=|<=|>|<|CONTAINS)\\s*(?:'[^']*'|\"[^\"]*\"|[^\\s()]+))|([a-zA-Z_][a-zA-Z0-9_]*\\s*=\\s*(?:'[^']*'|\"[^\"]*\"|[^\\s()]+))|([a-zA-Z_][a-zA-Z0-9_]*\\s+(?:CONTAINS|contains)\\s+(?:'[^']*'|\"[^\"]*\"|[^\\s()]+)))\\s*"
    );

    private static final Pattern HEADER_EQ_PATTERN = Pattern.compile(
            "^([a-zA-Z_][a-zA-Z0-9_]*)\\s*=\\s*(?:'([^']*)'|\"([^\"]*)'?|([^\\s]+))$"
    );

    private static final Pattern HEADER_CONTAINS_PATTERN = Pattern.compile(
            "^([a-zA-Z_][a-zA-Z0-9_]*)\\s+(?:CONTAINS|contains)\\s+(?:'([^']*)'|\"([^\"]*)'?|([^\\s]+))$"
    );

    private static final Pattern JSONPATH_PATTERN = Pattern.compile(
            "^([$][a-zA-Z0-9_\\[\\].]*)\\s*(=|!=|>=|<=|>|<|CONTAINS)\\s*(?:'([^']*)'|\"([^\"]*)'?|([^\\s]+))$"
    );

    public static FilterExpression parse(String expression) {
        if (expression == null || expression.trim().isEmpty()) {
            return message -> true;
        }

        List<String> tokens = tokenize(expression);
        return parseExpression(tokens, 0, tokens.size());
    }

    private static List<String> tokenize(String expression) {
        List<String> tokens = new ArrayList<>();
        Matcher matcher = TOKEN_PATTERN.matcher(expression);
        int lastEnd = 0;

        while (matcher.find()) {
            if (matcher.start() > lastEnd) {
                String skipped = expression.substring(lastEnd, matcher.start()).trim();
                if (!skipped.isEmpty()) {
                    throw new IllegalArgumentException("Invalid token: " + skipped);
                }
            }

            for (int i = 1; i <= matcher.groupCount(); i++) {
                if (matcher.group(i) != null) {
                    tokens.add(matcher.group(i).trim());
                    break;
                }
            }
            lastEnd = matcher.end();
        }

        if (lastEnd < expression.length()) {
            String remaining = expression.substring(lastEnd).trim();
            if (!remaining.isEmpty()) {
                throw new IllegalArgumentException("Invalid expression at: " + remaining);
            }
        }

        return tokens;
    }

    private static FilterExpression parseExpression(List<String> tokens, int start, int end) {
        if (start >= end) {
            throw new IllegalArgumentException("Empty expression");
        }

        int orIndex = findTopLevelOr(tokens, start, end);
        if (orIndex != -1) {
            FilterExpression left = parseExpression(tokens, start, orIndex);
            FilterExpression right = parseExpression(tokens, orIndex + 1, end);
            return new OrExpression(left, right);
        }

        int andIndex = findTopLevelAnd(tokens, start, end);
        if (andIndex != -1) {
            FilterExpression left = parseExpression(tokens, start, andIndex);
            FilterExpression right = parseExpression(tokens, andIndex + 1, end);
            return new AndExpression(left, right);
        }

        if (tokens.get(start).equals("(")) {
            if (!tokens.get(end - 1).equals(")")) {
                throw new IllegalArgumentException("Unbalanced parentheses");
            }
            return parseExpression(tokens, start + 1, end - 1);
        }

        return parseAtom(tokens.get(start));
    }

    private static int findTopLevelOr(List<String> tokens, int start, int end) {
        int depth = 0;
        for (int i = start; i < end; i++) {
            String token = tokens.get(i);
            if (token.equals("(")) depth++;
            else if (token.equals(")")) depth--;
            else if (depth == 0 && token.equals("OR")) return i;
        }
        return -1;
    }

    private static int findTopLevelAnd(List<String> tokens, int start, int end) {
        int depth = 0;
        for (int i = start; i < end; i++) {
            String token = tokens.get(i);
            if (token.equals("(")) depth++;
            else if (token.equals(")")) depth--;
            else if (depth == 0 && token.equals("AND")) return i;
        }
        return -1;
    }

    private static FilterExpression parseAtom(String token) {
        Matcher headerEq = HEADER_EQ_PATTERN.matcher(token);
        if (headerEq.matches()) {
            String name = headerEq.group(1);
            String value = extractValue(headerEq, 2, 3, 4);
            return new HeaderEqualsExpression(name, value);
        }

        Matcher headerContains = HEADER_CONTAINS_PATTERN.matcher(token);
        if (headerContains.matches()) {
            String name = headerContains.group(1);
            String value = extractValue(headerContains, 2, 3, 4);
            return new HeaderContainsExpression(name, value);
        }

        Matcher jsonPath = JSONPATH_PATTERN.matcher(token);
        if (jsonPath.matches()) {
            String path = jsonPath.group(1);
            String op = jsonPath.group(2);
            String value = extractValue(jsonPath, 3, 4, 5);
            return new JsonPathExpression(path, op, value);
        }

        throw new IllegalArgumentException("Invalid atom expression: " + token);
    }

    private static String extractValue(Matcher matcher, int... groupIndices) {
        for (int idx : groupIndices) {
            String val = matcher.group(idx);
            if (val != null) {
                return val;
            }
        }
        return "";
    }
}

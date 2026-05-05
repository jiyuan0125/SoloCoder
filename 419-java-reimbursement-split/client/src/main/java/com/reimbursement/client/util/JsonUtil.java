package com.reimbursement.client.util;

import java.lang.reflect.Array;
import java.lang.reflect.Field;
import java.math.BigDecimal;
import java.time.LocalDateTime;
import java.time.format.DateTimeFormatter;
import java.util.*;

public class JsonUtil {
    private static final DateTimeFormatter DATE_FORMATTER = DateTimeFormatter.ISO_LOCAL_DATE_TIME;

    public static String toJson(Object obj) {
        if (obj == null) {
            return "null";
        }
        return serializeObject(obj);
    }

    @SuppressWarnings("unchecked")
    public static <T> T fromJson(String json, Class<T> clazz) {
        if (json == null || json.isEmpty()) {
            return null;
        }
        Map<String, Object> map = parseJson(json);
        return (T) parseToObject(map, clazz);
    }

    private static String serializeObject(Object obj) {
        if (obj == null) {
            return "null";
        }

        Class<?> clazz = obj.getClass();

        if (clazz == String.class) {
            return "\"" + escapeString((String) obj) + "\"";
        }
        if (clazz == Integer.class || clazz == int.class) {
            return obj.toString();
        }
        if (clazz == Long.class || clazz == long.class) {
            return obj.toString();
        }
        if (clazz == Boolean.class || clazz == boolean.class) {
            return obj.toString();
        }
        if (clazz == BigDecimal.class) {
            return "\"" + obj.toString() + "\"";
        }
        if (clazz == LocalDateTime.class) {
            return "\"" + ((LocalDateTime) obj).format(DATE_FORMATTER) + "\"";
        }
        if (Enum.class.isAssignableFrom(clazz)) {
            return "\"" + ((Enum<?>) obj).name() + "\"";
        }

        if (List.class.isAssignableFrom(clazz)) {
            return serializeList((List<?>) obj);
        }

        if (clazz.isArray()) {
            return serializeArray(obj);
        }

        return serializePojo(obj);
    }

    private static String serializeList(List<?> list) {
        StringBuilder sb = new StringBuilder("[");
        for (int i = 0; i < list.size(); i++) {
            if (i > 0) {
                sb.append(",");
            }
            sb.append(serializeObject(list.get(i)));
        }
        sb.append("]");
        return sb.toString();
    }

    private static String serializeArray(Object array) {
        int length = Array.getLength(array);
        StringBuilder sb = new StringBuilder("[");
        for (int i = 0; i < length; i++) {
            if (i > 0) {
                sb.append(",");
            }
            sb.append(serializeObject(Array.get(array, i)));
        }
        sb.append("]");
        return sb.toString();
    }

    private static String serializePojo(Object obj) {
        StringBuilder sb = new StringBuilder("{");
        boolean first = true;

        for (Field field : obj.getClass().getDeclaredFields()) {
            field.setAccessible(true);
            try {
                Object value = field.get(obj);
                if (value != null) {
                    if (!first) {
                        sb.append(",");
                    }
                    first = false;
                    sb.append("\"").append(field.getName()).append("\":");
                    sb.append(serializeObject(value));
                }
            } catch (IllegalAccessException e) {
            }
        }

        sb.append("}");
        return sb.toString();
    }

    private static String escapeString(String s) {
        return s.replace("\\", "\\\\")
                .replace("\"", "\\\"")
                .replace("\n", "\\n")
                .replace("\r", "\\r")
                .replace("\t", "\\t");
    }

    @SuppressWarnings("unchecked")
    private static Map<String, Object> parseJson(String json) {
        Map<String, Object> result = new LinkedHashMap<>();
        json = json.trim();
        
        if (json.startsWith("{")) {
            json = json.substring(1, json.length() - 1).trim();
        }

        int index = 0;
        while (index < json.length()) {
            while (index < json.length() && Character.isWhitespace(json.charAt(index))) {
                index++;
            }
            if (index >= json.length()) {
                break;
            }

            if (json.charAt(index) == '"') {
                int keyEnd = findNextQuote(json, index + 1);
                String key = json.substring(index + 1, keyEnd);
                index = keyEnd + 1;

                while (index < json.length() && Character.isWhitespace(json.charAt(index))) {
                    index++;
                }
                if (index < json.length() && json.charAt(index) == ':') {
                    index++;
                }

                while (index < json.length() && Character.isWhitespace(json.charAt(index))) {
                    index++;
                }

                Object value = parseValue(json, index);
                result.put(key, value);

                index = findNextSeparator(json, index);
            } else {
                break;
            }
        }

        return result;
    }

    private static Object parseValue(String json, int startIndex) {
        int index = startIndex;
        while (index < json.length() && Character.isWhitespace(json.charAt(index))) {
            index++;
        }

        if (index >= json.length()) {
            return null;
        }

        char c = json.charAt(index);
        
        if (c == '"') {
            int end = findNextQuote(json, index + 1);
            return json.substring(index + 1, end);
        }
        
        if (c == '{') {
            int end = findMatchingBrace(json, index);
            return parseJson(json.substring(index, end + 1));
        }
        
        if (c == '[') {
            int end = findMatchingBracket(json, index);
            return parseArray(json.substring(index, end + 1));
        }
        
        if (c == 't' && json.startsWith("true", index)) {
            return true;
        }
        
        if (c == 'f' && json.startsWith("false", index)) {
            return false;
        }
        
        if (c == 'n' && json.startsWith("null", index)) {
            return null;
        }

        int end = findEndOfValue(json, index);
        String valueStr = json.substring(index, end);
        
        try {
            if (valueStr.contains(".")) {
                return new BigDecimal(valueStr);
            }
            return Integer.parseInt(valueStr);
        } catch (NumberFormatException e) {
            return valueStr;
        }
    }

    private static List<Object> parseArray(String json) {
        List<Object> result = new ArrayList<>();
        json = json.substring(1, json.length() - 1).trim();
        
        int index = 0;
        while (index < json.length()) {
            while (index < json.length() && Character.isWhitespace(json.charAt(index))) {
                index++;
            }
            if (index >= json.length()) {
                break;
            }

            Object value = parseValue(json, index);
            result.add(value);

            index = findNextSeparator(json, index);
        }

        return result;
    }

    private static int findNextQuote(String json, int start) {
        for (int i = start; i < json.length(); i++) {
            if (json.charAt(i) == '"' && (i == 0 || json.charAt(i - 1) != '\\')) {
                return i;
            }
        }
        return json.length();
    }

    private static int findMatchingBrace(String json, int start) {
        int count = 0;
        for (int i = start; i < json.length(); i++) {
            if (json.charAt(i) == '{') {
                count++;
            } else if (json.charAt(i) == '}') {
                count--;
                if (count == 0) {
                    return i;
                }
            }
        }
        return json.length() - 1;
    }

    private static int findMatchingBracket(String json, int start) {
        int count = 0;
        for (int i = start; i < json.length(); i++) {
            if (json.charAt(i) == '[') {
                count++;
            } else if (json.charAt(i) == ']') {
                count--;
                if (count == 0) {
                    return i;
                }
            }
        }
        return json.length() - 1;
    }

    private static int findEndOfValue(String json, int start) {
        for (int i = start; i < json.length(); i++) {
            char c = json.charAt(i);
            if (c == ',' || c == '}' || c == ']' || Character.isWhitespace(c)) {
                return i;
            }
        }
        return json.length();
    }

    private static int findNextSeparator(String json, int start) {
        for (int i = start; i < json.length(); i++) {
            char c = json.charAt(i);
            if (c == ',') {
                return i + 1;
            }
            if (c == '}' || c == ']') {
                return json.length();
            }
        }
        return json.length();
    }

    @SuppressWarnings("unchecked")
    public static <T> T parseToObject(Map<String, Object> map, Class<T> clazz) {
        if (map == null) {
            return null;
        }

        try {
            Object instance = clazz.getDeclaredConstructor().newInstance();

            for (Field field : clazz.getDeclaredFields()) {
                field.setAccessible(true);
                Object value = map.get(field.getName());

                if (value != null) {
                    if (value instanceof Map) {
                        Object nested = parseToObject((Map<String, Object>) value, field.getType());
                        field.set(instance, nested);
                    } else if (value instanceof List) {
                        field.set(instance, value);
                    } else if (field.getType() == BigDecimal.class && value instanceof String) {
                        field.set(instance, new BigDecimal((String) value));
                    } else if (field.getType() == LocalDateTime.class && value instanceof String) {
                        field.set(instance, LocalDateTime.parse((String) value, DATE_FORMATTER));
                    } else if (field.getType().isEnum() && value instanceof String) {
                        @SuppressWarnings("rawtypes")
                        Class<? extends Enum> enumType = (Class<? extends Enum>) field.getType();
                        field.set(instance, Enum.valueOf(enumType, (String) value));
                    } else {
                        field.set(instance, value);
                    }
                }
            }

            return (T) instance;
        } catch (Exception e) {
            e.printStackTrace();
            return null;
        }
    }
}

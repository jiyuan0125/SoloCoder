package com.example.gateway.compatibility;

import com.example.gateway.util.JsonUtils;
import com.fasterxml.jackson.databind.JsonNode;
import com.fasterxml.jackson.databind.node.ArrayNode;
import com.fasterxml.jackson.databind.node.ObjectNode;

import java.util.List;
import java.util.Map;
import java.util.concurrent.ConcurrentHashMap;

public class CompatibilityService {
    private final Map<String, CompatibilityRule> rules = new ConcurrentHashMap<>();

    public void addRule(CompatibilityRule rule) {
        if (rule.getSourceVersion() != null) {
            rules.put(rule.getSourceVersion(), rule);
        }
    }

    public void removeRule(String version) {
        rules.remove(version);
    }

    public boolean hasRule(String version) {
        return rules.containsKey(version);
    }

    public String transformRequest(String version, String body) {
        CompatibilityRule rule = rules.get(version);
        if (rule == null || rule.getRequestMappings() == null || body == null || body.isEmpty()) {
            return body;
        }
        return transformJson(body, rule.getRequestMappings());
    }

    public String transformResponse(String version, String body) {
        CompatibilityRule rule = rules.get(version);
        if (rule == null || rule.getResponseMappings() == null || body == null || body.isEmpty()) {
            return body;
        }
        return transformJson(body, rule.getResponseMappings());
    }

    private String transformJson(String json, List<FieldMapping> mappings) {
        try {
            JsonNode root = JsonUtils.getMapper().readTree(json);
            if (root.isArray()) {
                ArrayNode result = JsonUtils.getMapper().createArrayNode();
                for (JsonNode item : root) {
                    result.add(transformObject((ObjectNode) item, mappings));
                }
                return JsonUtils.getMapper().writeValueAsString(result);
            } else if (root.isObject()) {
                ObjectNode result = transformObject((ObjectNode) root, mappings);
                return JsonUtils.getMapper().writeValueAsString(result);
            }
            return json;
        } catch (Exception e) {
            return json;
        }
    }

    private ObjectNode transformObject(ObjectNode source, List<FieldMapping> mappings) {
        ObjectNode result = JsonUtils.getMapper().createObjectNode();
        for (FieldMapping mapping : mappings) {
            JsonNode value = source.get(mapping.getFrom());
            if (value == null || value.isNull()) {
                if (mapping.getDefaultValue() != null) {
                    setField(result, mapping.getTo(), mapping.getDefaultValue(), mapping.getType());
                }
            } else {
                Object converted = convertValue(value, mapping.getType());
                setField(result, mapping.getTo(), converted, mapping.getType());
            }
        }
        return result;
    }

    private Object convertValue(JsonNode node, String type) {
        if (type == null) {
            if (node.isTextual()) return node.asText();
            if (node.isInt()) return node.asInt();
            if (node.isLong()) return node.asLong();
            if (node.isBoolean()) return node.asBoolean();
            if (node.isDouble()) return node.asDouble();
            return node;
        }
        switch (type.toLowerCase()) {
            case "string":
                return node.asText();
            case "int":
            case "integer":
                return node.asInt();
            case "long":
                return node.asLong();
            case "double":
            case "float":
                return node.asDouble();
            case "boolean":
                return node.asBoolean();
            default:
                return node;
        }
    }

    private void setField(ObjectNode node, String field, Object value, String type) {
        if (value == null) {
            node.putNull(field);
            return;
        }
        String actualType = type != null ? type : inferType(value);
        switch (actualType.toLowerCase()) {
            case "string":
                node.put(field, String.valueOf(value));
                break;
            case "int":
            case "integer":
                node.put(field, toInt(value));
                break;
            case "long":
                node.put(field, toLong(value));
                break;
            case "double":
            case "float":
                node.put(field, toDouble(value));
                break;
            case "boolean":
                node.put(field, toBoolean(value));
                break;
            default:
                node.putPOJO(field, value);
        }
    }

    private String inferType(Object value) {
        if (value instanceof String) return "string";
        if (value instanceof Integer) return "int";
        if (value instanceof Long) return "long";
        if (value instanceof Double || value instanceof Float) return "double";
        if (value instanceof Boolean) return "boolean";
        return "string";
    }

    private int toInt(Object value) {
        if (value instanceof Number) return ((Number) value).intValue();
        if (value instanceof String) return Integer.parseInt((String) value);
        return 0;
    }

    private long toLong(Object value) {
        if (value instanceof Number) return ((Number) value).longValue();
        if (value instanceof String) return Long.parseLong((String) value);
        return 0L;
    }

    private double toDouble(Object value) {
        if (value instanceof Number) return ((Number) value).doubleValue();
        if (value instanceof String) return Double.parseDouble((String) value);
        return 0.0;
    }

    private boolean toBoolean(Object value) {
        if (value instanceof Boolean) return (Boolean) value;
        if (value instanceof String) return Boolean.parseBoolean((String) value);
        return false;
    }
}

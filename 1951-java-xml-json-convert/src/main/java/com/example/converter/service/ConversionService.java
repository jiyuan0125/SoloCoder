package com.example.converter.service;

import com.example.converter.exception.ValidationException;
import com.example.converter.model.FieldMapping;
import com.example.converter.model.MappingConfig;
import com.fasterxml.jackson.databind.JsonNode;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.fasterxml.jackson.databind.node.ObjectNode;
import com.fasterxml.jackson.dataformat.xml.XmlMapper;
import com.fasterxml.jackson.dataformat.xml.ser.ToXmlGenerator;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.stereotype.Service;

import java.util.*;

@Service
@RequiredArgsConstructor
@Slf4j
public class ConversionService {

    private final ObjectMapper objectMapper;
    private final XmlMapper xmlMapper;
    private final ValidationService validationService;
    private final MappingService mappingService;

    public String convertXmlToJson(String xmlData, String mappingName, boolean validate) throws Exception {
        MappingConfig mapping = null;
        if (mappingName != null) {
            mapping = mappingService.getMapping(mappingName);
        }

        if (validate && mapping != null && mapping.getXsdSchema() != null) {
            validationService.validateXml(xmlData, mapping.getXsdSchema());
        }

        JsonNode root = xmlMapper.readTree(xmlData);
        JsonNode result = transformXmlToJson(root, mapping);
        return objectMapper.writerWithDefaultPrettyPrinter().writeValueAsString(result);
    }

    public String convertJsonToXml(String jsonData, String mappingName, boolean validate) throws Exception {
        MappingConfig mapping = null;
        if (mappingName != null) {
            mapping = mappingService.getMapping(mappingName);
        }

        if (validate && mapping != null && mapping.getJsonSchema() != null) {
            validationService.validateJson(jsonData, mapping.getJsonSchema());
        }

        JsonNode root = objectMapper.readTree(jsonData);
        String rootElementName = mapping != null ? mapping.getRootElementName() : "root";
        JsonNode transformed = transformJsonToXml(root, mapping);

        String xml = xmlMapper.writer()
                .withRootName(rootElementName)
                .withDefaultPrettyPrinter()
                .writeValueAsString(transformed);
        return "<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n" + xml;
    }

    private JsonNode transformXmlToJson(JsonNode input, MappingConfig mapping) {
        if (input.isObject()) {
            ObjectNode result = objectMapper.createObjectNode();
            String prefix = mapping != null ? mapping.getAttributePrefix() : "";

            Map<String, FieldMapping> fieldMap = new HashMap<>();
            if (mapping != null && mapping.getFieldMappings() != null) {
                for (FieldMapping fm : mapping.getFieldMappings()) {
                    fieldMap.put(fm.getSourceName(), fm);
                }
            }

            Iterator<Map.Entry<String, JsonNode>> fields = input.fields();
            while (fields.hasNext()) {
                Map.Entry<String, JsonNode> entry = fields.next();
                String key = entry.getKey();
                JsonNode value = entry.getValue();

                boolean isAttribute = key.startsWith("@");
                String baseKey = isAttribute ? key.substring(1) : key;

                FieldMapping fieldMapping = fieldMap.get(baseKey);

                String targetKey;
                if (fieldMapping != null && fieldMapping.getTargetName() != null) {
                    targetKey = fieldMapping.getTargetName();
                } else if (isAttribute && !prefix.isEmpty()) {
                    targetKey = prefix + baseKey;
                } else {
                    targetKey = baseKey;
                }

                JsonNode transformedValue = transformXmlToJson(value, mapping);
                JsonNode finalValue = convertType(transformedValue, fieldMapping);

                result.set(targetKey, finalValue);
            }

            if (mapping != null && mapping.getFieldMappings() != null) {
                for (FieldMapping fm : mapping.getFieldMappings()) {
                    if (fm.getDefaultValue() != null && !result.has(fm.getTargetName() != null ? fm.getTargetName() : fm.getSourceName())) {
                        String keyToUse = fm.getTargetName() != null ? fm.getTargetName() : fm.getSourceName();
                        result.set(keyToUse, objectMapper.valueToTree(fm.getDefaultValue()));
                    }
                }
            }

            return result;
        } else if (input.isArray()) {
            List<JsonNode> result = new ArrayList<>();
            for (JsonNode node : input) {
                result.add(transformXmlToJson(node, mapping));
            }
            return objectMapper.valueToTree(result);
        }
        return input;
    }

    private JsonNode transformJsonToXml(JsonNode input, MappingConfig mapping) {
        if (input.isObject()) {
            ObjectNode result = objectMapper.createObjectNode();

            Map<String, FieldMapping> fieldMap = new HashMap<>();
            if (mapping != null && mapping.getFieldMappings() != null) {
                for (FieldMapping fm : mapping.getFieldMappings()) {
                    fieldMap.put(fm.getSourceName(), fm);
                }
            }

            Iterator<Map.Entry<String, JsonNode>> fields = input.fields();
            while (fields.hasNext()) {
                Map.Entry<String, JsonNode> entry = fields.next();
                String key = entry.getKey();
                JsonNode value = entry.getValue();

                FieldMapping fieldMapping = fieldMap.get(key);

                String targetKey;
                if (fieldMapping != null && fieldMapping.getTargetName() != null) {
                    targetKey = fieldMapping.getTargetName();
                } else {
                    targetKey = key;
                }

                if (fieldMapping != null && Boolean.TRUE.equals(fieldMapping.getIsAttribute())) {
                    targetKey = "@" + targetKey;
                }

                JsonNode transformedValue = transformJsonToXml(value, mapping);
                JsonNode finalValue = convertType(transformedValue, fieldMapping);

                result.set(targetKey, finalValue);
            }

            if (mapping != null && mapping.getFieldMappings() != null) {
                for (FieldMapping fm : mapping.getFieldMappings()) {
                    String keyToCheck = fm.getTargetName() != null ? fm.getTargetName() : fm.getSourceName();
                    String actualKey = Boolean.TRUE.equals(fm.getIsAttribute()) ? "@" + keyToCheck : keyToCheck;

                    if (fm.getDefaultValue() != null && !result.has(actualKey)) {
                        result.set(actualKey, objectMapper.valueToTree(fm.getDefaultValue()));
                    }
                }
            }

            return result;
        } else if (input.isArray()) {
            List<JsonNode> result = new ArrayList<>();
            for (JsonNode node : input) {
                result.add(transformJsonToXml(node, mapping));
            }
            return objectMapper.valueToTree(result);
        }
        return input;
    }

    private JsonNode convertType(JsonNode value, FieldMapping fieldMapping) {
        if (fieldMapping == null || fieldMapping.getType() == null) {
            return value;
        }

        try {
            switch (fieldMapping.getType()) {
                case STRING:
                    if (value.isTextual()) return value;
                    return objectMapper.valueToTree(value.asText());
                case INTEGER:
                    if (value.isInt()) return value;
                    return objectMapper.valueToTree(value.asInt());
                case LONG:
                    if (value.isLong()) return value;
                    return objectMapper.valueToTree(value.asLong());
                case DOUBLE:
                case NUMBER:
                    if (value.isNumber()) return value;
                    return objectMapper.valueToTree(value.asDouble());
                case BOOLEAN:
                    if (value.isBoolean()) return value;
                    return objectMapper.valueToTree(value.asBoolean());
                default:
                    return value;
            }
        } catch (Exception e) {
            throw new ValidationException("Type conversion failed for field: " + fieldMapping.getSourceName() + ", expected: " + fieldMapping.getType());
        }
    }
}

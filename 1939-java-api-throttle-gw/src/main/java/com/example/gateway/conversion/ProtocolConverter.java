package com.example.gateway.conversion;

import com.fasterxml.jackson.core.JsonProcessingException;
import com.fasterxml.jackson.databind.JsonNode;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.fasterxml.jackson.databind.node.ObjectNode;
import com.fasterxml.jackson.dataformat.xml.XmlMapper;
import com.fasterxml.jackson.dataformat.xml.deser.FromXmlParser;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

public class ProtocolConverter {
    private static final Logger logger = LoggerFactory.getLogger(ProtocolConverter.class);
    private final ObjectMapper jsonMapper = new ObjectMapper();
    private final XmlMapper xmlMapper = new XmlMapper();

    public ProtocolConverter() {
        xmlMapper.configure(FromXmlParser.Feature.EMPTY_ELEMENT_AS_NULL, true);
    }

    public ConversionResult xmlToJson(String xmlContent) {
        try {
            JsonNode root = xmlMapper.readTree(xmlContent);
            ObjectNode unwrapped = unwrapSingleRoot(root);
            String json = jsonMapper.writeValueAsString(unwrapped);
            logger.debug("XML -> JSON conversion successful");
            return ConversionResult.success(json);
        } catch (JsonProcessingException e) {
            logger.warn("XML to JSON conversion failed: {}", e.getMessage());
            return ConversionResult.failure(parseError(e));
        }
    }

    public ConversionResult jsonToXml(String jsonContent) {
        try {
            JsonNode root = jsonMapper.readTree(jsonContent);
            String xml;
            if (root.isObject() && root.size() == 1) {
                xml = xmlMapper.writer().withRootName(root.fieldNames().next()).writeValueAsString(root);
            } else {
                xml = xmlMapper.writer().withRootName("root").writeValueAsString(root);
            }
            logger.debug("JSON -> XML conversion successful");
            return ConversionResult.success(xml);
        } catch (JsonProcessingException e) {
            logger.warn("JSON to XML conversion failed: {}", e.getMessage());
            return ConversionResult.failure(parseError(e));
        }
    }

    private ObjectNode unwrapSingleRoot(JsonNode root) {
        if (root.isObject() && root.size() == 1) {
            String fieldName = root.fieldNames().next();
            JsonNode child = root.get(fieldName);
            if (child.isObject()) {
                return (ObjectNode) child;
            }
        }
        return (ObjectNode) root;
    }

    private String parseError(JsonProcessingException e) {
        String message = e.getOriginalMessage();
        if (message == null) {
            message = e.getMessage();
        }

        if (message.contains("Unexpected close tag") || message.contains("mismatched")) {
            return "标签不闭合或不匹配: " + message;
        }
        if (message.contains("Unexpected character")) {
            return "意外字符: " + message;
        }
        if (message.contains("Unrecognized token")) {
            return "无法识别的令牌: " + message;
        }
        if (message.contains("Expected")) {
            return "格式错误: " + message;
        }

        return message;
    }

    public static class ConversionResult {
        private final boolean success;
        private final String content;
        private final String error;

        private ConversionResult(boolean success, String content, String error) {
            this.success = success;
            this.content = content;
            this.error = error;
        }

        public static ConversionResult success(String content) {
            return new ConversionResult(true, content, null);
        }

        public static ConversionResult failure(String error) {
            return new ConversionResult(false, null, error);
        }

        public boolean isSuccess() {
            return success;
        }

        public String getContent() {
            return content;
        }

        public String getError() {
            return error;
        }
    }
}

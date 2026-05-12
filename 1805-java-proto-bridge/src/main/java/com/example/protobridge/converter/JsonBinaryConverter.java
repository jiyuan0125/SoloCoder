package com.example.protobridge.converter;

import com.example.protobridge.config.ConversionRule;
import com.example.protobridge.protocol.ProtocolCodec;
import com.example.protobridge.protocol.ProtocolMessage;
import com.fasterxml.jackson.databind.JsonNode;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.fasterxml.jackson.databind.node.ObjectNode;
import lombok.extern.slf4j.Slf4j;
import org.springframework.stereotype.Component;

import java.nio.ByteBuffer;
import java.nio.ByteOrder;
import java.nio.charset.StandardCharsets;
import java.util.List;

@Slf4j
@Component
public class JsonBinaryConverter {

    private final ObjectMapper objectMapper = new ObjectMapper();

    public byte[] jsonToBinary(JsonNode jsonNode, List<ConversionRule.FieldMapping> mappings) {
        if (mappings == null || mappings.isEmpty()) {
            return new byte[0];
        }

        int totalSize = calculateTotalSize(mappings);
        ByteBuffer buffer = ByteBuffer.allocate(totalSize);
        buffer.order(ByteOrder.BIG_ENDIAN);

        for (ConversionRule.FieldMapping mapping : mappings) {
            JsonNode value = jsonNode.get(mapping.getJsonField());
            writeField(buffer, mapping, value);
        }

        return buffer.array();
    }

    public JsonNode binaryToJson(byte[] payload, List<ConversionRule.FieldMapping> mappings) {
        ObjectNode result = objectMapper.createObjectNode();

        if (mappings == null || mappings.isEmpty() || payload == null || payload.length == 0) {
            return result;
        }

        ByteBuffer buffer = ByteBuffer.wrap(payload);
        buffer.order(ByteOrder.BIG_ENDIAN);

        for (ConversionRule.FieldMapping mapping : mappings) {
            Object value = readField(buffer, mapping);
            putValue(result, mapping.getJsonField(), value);
        }

        return result;
    }

    public ProtocolMessage createRequestMessage(byte type, byte[] payload) {
        return ProtocolCodec.createMessage(type, payload);
    }

    public ProtocolMessage parseResponseMessage(ByteBuffer buffer) {
        return ProtocolCodec.decode(buffer);
    }

    private int calculateTotalSize(List<ConversionRule.FieldMapping> mappings) {
        int maxEnd = 0;
        for (ConversionRule.FieldMapping mapping : mappings) {
            int end = mapping.getOffset() + mapping.getLength();
            if (end > maxEnd) {
                maxEnd = end;
            }
        }
        return maxEnd;
    }

    private void writeField(ByteBuffer buffer, ConversionRule.FieldMapping mapping, JsonNode value) {
        buffer.position(mapping.getOffset());
        String type = mapping.getType().toUpperCase();

        switch (type) {
            case "INT":
            case "INTEGER":
                if (value != null && value.isNumber()) {
                    if (mapping.getLength() == 4) {
                        buffer.putInt(value.asInt());
                    } else if (mapping.getLength() == 2) {
                        buffer.putShort((short) value.asInt());
                    } else if (mapping.getLength() == 1) {
                        buffer.put((byte) value.asInt());
                    } else if (mapping.getLength() == 8) {
                        buffer.putLong(value.asLong());
                    }
                }
                break;
            case "LONG":
                if (value != null && value.isNumber()) {
                    buffer.putLong(value.asLong());
                }
                break;
            case "STRING":
                if (value != null && value.isTextual()) {
                    byte[] bytes = value.asText().getBytes(StandardCharsets.UTF_8);
                    int writeLen = Math.min(bytes.length, mapping.getLength());
                    buffer.put(bytes, 0, writeLen);
                    for (int i = writeLen; i < mapping.getLength(); i++) {
                        buffer.put((byte) 0);
                    }
                } else {
                    for (int i = 0; i < mapping.getLength(); i++) {
                        buffer.put((byte) 0);
                    }
                }
                break;
            case "BOOLEAN":
                buffer.put((byte) (value != null && value.asBoolean() ? 1 : 0));
                break;
            default:
                log.warn("不支持的字段类型: {}", type);
        }
    }

    private Object readField(ByteBuffer buffer, ConversionRule.FieldMapping mapping) {
        buffer.position(mapping.getOffset());
        String type = mapping.getType().toUpperCase();

        switch (type) {
            case "INT":
            case "INTEGER":
                if (mapping.getLength() == 4) {
                    return buffer.getInt();
                } else if (mapping.getLength() == 2) {
                    return buffer.getShort();
                } else if (mapping.getLength() == 1) {
                    return buffer.get();
                } else if (mapping.getLength() == 8) {
                    return buffer.getLong();
                }
                return 0;
            case "LONG":
                return buffer.getLong();
            case "STRING":
                byte[] bytes = new byte[mapping.getLength()];
                buffer.get(bytes);
                int len = 0;
                while (len < bytes.length && bytes[len] != 0) {
                    len++;
                }
                return new String(bytes, 0, len, StandardCharsets.UTF_8);
            case "BOOLEAN":
                return buffer.get() != 0;
            default:
                log.warn("不支持的字段类型: {}", type);
                return null;
        }
    }

    private void putValue(ObjectNode node, String field, Object value) {
        if (value == null) {
            node.putNull(field);
        } else if (value instanceof Integer) {
            node.put(field, (Integer) value);
        } else if (value instanceof Long) {
            node.put(field, (Long) value);
        } else if (value instanceof Short) {
            node.put(field, (Short) value);
        } else if (value instanceof Byte) {
            node.put(field, (Byte) value);
        } else if (value instanceof String) {
            node.put(field, (String) value);
        } else if (value instanceof Boolean) {
            node.put(field, (Boolean) value);
        }
    }
}

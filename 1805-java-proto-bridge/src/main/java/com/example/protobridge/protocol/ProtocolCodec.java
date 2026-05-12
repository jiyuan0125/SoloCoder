package com.example.protobridge.protocol;

import java.nio.ByteBuffer;
import java.nio.ByteOrder;

public class ProtocolCodec {

    private ProtocolCodec() {
    }

    public static ByteBuffer encode(ProtocolMessage message) {
        int payloadLength = message.getPayload() != null ? message.getPayload().length : 0;
        int totalSize = ProtocolConstants.HEADER_SIZE + payloadLength;

        ByteBuffer buffer = ByteBuffer.allocate(totalSize);
        buffer.order(ByteOrder.BIG_ENDIAN);

        buffer.putShort(ProtocolConstants.MAGIC_OFFSET, message.getMagic());
        buffer.put(ProtocolConstants.VERSION_OFFSET, message.getVersion());
        buffer.put(ProtocolConstants.TYPE_OFFSET, message.getType());
        buffer.putInt(ProtocolConstants.LENGTH_OFFSET, payloadLength);

        if (message.getPayload() != null && payloadLength > 0) {
            buffer.position(ProtocolConstants.PAYLOAD_OFFSET);
            buffer.put(message.getPayload());
        }

        buffer.flip();
        return buffer;
    }

    public static ProtocolMessage decode(ByteBuffer buffer) {
        buffer.order(ByteOrder.BIG_ENDIAN);

        if (buffer.remaining() < ProtocolConstants.HEADER_SIZE) {
            throw new ProtocolException("数据长度不足，无法解析协议头");
        }

        short magic = buffer.getShort();
        if (magic != ProtocolConstants.MAGIC) {
            throw new ProtocolException("协议头校验失败");
        }

        byte version = buffer.get();
        if (version != ProtocolConstants.VERSION) {
            throw new ProtocolException("不支持的协议版本");
        }

        byte type = buffer.get();
        int length = buffer.getInt();

        if (length < 0) {
            throw new ProtocolException("长度不匹配");
        }

        if (buffer.remaining() < length) {
            throw new ProtocolException("长度不匹配");
        }

        byte[] payload = new byte[length];
        if (length > 0) {
            buffer.get(payload);
        }

        return ProtocolMessage.builder()
                .magic(magic)
                .version(version)
                .type(type)
                .length(length)
                .payload(payload)
                .build();
    }

    public static ProtocolMessage createMessage(byte type, byte[] payload) {
        return ProtocolMessage.builder()
                .magic(ProtocolConstants.MAGIC)
                .version(ProtocolConstants.VERSION)
                .type(type)
                .length(payload != null ? payload.length : 0)
                .payload(payload)
                .build();
    }
}

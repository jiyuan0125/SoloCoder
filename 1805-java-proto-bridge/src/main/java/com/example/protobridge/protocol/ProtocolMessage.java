package com.example.protobridge.protocol;

import lombok.AllArgsConstructor;
import lombok.Builder;
import lombok.Data;
import lombok.NoArgsConstructor;

@Data
@Builder
@NoArgsConstructor
@AllArgsConstructor
public class ProtocolMessage {

    private short magic;
    private byte version;
    private byte type;
    private int length;
    private byte[] payload;

    public int getTotalSize() {
        return ProtocolConstants.HEADER_SIZE + (payload != null ? payload.length : 0);
    }
}

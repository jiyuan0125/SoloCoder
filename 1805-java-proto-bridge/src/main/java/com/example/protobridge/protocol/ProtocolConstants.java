package com.example.protobridge.protocol;

public class ProtocolConstants {

    public static final short MAGIC = (short) 0xABCD;
    public static final byte VERSION = 1;

    public static final int HEADER_SIZE = 8;
    public static final int MAGIC_OFFSET = 0;
    public static final int VERSION_OFFSET = 2;
    public static final int TYPE_OFFSET = 3;
    public static final int LENGTH_OFFSET = 4;
    public static final int PAYLOAD_OFFSET = 8;

    private ProtocolConstants() {
    }
}

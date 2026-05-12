package com.example.protobridge.service;

import com.example.protobridge.protocol.ProtocolCodec;
import com.example.protobridge.protocol.ProtocolMessage;
import lombok.extern.slf4j.Slf4j;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.stereotype.Service;

import java.io.IOException;
import java.io.InputStream;
import java.io.OutputStream;
import java.net.Socket;
import java.nio.ByteBuffer;

@Slf4j
@Service
public class LegacySystemClient {

    @Value("${proto.legacy.host:127.0.0.1}")
    private String host;

    @Value("${proto.legacy.port:9000}")
    private int port;

    @Value("${proto.legacy.connect-timeout:5000}")
    private int connectTimeout;

    @Value("${proto.legacy.read-timeout:30000}")
    private int readTimeout;

    public ProtocolMessage sendAndReceive(ProtocolMessage request) throws IOException {
        try (Socket socket = new Socket()) {
            socket.connect(new java.net.InetSocketAddress(host, port), connectTimeout);
            socket.setSoTimeout(readTimeout);

            try (OutputStream out = socket.getOutputStream();
                 InputStream in = socket.getInputStream()) {

                ByteBuffer requestBuffer = ProtocolCodec.encode(request);
                byte[] requestBytes = new byte[requestBuffer.remaining()];
                requestBuffer.get(requestBytes);
                out.write(requestBytes);
                out.flush();
                log.debug("已发送请求到老系统，长度: {}", requestBytes.length);

                byte[] headerBuffer = new byte[8];
                int headerRead = readFully(in, headerBuffer);
                if (headerRead < 8) {
                    throw new IOException("读取响应头失败，只读取到 " + headerRead + " 字节");
                }

                ByteBuffer headerByteBuffer = ByteBuffer.wrap(headerBuffer);
                int length = headerByteBuffer.getInt(4);
                log.debug("响应头解析完成，payload长度: {}", length);

                byte[] responseBuffer = new byte[8 + length];
                System.arraycopy(headerBuffer, 0, responseBuffer, 0, 8);

                if (length > 0) {
                    int payloadRead = readFully(in, responseBuffer, 8, length);
                    if (payloadRead < length) {
                        throw new IOException("读取响应payload失败，只读取到 " + payloadRead + " 字节");
                    }
                }

                return ProtocolCodec.decode(ByteBuffer.wrap(responseBuffer));
            }
        }
    }

    private int readFully(InputStream in, byte[] buffer) throws IOException {
        return readFully(in, buffer, 0, buffer.length);
    }

    private int readFully(InputStream in, byte[] buffer, int offset, int length) throws IOException {
        int totalRead = 0;
        while (totalRead < length) {
            int read = in.read(buffer, offset + totalRead, length - totalRead);
            if (read == -1) {
                break;
            }
            totalRead += read;
        }
        return totalRead;
    }
}

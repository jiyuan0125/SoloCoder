package com.example.apiproxymirror.filter;

import jakarta.servlet.ServletOutputStream;
import jakarta.servlet.WriteListener;
import jakarta.servlet.http.HttpServletResponse;
import jakarta.servlet.http.HttpServletResponseWrapper;

import java.io.ByteArrayOutputStream;
import java.io.IOException;
import java.io.PrintWriter;

public class CachedBodyHttpServletResponse extends HttpServletResponseWrapper {

    private final ByteArrayOutputStream cachedBodyOutputStream = new ByteArrayOutputStream();
    private ServletOutputStream outputStream;
    private PrintWriter writer;
    private int status = 200;

    public CachedBodyHttpServletResponse(HttpServletResponse response) {
        super(response);
    }

    @Override
    public ServletOutputStream getOutputStream() throws IOException {
        if (writer != null) {
            throw new IllegalStateException("getWriter() has already been called on this response.");
        }
        if (outputStream == null) {
            outputStream = new CachedBodyServletOutputStream(cachedBodyOutputStream, super.getOutputStream());
        }
        return outputStream;
    }

    @Override
    public PrintWriter getWriter() throws IOException {
        if (outputStream != null) {
            throw new IllegalStateException("getOutputStream() has already been called on this response.");
        }
        if (writer == null) {
            writer = new PrintWriter(new CachedBodyWriter(cachedBodyOutputStream, super.getResponse().getWriter()));
        }
        return writer;
    }

    @Override
    public void setStatus(int sc) {
        super.setStatus(sc);
        this.status = sc;
    }



    public int getStatus() {
        return status;
    }

    public byte[] getCachedBody() {
        return cachedBodyOutputStream.toByteArray();
    }

    public String getCachedBodyAsString() {
        return cachedBodyOutputStream.toString();
    }

    private static class CachedBodyServletOutputStream extends ServletOutputStream {

        private final ByteArrayOutputStream cachedOutputStream;
        private final ServletOutputStream originalOutputStream;

        public CachedBodyServletOutputStream(ByteArrayOutputStream cachedOutputStream, ServletOutputStream originalOutputStream) {
            this.cachedOutputStream = cachedOutputStream;
            this.originalOutputStream = originalOutputStream;
        }

        @Override
        public boolean isReady() {
            return true;
        }

        @Override
        public void setWriteListener(WriteListener writeListener) {
        }

        @Override
        public void write(int b) throws IOException {
            cachedOutputStream.write(b);
            originalOutputStream.write(b);
        }

        @Override
        public void write(byte[] b, int off, int len) throws IOException {
            cachedOutputStream.write(b, off, len);
            originalOutputStream.write(b, off, len);
        }

        @Override
        public void flush() throws IOException {
            super.flush();
            originalOutputStream.flush();
        }
    }

    private static class CachedBodyWriter extends java.io.Writer {

        private final ByteArrayOutputStream cachedOutputStream;
        private final PrintWriter originalWriter;

        public CachedBodyWriter(ByteArrayOutputStream cachedOutputStream, PrintWriter originalWriter) {
            this.cachedOutputStream = cachedOutputStream;
            this.originalWriter = originalWriter;
        }

        @Override
        public void write(int c) {
            cachedOutputStream.write(c);
            originalWriter.write(c);
        }

        @Override
        public void write(char[] cbuf, int off, int len) {
            String str = new String(cbuf, off, len);
            byte[] bytes = str.getBytes();
            cachedOutputStream.write(bytes, 0, bytes.length);
            originalWriter.write(cbuf, off, len);
        }

        @Override
        public void write(String str) {
            byte[] bytes = str.getBytes();
            cachedOutputStream.write(bytes, 0, bytes.length);
            originalWriter.write(str);
        }

        @Override
        public void flush() {
            originalWriter.flush();
        }

        @Override
        public void close() {
            originalWriter.close();
        }
    }
}

package com.gateway.core;

import com.fasterxml.jackson.databind.ObjectMapper;
import com.gateway.model.ErrorResponse;
import lombok.extern.slf4j.Slf4j;
import org.springframework.stereotype.Component;

import javax.servlet.http.HttpServletResponse;
import java.io.IOException;
import java.io.PrintWriter;

@Slf4j
@Component
public class ErrorResponseWriter {

    private final ObjectMapper objectMapper = new ObjectMapper();

    public void writeError(HttpServletResponse response, int httpStatus, String errorCode, String message) {
        writeError(response, httpStatus, errorCode, message, null);
    }

    public void writeError(HttpServletResponse response, int httpStatus, String errorCode, String message, String backendUrl) {
        if (response.isCommitted()) {
            log.warn("响应已提交，无法写入错误响应: {} - {}", errorCode, message);
            return;
        }

        response.setStatus(httpStatus);
        response.setContentType("application/json;charset=UTF-8");

        String actualMessage = message;
        if (backendUrl != null && !backendUrl.isEmpty()) {
            actualMessage = message + " [后端: " + backendUrl + "]";
        }

        ErrorResponse errorResponse = new ErrorResponse(errorCode, actualMessage);

        try {
            String json = objectMapper.writeValueAsString(errorResponse);
            PrintWriter writer = response.getWriter();
            writer.write(json);
            writer.flush();
        } catch (IOException e) {
            log.error("写入错误响应失败", e);
        }
    }
}

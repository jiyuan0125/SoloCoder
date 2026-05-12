package com.example.protobridge.controller;

import com.example.protobridge.config.ConversionRule;
import com.example.protobridge.config.ConversionRuleService;
import com.example.protobridge.converter.JsonBinaryConverter;
import com.example.protobridge.exception.ResourceNotFoundException;
import com.example.protobridge.log.ConversionLogService;
import com.example.protobridge.protocol.ProtocolMessage;
import com.example.protobridge.service.LegacySystemClient;
import com.fasterxml.jackson.databind.JsonNode;
import com.fasterxml.jackson.databind.ObjectMapper;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;

import javax.servlet.http.HttpServletRequest;
import java.util.concurrent.TimeUnit;

@Slf4j
@RestController
@RequiredArgsConstructor
public class DynamicProxyController {

    private final ConversionRuleService ruleService;
    private final JsonBinaryConverter converter;
    private final LegacySystemClient legacyClient;
    private final ConversionLogService logService;
    private final ObjectMapper objectMapper = new ObjectMapper();

    @RequestMapping(value = "/**", method = {RequestMethod.GET, RequestMethod.POST, RequestMethod.PUT, RequestMethod.DELETE})
    public ResponseEntity<JsonNode> handleRequest(
            @RequestBody(required = false) String body,
            HttpServletRequest request) {

        String path = request.getRequestURI();
        String method = request.getMethod();

        if (path.startsWith("/api/admin/")) {
            throw new ResourceNotFoundException("接口不存在: " + path);
        }

        ConversionRule rule = ruleService.getRuleByPath(path, method);
        if (rule == null) {
            log.warn("未找到转换规则: {} {}", method, path);
            throw new ResourceNotFoundException("未找到 " + method + " " + path + " 的转换规则");
        }

        long startTime = System.nanoTime();
        String ruleId = rule.getId();

        try {
            JsonNode requestJson = body != null && !body.isEmpty()
                    ? objectMapper.readTree(body)
                    : objectMapper.createObjectNode();

            log.debug("JSON→Binary 转换开始: {} {}", method, path);
            byte[] payload = converter.jsonToBinary(requestJson, rule.getRequestMappings());
            ProtocolMessage requestMessage = converter.createRequestMessage(rule.getRequestType(), payload);

            log.debug("发送请求到老系统，type: 0x{}, payload长度: {}",
                    String.format("%02X", requestMessage.getType()),
                    payload.length);

            long networkStart = System.nanoTime();
            ProtocolMessage responseMessage = legacyClient.sendAndReceive(requestMessage);
            long networkEnd = System.nanoTime();
            log.debug("收到老系统响应，type: 0x{}, payload长度: {}, 网络耗时: {}ms",
                    String.format("%02X", responseMessage.getType()),
                    responseMessage.getPayload() != null ? responseMessage.getPayload().length : 0,
                    TimeUnit.NANOSECONDS.toMillis(networkEnd - networkStart));

            log.debug("Binary→JSON 转换开始");
            JsonNode responseJson = converter.binaryToJson(
                    responseMessage.getPayload(),
                    rule.getResponseMappings());

            long totalDuration = TimeUnit.NANOSECONDS.toMillis(System.nanoTime() - startTime);

            logService.recordSuccess(
                    "JSON→Binary→JSON",
                    path,
                    method,
                    totalDuration,
                    ruleId);

            log.info("转换完成: {} {}, 总耗时: {}ms", method, path, totalDuration);
            return ResponseEntity.ok(responseJson);

        } catch (ResourceNotFoundException e) {
            throw e;
        } catch (Exception e) {
            long totalDuration = TimeUnit.NANOSECONDS.toMillis(System.nanoTime() - startTime);
            log.error("转换失败: {} {}, 耗时: {}ms, 错误: {}", method, path, totalDuration, e.getMessage(), e);

            logService.recordFailure(
                    "JSON→Binary→JSON",
                    path,
                    method,
                    totalDuration,
                    e.getMessage(),
                    ruleId);

            if (e instanceof RuntimeException) {
                throw (RuntimeException) e;
            }
            throw new RuntimeException("转换失败: " + e.getMessage(), e);
        }
    }
}

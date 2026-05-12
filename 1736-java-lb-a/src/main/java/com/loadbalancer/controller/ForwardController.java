package com.loadbalancer.controller;

import com.loadbalancer.service.RequestForwarder;
import jakarta.servlet.http.HttpServletRequest;
import org.springframework.http.*;
import org.springframework.web.bind.annotation.*;
import org.springframework.web.servlet.HandlerMapping;

import java.io.IOException;
import java.util.Enumeration;
import java.util.List;

@RestController
public class ForwardController {

    private final RequestForwarder requestForwarder;

    public ForwardController(RequestForwarder requestForwarder) {
        this.requestForwarder = requestForwarder;
    }

    @RequestMapping("/**")
    public ResponseEntity<byte[]> forward(HttpServletRequest request) throws IOException {
        String path = (String) request.getAttribute(HandlerMapping.PATH_WITHIN_HANDLER_MAPPING_ATTRIBUTE);
        
        HttpMethod method = HttpMethod.valueOf(request.getMethod());
        
        HttpHeaders headers = new HttpHeaders();
        Enumeration<String> headerNames = request.getHeaderNames();
        while (headerNames.hasMoreElements()) {
            String headerName = headerNames.nextElement();
            Enumeration<String> headerValues = request.getHeaders(headerName);
            while (headerValues.hasMoreElements()) {
                headers.add(headerName, headerValues.nextElement());
            }
        }
        
        byte[] body = request.getInputStream().readAllBytes();
        
        return requestForwarder.forwardRequest(path, method, headers, body);
    }
}

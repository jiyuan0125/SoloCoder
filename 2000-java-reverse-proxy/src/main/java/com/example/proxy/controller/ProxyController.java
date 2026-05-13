package com.example.proxy.controller;

import com.example.proxy.service.ProxyService;
import lombok.extern.slf4j.Slf4j;
import org.springframework.web.bind.annotation.RequestMapping;
import org.springframework.web.bind.annotation.RestController;

import javax.servlet.http.HttpServletRequest;
import javax.servlet.http.HttpServletResponse;
import java.io.IOException;

@Slf4j
@RestController
public class ProxyController {
    
    private final ProxyService proxyService;
    
    public ProxyController(ProxyService proxyService) {
        this.proxyService = proxyService;
    }
    
    @RequestMapping("/**")
    public void proxyRequest(HttpServletRequest request, HttpServletResponse response) throws IOException {
        proxyService.forwardRequest(request, response);
    }
}

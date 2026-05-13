package com.loadbalancer.filter;

import com.loadbalancer.service.RequestForwardingService;
import jakarta.servlet.*;
import jakarta.servlet.http.HttpServletRequest;
import jakarta.servlet.http.HttpServletResponse;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.stereotype.Component;

import java.io.IOException;
import java.util.List;

@Slf4j
@Component
@RequiredArgsConstructor
public class RequestForwardingFilter implements Filter {
    private final RequestForwardingService requestForwardingService;
    
    private static final List<String> EXCLUDED_PATHS = List.of(
            "/nodes",
            "/dashboard"
    );

    @Override
    public void doFilter(ServletRequest request, ServletResponse response, FilterChain chain)
            throws IOException, ServletException {
        
        HttpServletRequest httpRequest = (HttpServletRequest) request;
        String path = httpRequest.getRequestURI();
        
        if (shouldExclude(path)) {
            chain.doFilter(request, response);
            return;
        }
        
        log.debug("Forwarding request to backend: {} {}", httpRequest.getMethod(), path);
        requestForwardingService.forwardRequest(httpRequest, (HttpServletResponse) response);
    }

    private boolean shouldExclude(String path) {
        for (String excluded : EXCLUDED_PATHS) {
            if (path.startsWith(excluded)) {
                return true;
            }
        }
        return false;
    }
}

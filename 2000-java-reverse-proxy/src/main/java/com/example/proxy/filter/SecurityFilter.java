package com.example.proxy.filter;

import lombok.extern.slf4j.Slf4j;
import org.springframework.core.Ordered;
import org.springframework.core.annotation.Order;
import org.springframework.stereotype.Component;

import javax.servlet.*;
import javax.servlet.http.HttpServletRequest;
import javax.servlet.http.HttpServletResponse;
import java.io.IOException;
import java.util.regex.Pattern;

@Slf4j
@Component
@Order(Ordered.HIGHEST_PRECEDENCE)
public class SecurityFilter implements Filter {
    
    private static final Pattern PATH_TRAVERSAL_PATTERN = Pattern.compile("\\.\\.");
    private static final Pattern SPECIAL_CHAR_PATTERN = Pattern.compile("[<>\"'%;()&+]");
    private static final Pattern NULL_BYTE_PATTERN = Pattern.compile("%00|\\x00");
    
    @Override
    public void doFilter(ServletRequest request, ServletResponse response, FilterChain chain) 
            throws IOException, ServletException {
        
        HttpServletRequest httpRequest = (HttpServletRequest) request;
        HttpServletResponse httpResponse = (HttpServletResponse) response;
        
        String requestUri = httpRequest.getRequestURI();
        String queryString = httpRequest.getQueryString();
        
        if (!isPathSafe(requestUri)) {
            log.warn("Blocked unsafe request path: {}", requestUri);
            sendForbiddenResponse(httpResponse, "Unsafe path detected");
            return;
        }
        
        if (queryString != null && !isPathSafe(queryString)) {
            log.warn("Blocked unsafe query string: {}", queryString);
            sendForbiddenResponse(httpResponse, "Unsafe query string detected");
            return;
        }
        
        chain.doFilter(request, response);
    }
    
    private boolean isPathSafe(String path) {
        if (path == null || path.isEmpty()) {
            return true;
        }
        
        String decodedPath = java.net.URLDecoder.decode(path, java.nio.charset.StandardCharsets.UTF_8);
        
        if (PATH_TRAVERSAL_PATTERN.matcher(decodedPath).find()) {
            log.debug("Path traversal detected: {}", decodedPath);
            return false;
        }
        
        if (SPECIAL_CHAR_PATTERN.matcher(decodedPath).find()) {
            log.debug("Special character detected: {}", decodedPath);
            return false;
        }
        
        if (NULL_BYTE_PATTERN.matcher(decodedPath).find()) {
            log.debug("Null byte detected: {}", decodedPath);
            return false;
        }
        
        return true;
    }
    
    private void sendForbiddenResponse(HttpServletResponse response, String message) throws IOException {
        response.setStatus(HttpServletResponse.SC_FORBIDDEN);
        response.setContentType("application/json");
        response.getWriter().write(String.format("{\"error\": \"%s\", \"status\": 403}", message));
    }
    
    @Override
    public void init(FilterConfig filterConfig) {
        log.info("SecurityFilter initialized");
    }
    
    @Override
    public void destroy() {
        log.info("SecurityFilter destroyed");
    }
}

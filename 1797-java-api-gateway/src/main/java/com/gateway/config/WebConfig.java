package com.gateway.config;

import com.gateway.core.GatewayServlet;
import org.springframework.beans.factory.annotation.Value;
import org.springframework.context.annotation.Configuration;
import org.springframework.web.servlet.config.annotation.InterceptorRegistry;
import org.springframework.web.servlet.config.annotation.WebMvcConfigurer;

@Configuration
public class WebConfig implements WebMvcConfigurer {

    private final GatewayServlet gatewayServlet;
    private final String adminPrefix;

    public WebConfig(GatewayServlet gatewayServlet,
                     @Value("${gateway.admin.prefix:/admin}") String adminPrefix) {
        this.gatewayServlet = gatewayServlet;
        this.adminPrefix = adminPrefix;
    }

    @Override
    public void addInterceptors(InterceptorRegistry registry) {
        registry.addInterceptor(gatewayServlet)
                .addPathPatterns("/**")
                .excludePathPatterns(adminPrefix + "/**", "/error");
    }
}

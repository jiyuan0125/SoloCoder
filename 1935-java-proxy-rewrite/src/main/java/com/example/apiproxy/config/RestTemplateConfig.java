package com.example.apiproxy.config;

import org.springframework.boot.web.client.RestTemplateBuilder;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;
import org.springframework.http.client.ClientHttpRequestFactory;
import org.springframework.http.client.SimpleClientHttpRequestFactory;
import org.springframework.web.client.RestTemplate;

@Configuration
public class RestTemplateConfig {

    @Bean
    public RestTemplate restTemplate(ProxyProperties properties, RestTemplateBuilder builder) {
        ClientHttpRequestFactory factory = createRequestFactory(properties.getTimeout());
        return builder.requestFactory(() -> factory).build();
    }

    private ClientHttpRequestFactory createRequestFactory(int timeoutMs) {
        SimpleClientHttpRequestFactory factory = new SimpleClientHttpRequestFactory();
        factory.setConnectTimeout(timeoutMs);
        factory.setReadTimeout(timeoutMs);
        return factory;
    }
}

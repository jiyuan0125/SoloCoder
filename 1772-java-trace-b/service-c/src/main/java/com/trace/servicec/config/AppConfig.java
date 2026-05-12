package com.trace.servicec.config;

import io.micrometer.tracing.Tracer;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;
import org.springframework.web.client.RestClient;
import org.springframework.web.client.support.RestClientAdapter;
import org.springframework.web.service.invoker.HttpServiceProxyFactory;

@Configuration
public class AppConfig {

    @Bean
    public RestClient.Builder restClientBuilder(Tracer tracer) {
        return RestClient.builder();
    }

    @Bean
    public ServiceDClient serviceDClient(RestClient.Builder builder) {
        var restClient = builder.baseUrl("http://localhost:8084").build();
        var factory = HttpServiceProxyFactory.builderFor(RestClientAdapter.create(restClient)).build();
        return factory.createClient(ServiceDClient.class);
    }

}

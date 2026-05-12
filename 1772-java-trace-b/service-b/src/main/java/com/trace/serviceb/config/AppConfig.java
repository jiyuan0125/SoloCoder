package com.trace.serviceb.config;

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
    public ServiceCClient serviceCClient(RestClient.Builder builder) {
        var restClient = builder.baseUrl("http://localhost:8083").build();
        var factory = HttpServiceProxyFactory.builderFor(RestClientAdapter.create(restClient)).build();
        return factory.createClient(ServiceCClient.class);
    }

}

package com.purchase.approval.config;

import com.fasterxml.jackson.databind.ObjectMapper;
import com.fasterxml.jackson.databind.SerializationFeature;
import com.fasterxml.jackson.datatype.hibernate5.Hibernate5Module;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;
import org.springframework.http.converter.json.Jackson2ObjectMapperBuilder;
import org.springframework.http.converter.json.MappingJackson2HttpMessageConverter;

@Configuration
public class JacksonConfig {

    @Bean
    public MappingJackson2HttpMessageConverter mappingJackson2HttpMessageConverter() {
        ObjectMapper mapper = Jackson2ObjectMapperBuilder.json().build();
        
        Hibernate5Module hibernate5Module = new Hibernate5Module();
        hibernate5Module.enable(Hibernate5Module.Feature.SERIALIZE_IDENTIFIER_FOR_LAZY_NOT_LOADED_OBJECTS);
        hibernate5Module.disable(Hibernate5Module.Feature.FORCE_LAZY_LOADING);
        
        mapper.registerModule(hibernate5Module);
        mapper.disable(SerializationFeature.FAIL_ON_EMPTY_BEANS);
        
        return new MappingJackson2HttpMessageConverter(mapper);
    }
}

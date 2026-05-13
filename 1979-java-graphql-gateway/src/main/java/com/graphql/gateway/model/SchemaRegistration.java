package com.graphql.gateway.model;

public class SchemaRegistration {
    private String serviceName;
    private String endpoint;
    private String schema;

    public SchemaRegistration() {
    }

    public SchemaRegistration(String serviceName, String endpoint, String schema) {
        this.serviceName = serviceName;
        this.endpoint = endpoint;
        this.schema = schema;
    }

    public String getServiceName() {
        return serviceName;
    }

    public void setServiceName(String serviceName) {
        this.serviceName = serviceName;
    }

    public String getEndpoint() {
        return endpoint;
    }

    public void setEndpoint(String endpoint) {
        this.endpoint = endpoint;
    }

    public String getSchema() {
        return schema;
    }

    public void setSchema(String schema) {
        this.schema = schema;
    }
}

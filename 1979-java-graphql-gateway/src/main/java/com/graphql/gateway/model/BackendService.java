package com.graphql.gateway.model;

import java.util.HashSet;
import java.util.Set;

public class BackendService {
    private String name;
    private String endpoint;
    private Set<String> rootQueryFields = new HashSet<>();
    private Set<String> rootMutationFields = new HashSet<>();

    public BackendService() {
    }

    public BackendService(String name, String endpoint) {
        this.name = name;
        this.endpoint = endpoint;
    }

    public String getName() {
        return name;
    }

    public void setName(String name) {
        this.name = name;
    }

    public String getEndpoint() {
        return endpoint;
    }

    public void setEndpoint(String endpoint) {
        this.endpoint = endpoint;
    }

    public Set<String> getRootQueryFields() {
        return rootQueryFields;
    }

    public void setRootQueryFields(Set<String> rootQueryFields) {
        this.rootQueryFields = rootQueryFields;
    }

    public Set<String> getRootMutationFields() {
        return rootMutationFields;
    }

    public void setRootMutationFields(Set<String> rootMutationFields) {
        this.rootMutationFields = rootMutationFields;
    }

    public void addRootQueryField(String field) {
        this.rootQueryFields.add(field);
    }

    public void addRootMutationField(String field) {
        this.rootMutationFields.add(field);
    }
}

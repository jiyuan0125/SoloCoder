package com.graphql.gateway.schema;

import com.graphql.gateway.model.BackendService;
import com.graphql.gateway.model.SchemaRegistration;
import graphql.language.*;
import graphql.parser.Parser;
import graphql.schema.idl.TypeDefinitionRegistry;
import graphql.schema.idl.SchemaParser;

import java.util.*;
import java.util.concurrent.ConcurrentHashMap;
import java.util.stream.Collectors;

public class SchemaManager {
    private static SchemaManager instance;
    private final Map<String, BackendService> backendServices = new ConcurrentHashMap<>();
    private final Map<String, List<String>> typeToService = new ConcurrentHashMap<>();
    private String mergedSchema = "";
    private TypeDefinitionRegistry mergedRegistry;

    private SchemaManager() {
    }

    public static synchronized SchemaManager getInstance() {
        if (instance == null) {
            instance = new SchemaManager();
        }
        return instance;
    }

    public synchronized SchemaRegistrationResult registerSchema(SchemaRegistration registration) {
        if (registration.getServiceName() == null || registration.getServiceName().isEmpty()) {
            return SchemaRegistrationResult.error("serviceName is required");
        }
        if (registration.getEndpoint() == null || registration.getEndpoint().isEmpty()) {
            return SchemaRegistrationResult.error("endpoint is required");
        }
        if (registration.getSchema() == null || registration.getSchema().isEmpty()) {
            return SchemaRegistrationResult.error("schema is required");
        }

        Document document;
        try {
            document = Parser.parse(registration.getSchema());
        } catch (Exception e) {
            return SchemaRegistrationResult.error("Invalid GraphQL schema: " + e.getMessage());
        }

        Map<String, BackendService> tempServices = new ConcurrentHashMap<>(backendServices);
        Map<String, List<String>> tempTypeToService = new ConcurrentHashMap<>(typeToService);

        BackendService backend = new BackendService(registration.getServiceName(), registration.getEndpoint());

        for (Definition<?> def : document.getDefinitions()) {
            if (def instanceof ObjectTypeDefinition) {
                ObjectTypeDefinition objDef = (ObjectTypeDefinition) def;
                String typeName = objDef.getName();

                if (typeName.equals("Query") || typeName.equals("Mutation")) {
                    for (FieldDefinition field : objDef.getFieldDefinitions()) {
                        if (typeName.equals("Query")) {
                            backend.addRootQueryField(field.getName());
                        } else {
                            backend.addRootMutationField(field.getName());
                        }
                    }
                } else {
                    tempTypeToService.computeIfAbsent(typeName, k -> new ArrayList<>())
                            .add(registration.getServiceName());
                }
            }
        }

        SchemaConflictResult conflict = checkConflicts(registration.getServiceName(), tempTypeToService);
        if (conflict.hasConflict) {
            return SchemaRegistrationResult.conflict(conflict.message);
        }

        tempServices.put(registration.getServiceName(), backend);
        backendServices.putAll(tempServices);
        typeToService.putAll(tempTypeToService);

        try {
            mergeAllSchemas();
        } catch (Exception e) {
            backendServices.remove(registration.getServiceName());
            for (Map.Entry<String, List<String>> entry : tempTypeToService.entrySet()) {
                entry.getValue().remove(registration.getServiceName());
                if (entry.getValue().isEmpty()) {
                    typeToService.remove(entry.getKey());
                }
            }
            return SchemaRegistrationResult.error("Failed to merge schemas: " + e.getMessage());
        }

        return SchemaRegistrationResult.success();
    }

    private SchemaConflictResult checkConflicts(String newService, Map<String, List<String>> tempTypeToService) {
        for (Map.Entry<String, List<String>> entry : tempTypeToService.entrySet()) {
            String typeName = entry.getKey();
            List<String> services = entry.getValue();

            if (services.size() > 1 && services.contains(newService)) {
                List<String> otherServices = services.stream()
                        .filter(s -> !s.equals(newService))
                        .collect(Collectors.toList());
                if (!otherServices.isEmpty()) {
                    return new SchemaConflictResult(true,
                            "Type conflict for '" + typeName + "': defined in services " +
                                    String.join(", ", services) +
                                    ". Please configure conflict resolution rules manually.");
                }
            }
        }
        return new SchemaConflictResult(false, null);
    }

    private void mergeAllSchemas() {
        StringBuilder sb = new StringBuilder();
        Set<String> typeNames = new HashSet<>();

        sb.append("type Query {\n");
        for (BackendService backend : backendServices.values()) {
            for (String field : backend.getRootQueryFields()) {
                if (!typeNames.contains(field)) {
                    sb.append("  ").append(field).append(": JSON\n");
                    typeNames.add(field);
                }
            }
        }
        sb.append("}\n\n");

        if (!backendServices.values().stream()
                .flatMap(b -> b.getRootMutationFields().stream())
                .collect(Collectors.toSet()).isEmpty()) {
            sb.append("type Mutation {\n");
            Set<String> mutationFields = new HashSet<>();
            for (BackendService backend : backendServices.values()) {
                for (String field : backend.getRootMutationFields()) {
                    if (!mutationFields.contains(field)) {
                        sb.append("  ").append(field).append(": JSON\n");
                        mutationFields.add(field);
                    }
                }
            }
            sb.append("}\n\n");
        }

        sb.append("scalar JSON\n");
        sb.append("scalar String\n");
        sb.append("scalar Int\n");
        sb.append("scalar Float\n");
        sb.append("scalar Boolean\n");
        sb.append("scalar ID\n");

        mergedSchema = sb.toString();
        mergedRegistry = new SchemaParser().parse(mergedSchema);
    }

    public String getMergedSchema() {
        return mergedSchema;
    }

    public TypeDefinitionRegistry getMergedRegistry() {
        return mergedRegistry;
    }

    public BackendService getBackendForQueryField(String fieldName) {
        for (BackendService backend : backendServices.values()) {
            if (backend.getRootQueryFields().contains(fieldName)) {
                return backend;
            }
        }
        return null;
    }

    public BackendService getBackendForMutationField(String fieldName) {
        for (BackendService backend : backendServices.values()) {
            if (backend.getRootMutationFields().contains(fieldName)) {
                return backend;
            }
        }
        return null;
    }

    public Collection<BackendService> getAllBackends() {
        return backendServices.values();
    }

    public static class SchemaRegistrationResult {
        private final boolean success;
        private final String errorMessage;
        private final boolean isConflict;

        private SchemaRegistrationResult(boolean success, String errorMessage, boolean isConflict) {
            this.success = success;
            this.errorMessage = errorMessage;
            this.isConflict = isConflict;
        }

        public static SchemaRegistrationResult success() {
            return new SchemaRegistrationResult(true, null, false);
        }

        public static SchemaRegistrationResult error(String message) {
            return new SchemaRegistrationResult(false, message, false);
        }

        public static SchemaRegistrationResult conflict(String message) {
            return new SchemaRegistrationResult(false, message, true);
        }

        public boolean isSuccess() {
            return success;
        }

        public String getErrorMessage() {
            return errorMessage;
        }

        public boolean isConflict() {
            return isConflict;
        }
    }

    private static class SchemaConflictResult {
        final boolean hasConflict;
        final String message;

        SchemaConflictResult(boolean hasConflict, String message) {
            this.hasConflict = hasConflict;
            this.message = message;
        }
    }
}

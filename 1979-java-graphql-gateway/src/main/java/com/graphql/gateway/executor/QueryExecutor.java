package com.graphql.gateway.executor;

import com.fasterxml.jackson.databind.JsonNode;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.fasterxml.jackson.databind.node.ObjectNode;
import com.graphql.gateway.config.GatewayConfig;
import com.graphql.gateway.model.BackendService;
import com.graphql.gateway.model.QueryStats;
import com.graphql.gateway.parser.QueryParser;
import com.graphql.gateway.schema.SchemaManager;
import graphql.language.*;

import java.net.URI;
import java.net.http.HttpClient;
import java.net.http.HttpRequest;
import java.net.http.HttpResponse;
import java.time.Duration;
import java.util.*;
import java.util.concurrent.*;

public class QueryExecutor {
    private static QueryExecutor instance;
    private final HttpClient httpClient;
    private final ObjectMapper objectMapper;
    private final int timeoutSeconds;
    private final QueryStats stats;
    private final SchemaManager schemaManager;

    private QueryExecutor() {
        this.httpClient = HttpClient.newBuilder()
                .executor(Executors.newFixedThreadPool(20))
                .build();
        this.objectMapper = new ObjectMapper();
        this.timeoutSeconds = GatewayConfig.getInstance().getBackendTimeoutSeconds();
        this.stats = new QueryStats();
        this.schemaManager = SchemaManager.getInstance();
    }

    public static synchronized QueryExecutor getInstance() {
        if (instance == null) {
            instance = new QueryExecutor();
        }
        return instance;
    }

    public QueryStats getStats() {
        return stats;
    }

    public ExecutionResult execute(QueryParser.ParseResult parseResult, JsonNode variables) {
        long parseTimeNanos = 0;
        long startTimeNanos = System.nanoTime();

        try {
            ObjectNode mergedData = objectMapper.createObjectNode();
            List<String> errors = new ArrayList<>();

            for (QueryParser.QueryPartition partition : parseResult.getPartitions()) {
                Map<String, List<Field>> fieldsByRoot = partition.getFieldsByRootField();
                Map<BackendService, Map<String, List<Field>>> groupedFields = new HashMap<>();

                for (Map.Entry<String, List<Field>> entry : fieldsByRoot.entrySet()) {
                    String rootFieldName = entry.getKey();
                    List<Field> fields = entry.getValue();

                    BackendService backend = partition.isMutation()
                            ? schemaManager.getBackendForMutationField(rootFieldName)
                            : schemaManager.getBackendForQueryField(rootFieldName);

                    if (backend == null) {
                        errors.add("Field '" + rootFieldName + "' not found in any registered backend service");
                        continue;
                    }

                    groupedFields.computeIfAbsent(backend, k -> new HashMap<>())
                            .put(rootFieldName, fields);
                }

                if (!groupedFields.isEmpty()) {
                    Map<BackendService, Future<BackendResponse>> futures = new HashMap<>();
                    ExecutorService executor = Executors.newFixedThreadPool(groupedFields.size());

                    for (Map.Entry<BackendService, Map<String, List<Field>>> entry : groupedFields.entrySet()) {
                        BackendService backend = entry.getKey();
                        Map<String, List<Field>> backendFields = entry.getValue();

                        Callable<BackendResponse> task = () -> executeBackendCall(backend, backendFields, variables);
                        futures.put(backend, executor.submit(task));
                    }

                    executor.shutdown();
                    try {
                        executor.awaitTermination(timeoutSeconds, TimeUnit.SECONDS);
                    } catch (InterruptedException e) {
                        Thread.currentThread().interrupt();
                    }

                    for (Map.Entry<BackendService, Future<BackendResponse>> entry : futures.entrySet()) {
                        BackendService backend = entry.getKey();
                        try {
                            BackendResponse response = entry.getValue().get(timeoutSeconds, TimeUnit.SECONDS);
                            stats.recordBackendCall(backend.getName(), response.durationNanos, response.success);

                            if (response.success && response.data != null) {
                                Iterator<String> fieldNames = response.data.fieldNames();
                                while (fieldNames.hasNext()) {
                                    String fieldName = fieldNames.next();
                                    mergedData.set(fieldName, response.data.get(fieldName));
                                }
                            } else if (response.errorMessage != null) {
                                errors.add("Backend '" + backend.getName() + "' error: " + response.errorMessage);
                            }
                        } catch (TimeoutException e) {
                            stats.recordBackendCall(backend.getName(), 0, false);
                            errors.add("Backend '" + backend.getName() + "' timeout");
                        } catch (Exception e) {
                            stats.recordBackendCall(backend.getName(), 0, false);
                            errors.add("Backend '" + backend.getName() + "' failed: " + e.getMessage());
                        }
                    }
                }
            }

            long executionTimeNanos = System.nanoTime() - startTimeNanos;
            stats.recordQuery(parseTimeNanos, executionTimeNanos);

            return new ExecutionResult(mergedData, errors);

        } catch (Exception e) {
            return new ExecutionResult(null, Collections.singletonList("Execution error: " + e.getMessage()));
        }
    }

    private BackendResponse executeBackendCall(BackendService backend, Map<String, List<Field>> fields,
                                                JsonNode variables) {
        long startTimeNanos = System.nanoTime();
        try {
            String subQuery = buildSubQuery(fields);
            ObjectNode requestBody = objectMapper.createObjectNode();
            requestBody.put("query", subQuery);
            if (variables != null && !variables.isNull()) {
                requestBody.set("variables", variables);
            }

            String bodyString = objectMapper.writeValueAsString(requestBody);

            HttpRequest request = HttpRequest.newBuilder()
                    .uri(URI.create(backend.getEndpoint()))
                    .header("Content-Type", "application/json")
                    .timeout(Duration.ofSeconds(timeoutSeconds))
                    .POST(HttpRequest.BodyPublishers.ofString(bodyString))
                    .build();

            HttpResponse<String> response = httpClient.send(request, HttpResponse.BodyHandlers.ofString());

            long durationNanos = System.nanoTime() - startTimeNanos;

            if (response.statusCode() >= 200 && response.statusCode() < 300) {
                JsonNode responseJson = objectMapper.readTree(response.body());
                JsonNode data = responseJson.get("data");
                JsonNode errors = responseJson.get("errors");

                if (errors != null && errors.isArray() && errors.size() > 0) {
                    List<String> errorMsgs = new ArrayList<>();
                    for (JsonNode error : errors) {
                        errorMsgs.add(error.path("message").asText(error.toString()));
                    }
                    return new BackendResponse(false, data, String.join("; ", errorMsgs), durationNanos);
                }

                return new BackendResponse(true, data, null, durationNanos);
            } else {
                return new BackendResponse(false, null,
                        "HTTP " + response.statusCode() + ": " + response.body(),
                        System.nanoTime() - startTimeNanos);
            }
        } catch (Exception e) {
            return new BackendResponse(false, null, e.getMessage(),
                    System.nanoTime() - startTimeNanos);
        }
    }

    private String buildSubQuery(Map<String, List<Field>> fieldsByRoot) {
        StringBuilder sb = new StringBuilder("{ ");

        for (Map.Entry<String, List<Field>> entry : fieldsByRoot.entrySet()) {
            String rootField = entry.getKey();
            List<Field> fields = entry.getValue();

            for (Field field : fields) {
                sb.append(fieldToString(field)).append(" ");
            }
        }

        sb.append("}");
        return sb.toString();
    }

    private String fieldToString(Field field) {
        StringBuilder sb = new StringBuilder();

        if (field.getAlias() != null) {
            sb.append(field.getAlias()).append(": ");
        }

        sb.append(field.getName());

        if (!field.getArguments().isEmpty()) {
            sb.append("(");
            boolean firstArg = true;
            for (Argument arg : field.getArguments()) {
                if (!firstArg) sb.append(", ");
                sb.append(arg.getName()).append(": ").append(valueToString(arg.getValue()));
                firstArg = false;
            }
            sb.append(")");
        }

        SelectionSet selectionSet = field.getSelectionSet();
        if (selectionSet != null) {
            sb.append(" { ");
            for (Selection<?> selection : selectionSet.getSelections()) {
                if (selection instanceof Field) {
                    sb.append(fieldToString((Field) selection)).append(" ");
                }
            }
            sb.append("}");
        }

        return sb.toString();
    }

    private String valueToString(Value<?> value) {
        if (value instanceof StringValue) {
            return "\"" + ((StringValue) value).getValue().replace("\"", "\\\"") + "\"";
        } else if (value instanceof IntValue) {
            return String.valueOf(((IntValue) value).getValue());
        } else if (value instanceof FloatValue) {
            return String.valueOf(((FloatValue) value).getValue());
        } else if (value instanceof BooleanValue) {
            return String.valueOf(((BooleanValue) value).isValue());
        } else if (value instanceof NullValue) {
            return "null";
        } else if (value instanceof VariableReference) {
            return "$" + ((VariableReference) value).getName();
        } else if (value instanceof EnumValue) {
            return ((EnumValue) value).getName();
        } else if (value instanceof ObjectValue) {
            ObjectValue obj = (ObjectValue) value;
            StringBuilder sb = new StringBuilder("{ ");
            boolean first = true;
            for (ObjectField field : obj.getObjectFields()) {
                if (!first) sb.append(", ");
                sb.append(field.getName()).append(": ").append(valueToString(field.getValue()));
                first = false;
            }
            sb.append(" }");
            return sb.toString();
        } else if (value instanceof ArrayValue) {
            ArrayValue arr = (ArrayValue) value;
            StringBuilder sb = new StringBuilder("[ ");
            boolean first = true;
            for (Value<?> v : arr.getValues()) {
                if (!first) sb.append(", ");
                sb.append(valueToString(v));
                first = false;
            }
            sb.append(" ]");
            return sb.toString();
        }
        return value.toString();
    }

    public static class ExecutionResult {
        private final ObjectNode data;
        private final List<String> errors;

        public ExecutionResult(ObjectNode data, List<String> errors) {
            this.data = data;
            this.errors = errors;
        }

        public ObjectNode getData() {
            return data;
        }

        public List<String> getErrors() {
            return errors;
        }

        public boolean hasErrors() {
            return errors != null && !errors.isEmpty();
        }
    }

    private static class BackendResponse {
        final boolean success;
        final ObjectNode data;
        final String errorMessage;
        final long durationNanos;

        BackendResponse(boolean success, JsonNode data, String errorMessage, long durationNanos) {
            this.success = success;
            this.data = (data != null && data instanceof ObjectNode) ? (ObjectNode) data : null;
            this.errorMessage = errorMessage;
            this.durationNanos = durationNanos;
        }
    }
}

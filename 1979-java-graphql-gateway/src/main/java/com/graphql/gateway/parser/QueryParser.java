package com.graphql.gateway.parser;

import com.graphql.gateway.config.GatewayConfig;
import graphql.language.*;
import graphql.parser.Parser;

import java.util.*;

public class QueryParser {
    private static QueryParser instance;
    private final int maxDepth;
    private final int maxFields;

    private QueryParser() {
        GatewayConfig config = GatewayConfig.getInstance();
        this.maxDepth = config.getMaxQueryDepth();
        this.maxFields = config.getMaxFieldCount();
    }

    public static synchronized QueryParser getInstance() {
        if (instance == null) {
            instance = new QueryParser();
        }
        return instance;
    }

    public ParseResult parse(String queryString) {
        Document document;
        try {
            document = Parser.parse(queryString);
        } catch (Exception e) {
            return ParseResult.error("Parse error: " + e.getMessage());
        }

        int depth = calculateMaxDepth(document);
        if (depth > maxDepth) {
            return ParseResult.error(
                    String.format("Query exceeds maximum depth. Max: %d, Actual: %d", maxDepth, depth));
        }

        int fieldCount = countFields(document);
        if (fieldCount > maxFields) {
            return ParseResult.error(
                    String.format("Query exceeds maximum field count. Max: %d, Actual: %d", maxFields, fieldCount));
        }

        List<QueryPartition> partitions = partitionQuery(document);
        return ParseResult.success(partitions, depth, fieldCount);
    }

    private int calculateMaxDepth(Document document) {
        int max = 0;
        for (Definition<?> def : document.getDefinitions()) {
            if (def instanceof OperationDefinition) {
                OperationDefinition op = (OperationDefinition) def;
                for (Selection<?> selection : op.getSelectionSet().getSelections()) {
                    max = Math.max(max, calculateDepth(selection, 1));
                }
            }
        }
        return max;
    }

    private int calculateDepth(Selection<?> selection, int currentDepth) {
        if (selection instanceof Field) {
            Field field = (Field) selection;
            SelectionSet selectionSet = field.getSelectionSet();
            if (selectionSet == null) {
                return currentDepth;
            }
            int max = currentDepth;
            for (Selection<?> subSelection : selectionSet.getSelections()) {
                max = Math.max(max, calculateDepth(subSelection, currentDepth + 1));
            }
            return max;
        } else if (selection instanceof InlineFragment) {
            InlineFragment fragment = (InlineFragment) selection;
            int max = currentDepth;
            for (Selection<?> subSelection : fragment.getSelectionSet().getSelections()) {
                max = Math.max(max, calculateDepth(subSelection, currentDepth));
            }
            return max;
        } else if (selection instanceof FragmentSpread) {
            return currentDepth;
        }
        return currentDepth;
    }

    private int countFields(Document document) {
        int count = 0;
        for (Definition<?> def : document.getDefinitions()) {
            if (def instanceof OperationDefinition) {
                OperationDefinition op = (OperationDefinition) def;
                count += countSelections(op.getSelectionSet());
            }
        }
        return count;
    }

    private int countSelections(SelectionSet selectionSet) {
        if (selectionSet == null) return 0;
        int count = 0;
        for (Selection<?> selection : selectionSet.getSelections()) {
            if (selection instanceof Field) {
                Field field = (Field) selection;
                count++;
                count += countSelections(field.getSelectionSet());
            } else if (selection instanceof InlineFragment) {
                InlineFragment fragment = (InlineFragment) selection;
                count += countSelections(fragment.getSelectionSet());
            }
        }
        return count;
    }

    private List<QueryPartition> partitionQuery(Document document) {
        List<QueryPartition> partitions = new ArrayList<>();

        for (Definition<?> def : document.getDefinitions()) {
            if (def instanceof OperationDefinition) {
                OperationDefinition op = (OperationDefinition) def;
                boolean isMutation = op.getOperation() == OperationDefinition.Operation.MUTATION;

                Map<String, List<Field>> fieldMap = new LinkedHashMap<>();

                for (Selection<?> selection : op.getSelectionSet().getSelections()) {
                    if (selection instanceof Field) {
                        Field field = (Field) selection;
                        String fieldName = field.getName();
                        fieldMap.computeIfAbsent(fieldName, k -> new ArrayList<>()).add(field);
                    }
                }

                partitions.add(new QueryPartition(isMutation, fieldMap, extractVariables(op)));
            }
        }

        return partitions;
    }

    private List<VariableDefinition> extractVariables(OperationDefinition op) {
        return op.getVariableDefinitions() != null ? op.getVariableDefinitions() : Collections.emptyList();
    }

    public static class ParseResult {
        private final boolean success;
        private final String errorMessage;
        private final List<QueryPartition> partitions;
        private final int depth;
        private final int fieldCount;

        private ParseResult(boolean success, String errorMessage, List<QueryPartition> partitions, int depth, int fieldCount) {
            this.success = success;
            this.errorMessage = errorMessage;
            this.partitions = partitions;
            this.depth = depth;
            this.fieldCount = fieldCount;
        }

        public static ParseResult success(List<QueryPartition> partitions, int depth, int fieldCount) {
            return new ParseResult(true, null, partitions, depth, fieldCount);
        }

        public static ParseResult error(String errorMessage) {
            return new ParseResult(false, errorMessage, null, 0, 0);
        }

        public boolean isSuccess() {
            return success;
        }

        public String getErrorMessage() {
            return errorMessage;
        }

        public List<QueryPartition> getPartitions() {
            return partitions;
        }

        public int getDepth() {
            return depth;
        }

        public int getFieldCount() {
            return fieldCount;
        }
    }

    public static class QueryPartition {
        private final boolean isMutation;
        private final Map<String, List<Field>> fieldsByRootField;
        private final List<VariableDefinition> variables;

        public QueryPartition(boolean isMutation, Map<String, List<Field>> fieldsByRootField,
                              List<VariableDefinition> variables) {
            this.isMutation = isMutation;
            this.fieldsByRootField = fieldsByRootField;
            this.variables = variables;
        }

        public boolean isMutation() {
            return isMutation;
        }

        public Map<String, List<Field>> getFieldsByRootField() {
            return fieldsByRootField;
        }

        public List<VariableDefinition> getVariables() {
            return variables;
        }
    }
}

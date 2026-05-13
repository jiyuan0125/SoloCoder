package com.graphql.gateway.server;

import com.fasterxml.jackson.databind.JsonNode;
import com.fasterxml.jackson.databind.ObjectMapper;
import com.fasterxml.jackson.databind.node.ArrayNode;
import com.fasterxml.jackson.databind.node.ObjectNode;
import com.graphql.gateway.executor.QueryExecutor;
import com.graphql.gateway.model.QueryStats;
import com.graphql.gateway.model.SchemaRegistration;
import com.graphql.gateway.parser.QueryParser;
import com.graphql.gateway.schema.SchemaManager;
import io.netty.buffer.Unpooled;
import io.netty.channel.ChannelFutureListener;
import io.netty.channel.ChannelHandler;
import io.netty.channel.ChannelHandlerContext;
import io.netty.channel.SimpleChannelInboundHandler;
import io.netty.handler.codec.http.*;
import io.netty.util.CharsetUtil;

import java.util.Map;

import static io.netty.handler.codec.http.HttpHeaderNames.*;
import static io.netty.handler.codec.http.HttpHeaderValues.*;
import static io.netty.handler.codec.http.HttpMethod.*;
import static io.netty.handler.codec.http.HttpResponseStatus.*;

@ChannelHandler.Sharable
public class HttpServerHandler extends SimpleChannelInboundHandler<FullHttpRequest> {
    private final ObjectMapper objectMapper = new ObjectMapper();
    private final SchemaManager schemaManager = SchemaManager.getInstance();
    private final QueryParser queryParser = QueryParser.getInstance();
    private final QueryExecutor queryExecutor = QueryExecutor.getInstance();

    @Override
    protected void channelRead0(ChannelHandlerContext ctx, FullHttpRequest request) throws Exception {
        if (!request.decoderResult().isSuccess()) {
            sendError(ctx, BAD_REQUEST);
            return;
        }

        String uri = request.uri();
        HttpMethod method = request.method();

        if (uri.equals("/graphql") && method.equals(POST)) {
            handleGraphQL(ctx, request);
        } else if (uri.equals("/schemas") && method.equals(POST)) {
            handleRegisterSchema(ctx, request);
        } else if (uri.equals("/schema") && method.equals(GET)) {
            handleGetSchema(ctx);
        } else if (uri.equals("/stats") && method.equals(GET)) {
            handleGetStats(ctx);
        } else if (uri.equals("/health") && method.equals(GET)) {
            handleHealth(ctx);
        } else {
            sendError(ctx, NOT_FOUND);
        }
    }

    private void handleGraphQL(ChannelHandlerContext ctx, FullHttpRequest request) {
        try {
            String body = request.content().toString(CharsetUtil.UTF_8);
            JsonNode json = objectMapper.readTree(body);

            String query = json.path("query").asText(null);
            if (query == null) {
                sendJsonError(ctx, BAD_REQUEST, "Query is required");
                return;
            }

            QueryParser.ParseResult parseResult = queryParser.parse(query);
            if (!parseResult.isSuccess()) {
                sendGraphQLError(ctx, BAD_REQUEST, parseResult.getErrorMessage());
                return;
            }

            JsonNode variables = json.path("variables");
            if (variables.isMissingNode() || variables.isNull()) {
                variables = null;
            }

            QueryExecutor.ExecutionResult result = queryExecutor.execute(parseResult, variables);

            ObjectNode response = objectMapper.createObjectNode();
            if (result.getData() != null) {
                response.set("data", result.getData());
            }
            if (result.hasErrors()) {
                ArrayNode errors = objectMapper.createArrayNode();
                for (String error : result.getErrors()) {
                    ObjectNode errorNode = objectMapper.createObjectNode();
                    errorNode.put("message", error);
                    errors.add(errorNode);
                }
                response.set("errors", errors);
            }

            sendJson(ctx, OK, response);

        } catch (Exception e) {
            sendGraphQLError(ctx, INTERNAL_SERVER_ERROR, "Internal error: " + e.getMessage());
        }
    }

    private void handleRegisterSchema(ChannelHandlerContext ctx, FullHttpRequest request) {
        try {
            String body = request.content().toString(CharsetUtil.UTF_8);
            SchemaRegistration registration = objectMapper.readValue(body, SchemaRegistration.class);

            SchemaManager.SchemaRegistrationResult result = schemaManager.registerSchema(registration);

            if (result.isSuccess()) {
                ObjectNode response = objectMapper.createObjectNode();
                response.put("success", true);
                response.put("message", "Schema registered successfully");
                sendJson(ctx, OK, response);
            } else if (result.isConflict()) {
                ObjectNode response = objectMapper.createObjectNode();
                response.put("success", false);
                response.put("error", result.getErrorMessage());
                sendJson(ctx, CONFLICT, response);
            } else {
                ObjectNode response = objectMapper.createObjectNode();
                response.put("success", false);
                response.put("error", result.getErrorMessage());
                sendJson(ctx, BAD_REQUEST, response);
            }

        } catch (Exception e) {
            sendJsonError(ctx, BAD_REQUEST, "Invalid request: " + e.getMessage());
        }
    }

    private void handleGetSchema(ChannelHandlerContext ctx) {
        String schema = schemaManager.getMergedSchema();
        if (schema == null || schema.isEmpty()) {
            schema = "# No schemas registered yet. Use POST /schemas to register.\n";
        }
        sendText(ctx, OK, schema);
    }

    private void handleGetStats(ChannelHandlerContext ctx) {
        QueryStats stats = queryExecutor.getStats();

        ObjectNode response = objectMapper.createObjectNode();
        response.put("totalQueries", stats.getTotalQueries());
        response.put("averageParseTimeMillis", stats.getAverageParseTimeMillis());
        response.put("averageExecutionTimeMillis", stats.getAverageExecutionTimeMillis());

        ObjectNode backendStats = objectMapper.createObjectNode();
        for (Map.Entry<String, QueryStats.BackendCallStats> entry : stats.getBackendStats().entrySet()) {
            ObjectNode backendNode = objectMapper.createObjectNode();
            backendNode.put("callCount", entry.getValue().getCallCount());
            backendNode.put("successCount", entry.getValue().getSuccessCount());
            backendNode.put("averageTimeMillis", entry.getValue().getAverageTimeMillis());
            backendStats.set(entry.getKey(), backendNode);
        }
        response.set("backendStats", backendStats);

        ObjectNode registeredServices = objectMapper.createObjectNode();
        for (com.graphql.gateway.model.BackendService backend : schemaManager.getAllBackends()) {
            ObjectNode serviceNode = objectMapper.createObjectNode();
            serviceNode.put("endpoint", backend.getEndpoint());
            ArrayNode queryFields = objectMapper.createArrayNode();
            backend.getRootQueryFields().forEach(queryFields::add);
            serviceNode.set("queryFields", queryFields);
            ArrayNode mutationFields = objectMapper.createArrayNode();
            backend.getRootMutationFields().forEach(mutationFields::add);
            serviceNode.set("mutationFields", mutationFields);
            registeredServices.set(backend.getName(), serviceNode);
        }
        response.set("registeredServices", registeredServices);

        sendJson(ctx, OK, response);
    }

    private void handleHealth(ChannelHandlerContext ctx) {
        ObjectNode response = objectMapper.createObjectNode();
        response.put("status", "healthy");
        sendJson(ctx, OK, response);
    }

    private void sendJson(ChannelHandlerContext ctx, HttpResponseStatus status, ObjectNode content) {
        try {
            String jsonStr = objectMapper.writeValueAsString(content);
            FullHttpResponse response = new DefaultFullHttpResponse(
                    HttpVersion.HTTP_1_1, status, Unpooled.copiedBuffer(jsonStr, CharsetUtil.UTF_8));
            response.headers().set(CONTENT_TYPE, APPLICATION_JSON + "; charset=UTF-8");
            response.headers().set(CONTENT_LENGTH, response.content().readableBytes());
            response.headers().set(CONNECTION, HttpHeaderValues.KEEP_ALIVE);
            ctx.writeAndFlush(response).addListener(ChannelFutureListener.CLOSE_ON_FAILURE);
        } catch (Exception e) {
            sendError(ctx, INTERNAL_SERVER_ERROR);
        }
    }

    private void sendText(ChannelHandlerContext ctx, HttpResponseStatus status, String content) {
        FullHttpResponse response = new DefaultFullHttpResponse(
                HttpVersion.HTTP_1_1, status, Unpooled.copiedBuffer(content, CharsetUtil.UTF_8));
        response.headers().set(CONTENT_TYPE, "text/plain; charset=UTF-8");
        response.headers().set(CONTENT_LENGTH, response.content().readableBytes());
        response.headers().set(CONNECTION, HttpHeaderValues.KEEP_ALIVE);
        ctx.writeAndFlush(response).addListener(ChannelFutureListener.CLOSE_ON_FAILURE);
    }

    private void sendJsonError(ChannelHandlerContext ctx, HttpResponseStatus status, String message) {
        ObjectNode error = objectMapper.createObjectNode();
        error.put("error", message);
        sendJson(ctx, status, error);
    }

    private void sendGraphQLError(ChannelHandlerContext ctx, HttpResponseStatus status, String message) {
        ObjectNode response = objectMapper.createObjectNode();
        ArrayNode errors = objectMapper.createArrayNode();
        ObjectNode errorNode = objectMapper.createObjectNode();
        errorNode.put("message", message);
        errors.add(errorNode);
        response.set("errors", errors);
        sendJson(ctx, status, response);
    }

    private void sendError(ChannelHandlerContext ctx, HttpResponseStatus status) {
        FullHttpResponse response = new DefaultFullHttpResponse(
                HttpVersion.HTTP_1_1, status,
                Unpooled.copiedBuffer("Failure: " + status + "\r\n", CharsetUtil.UTF_8));
        response.headers().set(CONTENT_TYPE, "text/plain; charset=UTF-8");
        ctx.writeAndFlush(response).addListener(ChannelFutureListener.CLOSE);
    }

    @Override
    public void exceptionCaught(ChannelHandlerContext ctx, Throwable cause) {
        cause.printStackTrace();
        ctx.close();
    }
}

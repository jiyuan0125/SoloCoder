package com.graphql.gateway;

public class GatewayServer {
    public static void main(String[] args) throws InterruptedException {
        com.graphql.gateway.server.GatewayServer server = new com.graphql.gateway.server.GatewayServer();
        server.start();
    }
}

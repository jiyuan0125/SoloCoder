package com.gateway.model;

public class Version {
    private String name;
    private String host;
    private int port;
    private int weight;

    public Version() {
    }

    public Version(String name, String host, int port, int weight) {
        this.name = name;
        this.host = host;
        this.port = port;
        this.weight = weight;
    }

    public String getName() {
        return name;
    }

    public void setName(String name) {
        this.name = name;
    }

    public String getHost() {
        return host;
    }

    public void setHost(String host) {
        this.host = host;
    }

    public int getPort() {
        return port;
    }

    public void setPort(int port) {
        this.port = port;
    }

    public int getWeight() {
        return weight;
    }

    public void setWeight(int weight) {
        this.weight = weight;
    }
}

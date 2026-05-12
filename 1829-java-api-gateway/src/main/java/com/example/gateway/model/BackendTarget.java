package com.example.gateway.model;

public class BackendTarget {
    private String url;
    private int weight;

    public BackendTarget() {}

    public BackendTarget(String url, int weight) {
        this.url = url;
        this.weight = weight;
    }

    public String getUrl() { return url; }
    public void setUrl(String url) { this.url = url; }

    public int getWeight() { return weight; }
    public void setWeight(int weight) { this.weight = weight; }
}

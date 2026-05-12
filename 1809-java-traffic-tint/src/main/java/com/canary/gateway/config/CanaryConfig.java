package com.canary.gateway.config;

import org.springframework.boot.context.properties.ConfigurationProperties;
import org.springframework.stereotype.Component;

@Component
@ConfigurationProperties(prefix = "canary")
public class CanaryConfig {
    private int ratio;
    private String grayBackend;
    private String prodBackend;

    public int getRatio() {
        return ratio;
    }

    public void setRatio(int ratio) {
        this.ratio = ratio;
    }

    public String getGrayBackend() {
        return grayBackend;
    }

    public void setGrayBackend(String grayBackend) {
        this.grayBackend = grayBackend;
    }

    public String getProdBackend() {
        return prodBackend;
    }

    public void setProdBackend(String prodBackend) {
        this.prodBackend = prodBackend;
    }
}

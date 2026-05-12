package com.logcollector.config;

import org.springframework.boot.context.properties.ConfigurationProperties;
import org.springframework.stereotype.Component;

@Component
@ConfigurationProperties(prefix = "log-collector")
public class LogCollectorProperties {

    private String logDirectory = "./logs";
    private int loadDays = 7;
    private String defaultLevel = "INFO";
    private String filePrefix = "access";

    public String getLogDirectory() {
        return logDirectory;
    }

    public void setLogDirectory(String logDirectory) {
        this.logDirectory = logDirectory;
    }

    public int getLoadDays() {
        return loadDays;
    }

    public void setLoadDays(int loadDays) {
        this.loadDays = loadDays;
    }

    public String getDefaultLevel() {
        return defaultLevel;
    }

    public void setDefaultLevel(String defaultLevel) {
        this.defaultLevel = defaultLevel;
    }

    public String getFilePrefix() {
        return filePrefix;
    }

    public void setFilePrefix(String filePrefix) {
        this.filePrefix = filePrefix;
    }
}

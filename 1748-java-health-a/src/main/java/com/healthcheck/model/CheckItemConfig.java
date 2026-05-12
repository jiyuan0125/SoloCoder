package com.healthcheck.model;

import lombok.Data;
import lombok.NoArgsConstructor;
import lombok.AllArgsConstructor;

@Data
@NoArgsConstructor
@AllArgsConstructor
public class CheckItemConfig {
    private String name;
    private CheckType type;
    private String target;
    private Integer timeout = 5000;
    private Integer retries = 1;
    private String scriptPath;
    private String method = "GET";
    private Integer expectedStatusCode = 200;
    private String expectedContent;
    private String host;
    private Integer port;
}

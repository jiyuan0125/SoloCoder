package com.example.apiproxymirror.model;

import lombok.Data;

import java.util.List;

@Data
public class RecordingSettings {
    private boolean enabled;
    private List<String> includePaths;
    private List<String> excludePaths;
}

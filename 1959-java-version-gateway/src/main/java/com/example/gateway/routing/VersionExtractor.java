package com.example.gateway.routing;

import java.util.regex.Matcher;
import java.util.regex.Pattern;

public class VersionExtractor {
    private static final Pattern VERSION_PATTERN = Pattern.compile("^/(v\\d+)(/.*)?$");

    private VersionExtractor() {
    }

    public static VersionPath extract(String path, String defaultVersion) {
        if (path == null) {
            return new VersionPath(defaultVersion, "/");
        }
        Matcher matcher = VERSION_PATTERN.matcher(path);
        if (matcher.matches()) {
            String version = matcher.group(1);
            String rest = matcher.group(2);
            if (rest == null || rest.isEmpty()) {
                rest = "/";
            }
            return new VersionPath(version, rest);
        }
        return new VersionPath(defaultVersion, path);
    }

    public static class VersionPath {
        private final String version;
        private final String remainingPath;

        public VersionPath(String version, String remainingPath) {
            this.version = version;
            this.remainingPath = remainingPath;
        }

        public String getVersion() {
            return version;
        }

        public String getRemainingPath() {
            return remainingPath;
        }
    }
}

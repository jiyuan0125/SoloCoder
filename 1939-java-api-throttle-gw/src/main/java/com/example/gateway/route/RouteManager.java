package com.example.gateway.route;

import com.example.gateway.model.RouteDefinition;
import com.example.gateway.model.RouteMatchResult;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import java.net.URI;
import java.net.URISyntaxException;
import java.util.ArrayList;
import java.util.Collections;
import java.util.HashMap;
import java.util.LinkedHashMap;
import java.util.List;
import java.util.Map;
import java.util.concurrent.ConcurrentHashMap;
import java.util.regex.Matcher;
import java.util.regex.Pattern;

public class RouteManager {
    private static final Logger logger = LoggerFactory.getLogger(RouteManager.class);
    private final Map<String, RouteDefinition> routes = new ConcurrentHashMap<>();
    private final Map<String, Pattern> compiledPatterns = new ConcurrentHashMap<>();

    private static final Pattern PARAM_PATTERN = Pattern.compile("\\{(\\w+)}");

    public void addRoute(String path, String targetUrl) {
        RouteDefinition route = new RouteDefinition();
        route.setPath(path);
        route.setTargetUrl(targetUrl);

        List<String> paramNames = extractParamNames(path);
        route.setPathParamNames(paramNames);

        try {
            URI uri = new URI(targetUrl);
            Map<String, String> hostInfo = new HashMap<>();
            hostInfo.put("host", uri.getHost());
            hostInfo.put("port", String.valueOf(uri.getPort() > 0 ? uri.getPort() : (uri.getScheme().equals("https") ? 443 : 80)));
            hostInfo.put("scheme", uri.getScheme());
            route.setTargetHost(hostInfo);
            route.setTargetPath(uri.getPath().isEmpty() ? "/" : uri.getPath());
        } catch (URISyntaxException e) {
            logger.error("Invalid target URL: {}", targetUrl, e);
            throw new IllegalArgumentException("Invalid target URL: " + targetUrl, e);
        }

        routes.put(path, route);
        compiledPatterns.put(path, compilePattern(path));
        logger.info("Route added: {} -> {}", path, targetUrl);
    }

    public void removeRoute(String path) {
        routes.remove(path);
        compiledPatterns.remove(path);
        logger.info("Route removed: {}", path);
    }

    public Map<String, RouteDefinition> getAllRoutes() {
        return Collections.unmodifiableMap(new LinkedHashMap<>(routes));
    }

    public RouteMatchResult matchRoute(String requestPath) {
        String pathWithoutQuery = requestPath.split("\\?")[0];

        for (Map.Entry<String, RouteDefinition> entry : routes.entrySet()) {
            Pattern pattern = compiledPatterns.get(entry.getKey());
            Matcher matcher = pattern.matcher(pathWithoutQuery);

            if (matcher.matches()) {
                RouteDefinition route = entry.getValue();
                Map<String, String> pathParams = new HashMap<>();

                List<String> paramNames = route.getPathParamNames();
                for (int i = 0; i < paramNames.size(); i++) {
                    pathParams.put(paramNames.get(i), matcher.group(i + 1));
                }

                String targetPath = buildTargetPath(route.getTargetPath(), pathParams);

                logger.debug("Route matched: {} -> {} (params: {})", requestPath, targetPath, pathParams);
                return new RouteMatchResult(true, route, pathParams, targetPath);
            }
        }

        logger.debug("No route matched for: {}", requestPath);
        return new RouteMatchResult(false, null, Collections.emptyMap(), null);
    }

    private Pattern compilePattern(String path) {
        String regexPath = path.replaceAll("\\{(\\w+)}", "([^/]+)");
        return Pattern.compile("^" + regexPath + "$");
    }

    private List<String> extractParamNames(String path) {
        List<String> names = new ArrayList<>();
        Matcher matcher = PARAM_PATTERN.matcher(path);
        while (matcher.find()) {
            names.add(matcher.group(1));
        }
        return names;
    }

    private String buildTargetPath(String targetPathTemplate, Map<String, String> pathParams) {
        String result = targetPathTemplate;
        for (Map.Entry<String, String> entry : pathParams.entrySet()) {
            result = result.replace("{" + entry.getKey() + "}", entry.getValue());
        }
        return result;
    }
}

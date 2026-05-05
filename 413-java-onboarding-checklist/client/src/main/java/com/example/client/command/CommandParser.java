package com.example.client.command;

import java.time.LocalDate;
import java.time.format.DateTimeFormatter;
import java.util.HashMap;
import java.util.Map;

public class CommandParser {
    private static final DateTimeFormatter DATE_FORMATTER = DateTimeFormatter.ISO_LOCAL_DATE;

    public Command parse(String[] args) {
        if (args.length == 0) {
            return new Command("help", null);
        }

        String action = args[0].toLowerCase();
        Map<String, String> parameters = new HashMap<>();

        for (int i = 1; i < args.length; i++) {
            if (args[i].startsWith("--")) {
                String key = args[i].substring(2);
                if (i + 1 < args.length && !args[i + 1].startsWith("--")) {
                    parameters.put(key, args[i + 1]);
                    i++;
                } else {
                    parameters.put(key, "true");
                }
            }
        }

        return new Command(action, parameters);
    }

    public static class Command {
        private final String action;
        private final Map<String, String> parameters;

        public Command(String action, Map<String, String> parameters) {
            this.action = action;
            this.parameters = parameters != null ? parameters : new HashMap<>();
        }

        public String getAction() {
            return action;
        }

        public String getParameter(String key) {
            return parameters.get(key);
        }

        public String getParameter(String key, String defaultValue) {
            return parameters.getOrDefault(key, defaultValue);
        }

        public int getParameterAsInt(String key, int defaultValue) {
            String value = parameters.get(key);
            if (value != null) {
                try {
                    return Integer.parseInt(value);
                } catch (NumberFormatException e) {
                    return defaultValue;
                }
            }
            return defaultValue;
        }

        public LocalDate getParameterAsDate(String key) {
            String value = parameters.get(key);
            if (value != null) {
                try {
                    return LocalDate.parse(value, DATE_FORMATTER);
                } catch (Exception e) {
                    return null;
                }
            }
            return null;
        }

        public boolean hasParameter(String key) {
            return parameters.containsKey(key);
        }

        public boolean getParameterAsBoolean(String key) {
            String value = parameters.get(key);
            return "true".equalsIgnoreCase(value) || "yes".equalsIgnoreCase(value) || "1".equals(value);
        }
    }
}

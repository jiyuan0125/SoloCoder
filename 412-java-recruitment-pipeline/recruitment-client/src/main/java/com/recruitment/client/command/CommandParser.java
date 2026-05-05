package com.recruitment.client.command;

import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;
import java.util.Map;

public class CommandParser {
    private final String[] args;
    private String command;
    private final Map<String, String> options;
    private final List<String> positionalArgs;

    public CommandParser(String[] args) {
        this.args = args;
        this.options = new HashMap<>();
        this.positionalArgs = new ArrayList<>();
        parse();
    }

    private void parse() {
        if (args.length == 0) {
            return;
        }

        command = args[0];

        for (int i = 1; i < args.length; i++) {
            String arg = args[i];
            if (arg.startsWith("--")) {
                String[] parts = arg.substring(2).split("=", 2);
                if (parts.length == 2) {
                    options.put(parts[0], parts[1]);
                } else {
                    if (i + 1 < args.length && !args[i + 1].startsWith("--")) {
                        options.put(parts[0], args[i + 1]);
                        i++;
                    } else {
                        options.put(parts[0], "true");
                    }
                }
            } else if (arg.startsWith("-")) {
                String key = arg.substring(1);
                if (i + 1 < args.length && !args[i + 1].startsWith("-")) {
                    options.put(key, args[i + 1]);
                    i++;
                } else {
                    options.put(key, "true");
                }
            } else {
                positionalArgs.add(arg);
            }
        }
    }

    public String getCommand() {
        return command;
    }

    public String getOption(String key) {
        return options.get(key);
    }

    public String getOption(String key, String defaultValue) {
        return options.getOrDefault(key, defaultValue);
    }

    public Integer getIntOption(String key, Integer defaultValue) {
        String value = options.get(key);
        if (value == null) {
            return defaultValue;
        }
        try {
            return Integer.parseInt(value);
        } catch (NumberFormatException e) {
            return defaultValue;
        }
    }

    public List<String> getPositionalArgs() {
        return positionalArgs;
    }

    public boolean hasOption(String key) {
        return options.containsKey(key);
    }
}

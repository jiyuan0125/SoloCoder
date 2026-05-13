package com.configcenter.config;

public class AppConfig {

    private static final int DEFAULT_PORT = 9106;
    private static final String DEFAULT_ENCRYPTION_KEY = "0123456789abcdef0123456789abcdef";

    private final int port;
    private final String encryptionKey;

    public AppConfig(int port, String encryptionKey) {
        this.port = port;
        this.encryptionKey = encryptionKey;
    }

    public int getPort() {
        return port;
    }

    public String getEncryptionKey() {
        return encryptionKey;
    }

    public static AppConfig parse(String[] args) {
        int port = DEFAULT_PORT;
        String encryptionKey = DEFAULT_ENCRYPTION_KEY;

        String envPort = System.getenv("PORT");
        if (envPort != null && !envPort.isEmpty()) {
            try {
                port = Integer.parseInt(envPort);
            } catch (NumberFormatException e) {
                System.err.println("Invalid PORT environment variable, using default: " + DEFAULT_PORT);
            }
        }

        for (int i = 0; i < args.length; i++) {
            switch (args[i]) {
                case "--port":
                    if (i + 1 < args.length) {
                        try {
                            port = Integer.parseInt(args[i + 1]);
                        } catch (NumberFormatException e) {
                            System.err.println("Invalid port, using default: " + DEFAULT_PORT);
                        }
                        i++;
                    }
                    break;
                case "--encryption-key":
                    if (i + 1 < args.length) {
                        encryptionKey = args[i + 1];
                        i++;
                    }
                    break;
                case "-p":
                    if (i + 1 < args.length) {
                        try {
                            port = Integer.parseInt(args[i + 1]);
                        } catch (NumberFormatException e) {
                            System.err.println("Invalid port, using default: " + DEFAULT_PORT);
                        }
                        i++;
                    }
                    break;
            }
        }

        return new AppConfig(port, encryptionKey);
    }
}

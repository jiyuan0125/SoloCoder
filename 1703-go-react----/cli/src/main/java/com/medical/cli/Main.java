package com.medical.cli;

import com.google.gson.Gson;
import com.google.gson.JsonArray;
import com.google.gson.JsonElement;
import com.google.gson.JsonObject;
import org.apache.commons.cli.*;

import java.io.*;
import java.net.HttpURLConnection;
import java.net.URL;
import java.nio.charset.StandardCharsets;

public class Main {
    private static final String DEFAULT_BASE_URL = "http://localhost:8080";
    private static final Gson gson = new Gson();

    public static void main(String[] args) {
        Options options = new Options();

        Option baseUrlOpt = Option.builder("u")
                .longOpt("url")
                .hasArg()
                .argName("URL")
                .desc("API base URL (default: http://localhost:8080)")
                .build();
        options.addOption(baseUrlOpt);

        Option startDateOpt = Option.builder("s")
                .longOpt("start-date")
                .hasArg()
                .argName("DATE")
                .desc("Start date (YYYY-MM-DD)")
                .build();
        options.addOption(startDateOpt);

        Option endDateOpt = Option.builder("e")
                .longOpt("end-date")
                .hasArg()
                .argName("DATE")
                .desc("End date (YYYY-MM-DD)")
                .build();
        options.addOption(endDateOpt);

        Option outputOpt = Option.builder("o")
                .longOpt("output")
                .hasArg()
                .argName("FILE")
                .desc("Output file path")
                .build();
        options.addOption(outputOpt);

        Option helpOpt = Option.builder("h")
                .longOpt("help")
                .desc("Show help")
                .build();
        options.addOption(helpOpt);

        CommandLineParser parser = new DefaultParser();
        CommandLine cmd;

        try {
            cmd = parser.parse(options, args);
        } catch (ParseException e) {
            System.err.println("Error parsing command line: " + e.getMessage());
            printUsage(options);
            System.exit(1);
            return;
        }

        if (cmd.hasOption("help") || args.length == 0) {
            printUsage(options);
            return;
        }

        String baseUrl = cmd.getOptionValue("url", DEFAULT_BASE_URL);

        String[] remainingArgs = cmd.getArgs();
        if (remainingArgs.length == 0) {
            System.err.println("Error: Command is required. Available commands: export-records");
            printUsage(options);
            System.exit(1);
            return;
        }

        String command = remainingArgs[0];

        switch (command) {
            case "export-records":
                handleExportRecords(cmd, baseUrl, options);
                break;
            default:
                System.err.println("Unknown command: " + command);
                System.err.println("Available commands: export-records");
                System.exit(1);
        }
    }

    private static void handleExportRecords(CommandLine cmd, String baseUrl, Options options) {
        String startDate = cmd.getOptionValue("start-date");
        String endDate = cmd.getOptionValue("end-date");
        String output = cmd.getOptionValue("output");

        if (startDate == null || endDate == null) {
            System.err.println("Error: --start-date and --end-date are required for export-records");
            printUsage(options);
            System.exit(1);
            return;
        }

        try {
            String urlStr = baseUrl + "/api/export-records?start_date=" + startDate + "&end_date=" + endDate;
            URL url = new URL(urlStr);
            HttpURLConnection conn = (HttpURLConnection) url.openConnection();
            conn.setRequestMethod("GET");
            conn.setRequestProperty("Accept", "text/csv");

            int responseCode = conn.getResponseCode();
            if (responseCode != HttpURLConnection.HTTP_OK) {
                System.err.println("Error: Server returned " + responseCode);
                try (InputStream errorStream = conn.getErrorStream()) {
                    if (errorStream != null) {
                        String error = new String(errorStream.readAllBytes(), StandardCharsets.UTF_8);
                        System.err.println("Response: " + error);
                    }
                }
                System.exit(1);
                return;
            }

            InputStream is = conn.getInputStream();
            OutputStream os;

            if (output != null) {
                os = new FileOutputStream(output);
                System.out.println("Downloading CSV to: " + output);
            } else {
                String fileName = "medical_records_" + startDate + "_" + endDate + ".csv";
                os = new FileOutputStream(fileName);
                System.out.println("Downloading CSV to: " + fileName);
            }

            byte[] buffer = new byte[8192];
            int bytesRead;
            long totalBytes = 0;
            while ((bytesRead = is.read(buffer)) != -1) {
                os.write(buffer, 0, bytesRead);
                totalBytes += bytesRead;
            }

            is.close();
            os.close();
            conn.disconnect();

            System.out.println("Download completed. Total bytes: " + totalBytes);

        } catch (Exception e) {
            System.err.println("Error during export: " + e.getMessage());
            e.printStackTrace();
            System.exit(1);
        }
    }

    private static void printUsage(Options options) {
        HelpFormatter formatter = new HelpFormatter();
        System.out.println("Medical Exam System CLI");
        System.out.println();
        System.out.println("Usage: java -jar medical-cli.jar [OPTIONS] COMMAND");
        System.out.println();
        System.out.println("Commands:");
        System.out.println("  export-records  Export medical records by date range");
        System.out.println();
        System.out.println("Options:");
        formatter.printOptions(new PrintWriter(System.out), 80, options, 2, 2);
        System.out.println();
        System.out.println("Examples:");
        System.out.println("  java -jar medical-cli.jar export-records -s 2026-01-01 -e 2026-01-31");
        System.out.println("  java -jar medical-cli.jar export-records -s 2026-01-01 -e 2026-01-31 -o records.csv");
        System.out.println("  java -jar medical-cli.jar -u http://server:8080 export-records -s 2026-01-01 -e 2026-01-31");
    }
}

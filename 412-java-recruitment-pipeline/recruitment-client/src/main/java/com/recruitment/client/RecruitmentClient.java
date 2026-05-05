package com.recruitment.client;

import com.recruitment.client.command.CommandExecutor;
import com.recruitment.client.command.CommandParser;

public class RecruitmentClient {
    public static void main(String[] args) {
        CommandExecutor executor = new CommandExecutor();
        try {
            CommandParser parser = new CommandParser(args);
            executor.execute(parser);
        } catch (Exception e) {
            System.err.println("执行出错: " + e.getMessage());
            e.printStackTrace();
        } finally {
            try {
                executor.close();
            } catch (Exception e) {
                // ignore
            }
        }
    }
}

package com.orgchart.client;

import com.orgchart.client.command.Command;
import com.orgchart.client.command.CommandRegistry;
import com.orgchart.client.http.HttpClient;

import java.util.Arrays;

public class OrgChartClient {

    public static void main(String[] args) {
        CommandRegistry registry = new CommandRegistry();
        HttpClient httpClient = new HttpClient();

        if (args.length == 0) {
            printWelcome();
            try {
                Command helpCommand = registry.getCommand("help");
                if (helpCommand != null) {
                    helpCommand.execute(new String[0], httpClient);
                }
            } catch (Exception e) {
                e.printStackTrace();
            }
            return;
        }

        String commandName = args[0];
        String[] remainingArgs = args.length > 1 ? Arrays.copyOfRange(args, 1, args.length) : new String[0];

        Command command = registry.getCommand(commandName);
        if (command == null) {
            System.out.println("未知命令: " + commandName);
            System.out.println("使用 'help' 查看可用命令");
            System.exit(1);
        }

        try {
            command.execute(remainingArgs, httpClient);
        } catch (Exception e) {
            System.err.println("执行命令时出错: " + e.getMessage());
            e.printStackTrace();
            System.exit(1);
        }
    }

    private static void printWelcome() {
        System.out.println();
        System.out.println("╔══════════════════════════════════════════════════════════════╗");
        System.out.println("║         组织架构命令行工具 (Org Chart CLI)                   ║");
        System.out.println("║                    版本 1.0.0                              ║");
        System.out.println("╚══════════════════════════════════════════════════════════════╝");
        System.out.println();
    }
}

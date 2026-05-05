package com.orgchart.client.command;

import com.orgchart.client.http.HttpClient;

import java.util.List;

public class HelpCommand implements Command {

    private final CommandRegistry registry;

    public HelpCommand(CommandRegistry registry) {
        this.registry = registry;
    }

    @Override
    public String getName() {
        return "help";
    }

    @Override
    public String getDescription() {
        return "显示帮助信息";
    }

    @Override
    public void execute(String[] args, HttpClient httpClient) {
        System.out.println();
        System.out.println("╔════════════════════════════════════════════════════════════╗");
        System.out.println("║            组织架构命令行工具 (Org Chart CLI)              ║");
        System.out.println("╚════════════════════════════════════════════════════════════╝");
        System.out.println();
        System.out.println("使用方法: org-chart-cli <command> [options]");
        System.out.println();
        System.out.println("可用命令:");
        System.out.println();

        List<Command> commands = registry.getAllCommands();
        int maxNameLen = commands.stream()
                .map(Command::getName)
                .mapToInt(String::length)
                .max()
                .orElse(0);

        for (Command cmd : commands) {
            String name = cmd.getName();
            String description = cmd.getDescription();
            System.out.println("  " + name + " ".repeat(maxNameLen - name.length() + 4) + description);
        }

        System.out.println();
        System.out.println("示例:");
        System.out.println("  org-chart-cli department list");
        System.out.println("  org-chart-cli employee list");
        System.out.println("  org-chart-cli help department");
        System.out.println();
    }
}

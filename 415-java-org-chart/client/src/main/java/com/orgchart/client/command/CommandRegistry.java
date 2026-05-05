package com.orgchart.client.command;

import java.util.ArrayList;
import java.util.LinkedHashMap;
import java.util.List;
import java.util.Map;

public class CommandRegistry {

    private final Map<String, Command> commands = new LinkedHashMap<>();

    public CommandRegistry() {
        registerCommand(new HelpCommand(this));
        registerCommand(new DepartmentCommands());
        registerCommand(new EmployeeCommands());
        registerCommand(new VirtualTeamCommands());
        registerCommand(new LogCommands());
    }

    public void registerCommand(Command command) {
        commands.put(command.getName(), command);
    }

    public Command getCommand(String name) {
        return commands.get(name);
    }

    public List<Command> getAllCommands() {
        return new ArrayList<>(commands.values());
    }

    public List<String> getCommandNames() {
        return new ArrayList<>(commands.keySet());
    }
}

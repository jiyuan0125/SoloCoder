package com.workshift.client.command;

import com.workshift.client.http.HttpClient;

public class CommandDispatcher {

    private final HttpClient httpClient;
    private final EmployeeCommand employeeCommand;
    private final ShiftCommand shiftCommand;
    private final SwapCommand swapCommand;
    private final ObjectionCommand objectionCommand;
    private final StatisticsCommand statisticsCommand;

    public CommandDispatcher(String baseUrl) {
        this.httpClient = new HttpClient(baseUrl);
        this.employeeCommand = new EmployeeCommand(httpClient);
        this.shiftCommand = new ShiftCommand(httpClient);
        this.swapCommand = new SwapCommand(httpClient);
        this.objectionCommand = new ObjectionCommand(httpClient);
        this.statisticsCommand = new StatisticsCommand(httpClient);
    }

    public void dispatch(String[] args) throws Exception {
        String category = args[0];

        switch (category.toLowerCase()) {
            case "employee":
                employeeCommand.execute(args);
                break;
            case "shift":
                shiftCommand.execute(args);
                break;
            case "swap":
                swapCommand.execute(args);
                break;
            case "objection":
                objectionCommand.execute(args);
                break;
            case "stats":
                statisticsCommand.execute(args);
                break;
            default:
                System.out.println("未知命令类别: " + category);
                System.out.println("可用类别: employee, shift, swap, objection, stats");
        }
    }
}
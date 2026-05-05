package com.payroll.client;

import com.payroll.client.command.CommandExecutor;

public class PayrollArchiveClient {

    public static void main(String[] args) {
        CommandExecutor executor = new CommandExecutor();
        executor.execute(args);
    }
}

package com.orgchart.client.command;

import com.orgchart.client.http.HttpClient;

public interface Command {

    String getName();
    
    String getDescription();
    
    void execute(String[] args, HttpClient httpClient) throws Exception;
}

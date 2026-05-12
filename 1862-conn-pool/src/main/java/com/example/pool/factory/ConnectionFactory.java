package com.example.pool.factory;

import com.example.pool.model.DataSourceType;

import java.util.Map;

public interface ConnectionFactory {
    
    DataSourceType getType();
    
    Object createConnection(Map<String, String> params) throws Exception;
    
    void closeConnection(Object connection);
    
    boolean validateConnection(Object connection);
}

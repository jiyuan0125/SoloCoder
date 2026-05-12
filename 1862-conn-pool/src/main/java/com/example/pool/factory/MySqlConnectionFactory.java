package com.example.pool.factory;

import com.example.pool.model.DataSourceType;
import lombok.extern.slf4j.Slf4j;
import org.springframework.stereotype.Component;

import java.sql.Connection;
import java.sql.DriverManager;
import java.util.Map;

@Slf4j
@Component
public class MySqlConnectionFactory implements ConnectionFactory {

    @Override
    public DataSourceType getType() {
        return DataSourceType.MYSQL;
    }

    @Override
    public Object createConnection(Map<String, String> params) throws Exception {
        String host = params.getOrDefault("host", "localhost");
        int port = Integer.parseInt(params.getOrDefault("port", "3306"));
        String database = params.getOrDefault("database", "");
        String username = params.getOrDefault("username", "root");
        String password = params.getOrDefault("password", "");

        String url = String.format("jdbc:mysql://%s:%d/%s?useSSL=false&allowPublicKeyRetrieval=true&serverTimezone=UTC",
            host, port, database);

        return DriverManager.getConnection(url, username, password);
    }

    @Override
    public void closeConnection(Object connection) {
        if (connection instanceof Connection conn) {
            try {
                conn.close();
            } catch (Exception e) {
                log.warn("Failed to close MySQL connection: {}", e.getMessage());
            }
        }
    }

    @Override
    public boolean validateConnection(Object connection) {
        if (connection instanceof Connection conn) {
            try {
                return conn.isValid(5);
            } catch (Exception e) {
                log.warn("Failed to validate MySQL connection: {}", e.getMessage());
                return false;
            }
        }
        return false;
    }
}

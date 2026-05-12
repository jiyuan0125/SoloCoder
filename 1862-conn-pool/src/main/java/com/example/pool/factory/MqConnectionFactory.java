package com.example.pool.factory;

import com.example.pool.model.DataSourceType;
import jakarta.jms.Connection;
import lombok.extern.slf4j.Slf4j;
import org.apache.activemq.ActiveMQConnection;
import org.apache.activemq.ActiveMQConnectionFactory;
import org.springframework.stereotype.Component;

import java.util.Map;

@Slf4j
@Component
public class MqConnectionFactory implements ConnectionFactory {

    @Override
    public DataSourceType getType() {
        return DataSourceType.MQ;
    }

    @Override
    public Object createConnection(Map<String, String> params) throws Exception {
        String host = params.getOrDefault("host", "localhost");
        int port = Integer.parseInt(params.getOrDefault("port", "61616"));
        String username = params.getOrDefault("username", ActiveMQConnection.DEFAULT_USER);
        String password = params.getOrDefault("password", ActiveMQConnection.DEFAULT_PASSWORD);

        String brokerUrl = String.format("tcp://%s:%d", host, port);
        ActiveMQConnectionFactory factory = new ActiveMQConnectionFactory(brokerUrl);
        Connection connection = factory.createConnection(username, password);
        connection.start();
        return connection;
    }

    @Override
    public void closeConnection(Object connection) {
        if (connection instanceof Connection conn) {
            try {
                conn.close();
            } catch (Exception e) {
                log.warn("Failed to close MQ connection: {}", e.getMessage());
            }
        }
    }

    @Override
    public boolean validateConnection(Object connection) {
        if (connection instanceof ActiveMQConnection conn) {
            try {
                return !conn.isClosed();
            } catch (Exception e) {
                log.warn("Failed to validate MQ connection: {}", e.getMessage());
                return false;
            }
        }
        return false;
    }
}

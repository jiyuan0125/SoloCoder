package com.example.pool.factory;

import com.example.pool.model.DataSourceType;
import lombok.extern.slf4j.Slf4j;
import org.springframework.stereotype.Component;
import redis.clients.jedis.Jedis;

import java.util.Map;

@Slf4j
@Component
public class RedisConnectionFactory implements ConnectionFactory {

    @Override
    public DataSourceType getType() {
        return DataSourceType.REDIS;
    }

    @Override
    public Object createConnection(Map<String, String> params) throws Exception {
        String host = params.getOrDefault("host", "localhost");
        int port = Integer.parseInt(params.getOrDefault("port", "6379"));
        int timeout = Integer.parseInt(params.getOrDefault("timeout", "5000"));
        String password = params.get("password");

        Jedis jedis = new Jedis(host, port, timeout);
        if (password != null && !password.isEmpty()) {
            jedis.auth(password);
        }
        jedis.connect();
        return jedis;
    }

    @Override
    public void closeConnection(Object connection) {
        if (connection instanceof Jedis jedis) {
            try {
                jedis.close();
            } catch (Exception e) {
                log.warn("Failed to close Redis connection: {}", e.getMessage());
            }
        }
    }

    @Override
    public boolean validateConnection(Object connection) {
        if (connection instanceof Jedis jedis) {
            try {
                return "PONG".equals(jedis.ping());
            } catch (Exception e) {
                log.warn("Failed to validate Redis connection: {}", e.getMessage());
                return false;
            }
        }
        return false;
    }
}

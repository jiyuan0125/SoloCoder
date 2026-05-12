package com.poolmgr.pool;

import com.poolmgr.model.DataSourceType;
import com.poolmgr.model.PoolConfig;

import java.sql.Connection;
import java.sql.DriverManager;
import java.sql.SQLException;
import redis.clients.jedis.Jedis;

public class ConnectionFactory {

    public static ConnectionWrapper createConnection(PoolConfig config) throws Exception {
        if (config.getType() == DataSourceType.MYSQL) {
            return createMysqlConnection(config);
        } else if (config.getType() == DataSourceType.REDIS) {
            return createRedisConnection(config);
        }
        throw new IllegalArgumentException("Unsupported data source type: " + config.getType());
    }

    private static ConnectionWrapper createMysqlConnection(PoolConfig config) throws SQLException {
        StringBuilder url = new StringBuilder("jdbc:mysql://");
        url.append(config.getHost()).append(":").append(config.getPort());
        if (config.getDatabase() != null && !config.getDatabase().isEmpty()) {
            url.append("/").append(config.getDatabase());
        }
        url.append("?useSSL=false&allowPublicKeyRetrieval=true&serverTimezone=UTC");
        
        Connection connection = DriverManager.getConnection(
            url.toString(),
            config.getUsername(),
            config.getPassword()
        );
        return new ConnectionWrapper(DataSourceType.MYSQL, connection);
    }

    private static ConnectionWrapper createRedisConnection(PoolConfig config) {
        Jedis jedis = new Jedis(config.getHost(), config.getPort());
        if (config.getPassword() != null && !config.getPassword().isEmpty()) {
            jedis.auth(config.getPassword());
        }
        return new ConnectionWrapper(DataSourceType.REDIS, jedis);
    }
}

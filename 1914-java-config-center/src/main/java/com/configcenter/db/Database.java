package com.configcenter.db;

import com.configcenter.model.ConfigEntry;
import com.configcenter.model.Watch;
import org.slf4j.Logger;
import org.slf4j.LoggerFactory;

import java.sql.*;
import java.time.Instant;
import java.util.ArrayList;
import java.util.List;
import java.util.Optional;

public class Database {

    private static final Logger logger = LoggerFactory.getLogger(Database.class);
    private static final String DB_URL = "jdbc:sqlite:config-center.db";

    public void init() {
        try (Connection conn = DriverManager.getConnection(DB_URL);
             Statement stmt = conn.createStatement()) {

            stmt.execute("CREATE TABLE IF NOT EXISTS configs (" +
                    "id INTEGER PRIMARY KEY AUTOINCREMENT," +
                    "config_key TEXT UNIQUE NOT NULL," +
                    "value TEXT," +
                    "is_secret BOOLEAN DEFAULT 0," +
                    "created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP," +
                    "updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP" +
                    ")");

            stmt.execute("CREATE TABLE IF NOT EXISTS watches (" +
                    "id INTEGER PRIMARY KEY AUTOINCREMENT," +
                    "client_id TEXT UNIQUE NOT NULL," +
                    "mode TEXT NOT NULL," +
                    "callback_url TEXT," +
                    "created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP" +
                    ")");

            stmt.execute("CREATE TABLE IF NOT EXISTS changes (" +
                    "id INTEGER PRIMARY KEY AUTOINCREMENT," +
                    "config_key TEXT NOT NULL," +
                    "action TEXT NOT NULL," +
                    "old_value TEXT," +
                    "new_value TEXT," +
                    "is_secret BOOLEAN DEFAULT 0," +
                    "timestamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP" +
                    ")");

            stmt.execute("CREATE TABLE IF NOT EXISTS client_changes (" +
                    "id INTEGER PRIMARY KEY AUTOINCREMENT," +
                    "client_id TEXT NOT NULL," +
                    "change_id INTEGER NOT NULL," +
                    "FOREIGN KEY (change_id) REFERENCES changes(id)" +
                    ")");

            logger.info("Database initialized successfully");
        } catch (SQLException e) {
            throw new RuntimeException("Failed to initialize database", e);
        }
    }

    public Optional<ConfigEntry> getConfig(String key) {
        String sql = "SELECT * FROM configs WHERE config_key = ?";
        try (Connection conn = DriverManager.getConnection(DB_URL);
             PreparedStatement pstmt = conn.prepareStatement(sql)) {
            pstmt.setString(1, key);
            ResultSet rs = pstmt.executeQuery();
            if (rs.next()) {
                return Optional.of(mapRowToConfig(rs));
            }
            return Optional.empty();
        } catch (SQLException e) {
            throw new RuntimeException("Failed to get config", e);
        }
    }

    public List<ConfigEntry> getAllConfigs() {
        String sql = "SELECT * FROM configs";
        List<ConfigEntry> configs = new ArrayList<>();
        try (Connection conn = DriverManager.getConnection(DB_URL);
             Statement stmt = conn.createStatement();
             ResultSet rs = stmt.executeQuery(sql)) {
            while (rs.next()) {
                configs.add(mapRowToConfig(rs));
            }
            return configs;
        } catch (SQLException e) {
            throw new RuntimeException("Failed to get all configs", e);
        }
    }

    public void saveConfig(ConfigEntry config) {
        String sql = "INSERT INTO configs (config_key, value, is_secret, updated_at) VALUES (?, ?, ?, ?) " +
                "ON CONFLICT(config_key) DO UPDATE SET value = ?, is_secret = ?, updated_at = ?";
        try (Connection conn = DriverManager.getConnection(DB_URL);
             PreparedStatement pstmt = conn.prepareStatement(sql)) {
            Timestamp now = Timestamp.from(Instant.now());
            pstmt.setString(1, config.getKey());
            pstmt.setString(2, config.getValue());
            pstmt.setBoolean(3, config.isSecret());
            pstmt.setTimestamp(4, now);
            pstmt.setString(5, config.getValue());
            pstmt.setBoolean(6, config.isSecret());
            pstmt.setTimestamp(7, now);
            pstmt.executeUpdate();
        } catch (SQLException e) {
            throw new RuntimeException("Failed to save config", e);
        }
    }

    public boolean deleteConfig(String key) {
        String sql = "DELETE FROM configs WHERE config_key = ?";
        try (Connection conn = DriverManager.getConnection(DB_URL);
             PreparedStatement pstmt = conn.prepareStatement(sql)) {
            pstmt.setString(1, key);
            return pstmt.executeUpdate() > 0;
        } catch (SQLException e) {
            throw new RuntimeException("Failed to delete config", e);
        }
    }

    public void saveWatch(Watch watch) {
        String sql = "INSERT INTO watches (client_id, mode, callback_url) VALUES (?, ?, ?) " +
                "ON CONFLICT(client_id) DO UPDATE SET mode = ?, callback_url = ?";
        try (Connection conn = DriverManager.getConnection(DB_URL);
             PreparedStatement pstmt = conn.prepareStatement(sql)) {
            pstmt.setString(1, watch.getClientId());
            pstmt.setString(2, watch.getMode());
            pstmt.setString(3, watch.getCallbackUrl());
            pstmt.setString(4, watch.getMode());
            pstmt.setString(5, watch.getCallbackUrl());
            pstmt.executeUpdate();
        } catch (SQLException e) {
            throw new RuntimeException("Failed to save watch", e);
        }
    }

    public Optional<Watch> getWatch(String clientId) {
        String sql = "SELECT * FROM watches WHERE client_id = ?";
        try (Connection conn = DriverManager.getConnection(DB_URL);
             PreparedStatement pstmt = conn.prepareStatement(sql)) {
            pstmt.setString(1, clientId);
            ResultSet rs = pstmt.executeQuery();
            if (rs.next()) {
                return Optional.of(new Watch(
                        rs.getString("client_id"),
                        rs.getString("mode"),
                        rs.getString("callback_url")
                ));
            }
            return Optional.empty();
        } catch (SQLException e) {
            throw new RuntimeException("Failed to get watch", e);
        }
    }

    public List<Watch> getAllWatches() {
        String sql = "SELECT * FROM watches";
        List<Watch> watches = new ArrayList<>();
        try (Connection conn = DriverManager.getConnection(DB_URL);
             Statement stmt = conn.createStatement();
             ResultSet rs = stmt.executeQuery(sql)) {
            while (rs.next()) {
                watches.add(new Watch(
                        rs.getString("client_id"),
                        rs.getString("mode"),
                        rs.getString("callback_url")
                ));
            }
            return watches;
        } catch (SQLException e) {
            throw new RuntimeException("Failed to get all watches", e);
        }
    }

    public void logChange(String key, String action, String oldValue, String newValue, boolean isSecret) {
        String sql = "INSERT INTO changes (config_key, action, old_value, new_value, is_secret) VALUES (?, ?, ?, ?, ?)";
        try (Connection conn = DriverManager.getConnection(DB_URL);
             PreparedStatement pstmt = conn.prepareStatement(sql, Statement.RETURN_GENERATED_KEYS)) {
            pstmt.setString(1, key);
            pstmt.setString(2, action);
            pstmt.setString(3, oldValue);
            pstmt.setString(4, newValue);
            pstmt.setBoolean(5, isSecret);
            pstmt.executeUpdate();

            try (ResultSet rs = pstmt.getGeneratedKeys()) {
                if (rs.next()) {
                    int changeId = rs.getInt(1);
                    recordPendingChange(changeId);
                }
            }
        } catch (SQLException e) {
            throw new RuntimeException("Failed to log change", e);
        }
    }

    private void recordPendingChange(int changeId) {
        List<Watch> watches = getAllWatches();
        String sql = "INSERT INTO client_changes (client_id, change_id) VALUES (?, ?)";
        try (Connection conn = DriverManager.getConnection(DB_URL);
             PreparedStatement pstmt = conn.prepareStatement(sql)) {
            for (Watch watch : watches) {
                pstmt.setString(1, watch.getClientId());
                pstmt.setInt(2, changeId);
                pstmt.addBatch();
            }
            pstmt.executeBatch();
        } catch (SQLException e) {
            throw new RuntimeException("Failed to record pending change", e);
        }
    }

    public List<ChangeRecord> getPendingChanges(String clientId) {
        String sql = "SELECT c.* FROM changes c " +
                "INNER JOIN client_changes cc ON c.id = cc.change_id " +
                "WHERE cc.client_id = ? ORDER BY c.id ASC";
        List<ChangeRecord> changes = new ArrayList<>();
        try (Connection conn = DriverManager.getConnection(DB_URL);
             PreparedStatement pstmt = conn.prepareStatement(sql)) {
            pstmt.setString(1, clientId);
            ResultSet rs = pstmt.executeQuery();
            while (rs.next()) {
                changes.add(new ChangeRecord(
                        rs.getInt("id"),
                        rs.getString("config_key"),
                        rs.getString("action"),
                        rs.getString("old_value"),
                        rs.getString("new_value"),
                        rs.getBoolean("is_secret"),
                        rs.getTimestamp("timestamp").toInstant()
                ));
            }
            return changes;
        } catch (SQLException e) {
            throw new RuntimeException("Failed to get pending changes", e);
        }
    }

    public void clearPendingChanges(String clientId) {
        String sql = "DELETE FROM client_changes WHERE client_id = ?";
        try (Connection conn = DriverManager.getConnection(DB_URL);
             PreparedStatement pstmt = conn.prepareStatement(sql)) {
            pstmt.setString(1, clientId);
            pstmt.executeUpdate();
        } catch (SQLException e) {
            throw new RuntimeException("Failed to clear pending changes", e);
        }
    }

    private ConfigEntry mapRowToConfig(ResultSet rs) throws SQLException {
        return new ConfigEntry(
                rs.getString("config_key"),
                rs.getString("value"),
                rs.getBoolean("is_secret"),
                rs.getTimestamp("created_at").toInstant(),
                rs.getTimestamp("updated_at").toInstant()
        );
    }
}

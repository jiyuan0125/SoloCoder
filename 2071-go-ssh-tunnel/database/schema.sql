CREATE TABLE IF NOT EXISTS tunnel_configs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    type TEXT NOT NULL,
    ssh_server TEXT NOT NULL,
    ssh_user TEXT NOT NULL,
    auth_type TEXT NOT NULL,
    auth_data TEXT NOT NULL,
    local_port INTEGER NOT NULL,
    remote_host TEXT NOT NULL,
    remote_port INTEGER NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS tunnel_states (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    tunnel_id INTEGER NOT NULL UNIQUE,
    status TEXT NOT NULL,
    bytes_up INTEGER DEFAULT 0,
    bytes_down INTEGER DEFAULT 0,
    reconnect_count INTEGER DEFAULT 0,
    last_connected_at DATETIME,
    last_disconnected_at DATETIME,
    last_disconnect_reason TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (tunnel_id) REFERENCES tunnel_configs(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS connection_logs (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    tunnel_id INTEGER NOT NULL,
    connected_at DATETIME NOT NULL,
    disconnected_at DATETIME,
    disconnect_reason TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (tunnel_id) REFERENCES tunnel_configs(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS tunnel_records (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    tunnel_config_id INTEGER NOT NULL,
    status TEXT NOT NULL,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (tunnel_config_id) REFERENCES tunnel_configs(id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS operation_histories (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    record_id INTEGER NOT NULL,
    op_type TEXT NOT NULL,
    operator TEXT NOT NULL,
    description TEXT,
    note TEXT,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    FOREIGN KEY (record_id) REFERENCES tunnel_records(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_tunnel_states_tunnel_id ON tunnel_states(tunnel_id);
CREATE INDEX IF NOT EXISTS idx_connection_logs_tunnel_id ON connection_logs(tunnel_id);
CREATE INDEX IF NOT EXISTS idx_operation_histories_record_id ON operation_histories(record_id);

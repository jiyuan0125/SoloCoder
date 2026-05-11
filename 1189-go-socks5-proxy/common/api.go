package common

type StartProxyRequest struct {
	ListenPort  int    `json:"listen_port"`
	RequireAuth bool   `json:"require_auth"`
	Users       []User `json:"users,omitempty"`
}

type User struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type StartProxyResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Address string `json:"address,omitempty"`
}

type ProxyStatusResponse struct {
	Running bool   `json:"running"`
	Address string `json:"address,omitempty"`
}

type ConnectionInfo struct {
	ClientAddr string `json:"client_addr"`
	TargetAddr string `json:"target_addr"`
	StartTime  string `json:"start_time"`
}

type UDPAssociationInfo struct {
	ClientAddr string `json:"client_addr"`
	RelayAddr  string `json:"relay_addr"`
	StartTime  string `json:"start_time"`
}

type ActiveConnectionsResponse struct {
	TCPConnections  []ConnectionInfo     `json:"tcp_connections"`
	UDPAssociations []UDPAssociationInfo `json:"udp_associations"`
}

type AddUserRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

type RemoveUserRequest struct {
	Username string `json:"username"`
}

type UserResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type ListUsersResponse struct {
	Users []string `json:"users"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

type TestProxyRequest struct {
	ProxyHost string `json:"proxy_host"`
	ProxyPort int    `json:"proxy_port"`
	Username  string `json:"username,omitempty"`
	Password  string `json:"password,omitempty"`
	TargetURL string `json:"target_url"`
}

type TestProxyResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Status  int    `json:"status,omitempty"`
}

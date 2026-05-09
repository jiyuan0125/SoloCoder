package common

import "stun-tool/pkg/stun"

type DetectRequest struct {
	ServerHost string `json:"server_host,omitempty"`
	ServerPort int    `json:"server_port,omitempty"`
	Username   string `json:"username,omitempty"`
	Password   string `json:"password,omitempty"`
}

type DetectResponse struct {
	Success      bool           `json:"success"`
	NATType      string         `json:"nat_type"`
	LocalIP      string         `json:"local_ip"`
	PublicIP     string         `json:"public_ip"`
	PublicPort   uint16         `json:"public_port"`
	Error        string         `json:"error,omitempty"`
}

type BindingRequest struct {
	ServerHost string `json:"server_host"`
	ServerPort int    `json:"server_port"`
	Username   string `json:"username,omitempty"`
	Password   string `json:"password,omitempty"`
}

type BindingResponse struct {
	Success    bool   `json:"success"`
	IP         string `json:"ip,omitempty"`
	Port       uint16 `json:"port,omitempty"`
	Family     byte   `json:"family,omitempty"`
	Error      string `json:"error,omitempty"`
}

func (r *DetectResponse) FromNATResult(result *stun.NATResult, err error) {
	if err != nil {
		r.Success = false
		r.Error = err.Error()
		if result != nil {
			r.NATType = string(result.NATType)
			r.LocalIP = result.LocalIP
			r.PublicIP = result.PublicIP
			if result.MappingAddress != nil {
				r.PublicPort = result.MappingAddress.Port
			}
		}
		return
	}

	r.Success = true
	r.NATType = string(result.NATType)
	r.LocalIP = result.LocalIP
	r.PublicIP = result.PublicIP
	if result.MappingAddress != nil {
		r.PublicPort = result.MappingAddress.Port
	}
}

func (r *BindingResponse) FromMappedAddress(addr *stun.MappedAddress, err error) {
	if err != nil {
		r.Success = false
		r.Error = err.Error()
		return
	}

	r.Success = true
	r.IP = addr.IP.String()
	r.Port = addr.Port
	r.Family = addr.Family
}

func DefaultServer() stun.StunServer {
	if len(stun.DefaultServers) > 0 {
		return stun.DefaultServers[0]
	}
	return stun.StunServer{Host: "stun.l.google.com", Port: 19302}
}

func (req *DetectRequest) GetServer() stun.StunServer {
	server := DefaultServer()
	if req.ServerHost != "" {
		server.Host = req.ServerHost
	}
	if req.ServerPort > 0 {
		server.Port = req.ServerPort
	}
	return server
}

func (req *BindingRequest) GetServer() stun.StunServer {
	server := DefaultServer()
	if req.ServerHost != "" {
		server.Host = req.ServerHost
	}
	if req.ServerPort > 0 {
		server.Port = req.ServerPort
	}
	return server
}

func (req *DetectRequest) GetCredentials() *stun.Credentials {
	if req.Username == "" && req.Password == "" {
		return nil
	}
	return &stun.Credentials{
		Username: req.Username,
		Password: req.Password,
	}
}

func (req *BindingRequest) GetCredentials() *stun.Credentials {
	if req.Username == "" && req.Password == "" {
		return nil
	}
	return &stun.Credentials{
		Username: req.Username,
		Password: req.Password,
	}
}

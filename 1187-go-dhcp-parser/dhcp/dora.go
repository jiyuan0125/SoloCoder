package dhcp

import (
	"fmt"
	"net"
)

type DORAResult struct {
	Success    bool
	Error      string
	Discover   *Packet
	Offer      *Packet
	Request    *Packet
	Ack        *Packet
	AssignedIP net.IP
	Lease      *Lease
}

type Server struct {
	pool *IPPool
}

func NewServer(pool *IPPool) *Server {
	return &Server{pool: pool}
}

func (s *Server) ProcessDiscover(discover *Packet) (*Packet, error) {
	if discover.Op != OpBootRequest {
		return nil, fmt.Errorf("not a BOOTREQUEST")
	}
	msgType, err := discover.MessageType()
	if err != nil {
		return nil, fmt.Errorf("failed to get message type: %w", err)
	}
	if msgType != MsgDiscover {
		return nil, fmt.Errorf("not a DISCOVER message, got: %s", MessageTypeToString(msgType))
	}
	if len(discover.Chaddr) == 0 {
		return nil, fmt.Errorf("no CHADDR in discover packet")
	}
	offeredIP, err := s.pool.Allocate(discover.Chaddr)
	if err != nil {
		return nil, fmt.Errorf("failed to allocate IP: %w", err)
	}
	offer := BuildOffer(
		discover,
		offeredIP,
		s.pool.ServerIP(),
		s.pool.LeaseTime(),
		s.pool.SubnetMask(),
		s.pool.Routers(),
		s.pool.DNSServers(),
	)
	return offer, nil
}

func (s *Server) ProcessRequest(request *Packet) (*Packet, *Lease, error) {
	if request.Op != OpBootRequest {
		return nil, nil, fmt.Errorf("not a BOOTREQUEST")
	}
	msgType, err := request.MessageType()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get message type: %w", err)
	}
	if msgType != MsgRequest {
		return nil, nil, fmt.Errorf("not a REQUEST message, got: %s", MessageTypeToString(msgType))
	}
	if len(request.Chaddr) == 0 {
		return nil, nil, fmt.Errorf("no CHADDR in request packet")
	}
	var requestedIP net.IP
	hasServerID := false
	serverID, err := request.Options.ServerIdentifier()
	if err == nil {
		hasServerID = true
	}
	if !request.Ciaddr.Equal(net.IPv4zero) {
		requestedIP = request.Ciaddr
	} else {
		requestedIP, _ = request.Options.RequestedIP()
	}
	if hasServerID {
		if !serverID.Equal(s.pool.ServerIP()) {
			return nil, nil, fmt.Errorf("request is for another server: %s", serverID)
		}
	}
	var assignedIP net.IP
	var lease *Lease
	if !request.Ciaddr.Equal(net.IPv4zero) {
		existingLease := s.pool.GetLeaseByMAC(request.Chaddr)
		if existingLease == nil {
			nak := BuildNak(request, s.pool.ServerIP())
			return nak, nil, fmt.Errorf("no active lease for renewal")
		}
		if !existingLease.IP.Equal(request.Ciaddr) {
			nak := BuildNak(request, s.pool.ServerIP())
			return nak, nil, fmt.Errorf("CIADDR does not match existing lease")
		}
		assignedIP = request.Ciaddr
		lease, err = s.pool.RenewLease(request.Chaddr)
		if err != nil {
			nak := BuildNak(request, s.pool.ServerIP())
			return nak, nil, fmt.Errorf("failed to renew lease: %w", err)
		}
	} else {
		if requestedIP == nil {
			nak := BuildNak(request, s.pool.ServerIP())
			return nak, nil, fmt.Errorf("no requested IP")
		}
		lease = s.pool.GrantLease(request.Chaddr, requestedIP)
		if lease == nil {
			nak := BuildNak(request, s.pool.ServerIP())
			return nak, nil, fmt.Errorf("failed to grant lease")
		}
		assignedIP = requestedIP
	}
	ack := BuildAck(
		request,
		assignedIP,
		s.pool.ServerIP(),
		s.pool.LeaseTime(),
		s.pool.SubnetMask(),
		s.pool.Routers(),
		s.pool.DNSServers(),
	)
	return ack, lease, nil
}

func (s *Server) ProcessRelease(release *Packet) error {
	if release.Op != OpBootRequest {
		return fmt.Errorf("not a BOOTREQUEST")
	}
	msgType, err := release.MessageType()
	if err != nil {
		return fmt.Errorf("failed to get message type: %w", err)
	}
	if msgType != MsgRelease {
		return fmt.Errorf("not a RELEASE message, got: %s", MessageTypeToString(msgType))
	}
	if len(release.Chaddr) == 0 {
		return fmt.Errorf("no CHADDR in release packet")
	}
	return s.pool.Release(release.Chaddr)
}

func (s *Server) SimulateDORA(mac net.HardwareAddr) (*DORAResult, error) {
	result := &DORAResult{Success: false}
	discover := BuildDiscover(mac)
	result.Discover = discover
	offer, err := s.ProcessDiscover(discover)
	if err != nil {
		result.Error = fmt.Sprintf("ProcessDiscover failed: %v", err)
		return result, err
	}
	result.Offer = offer
	offeredIP := offer.Yiaddr
	serverID, _ := offer.Options.ServerIdentifier()
	request := BuildRequestFromDiscover(discover, offeredIP, serverID)
	result.Request = request
	ack, lease, err := s.ProcessRequest(request)
	if err != nil {
		result.Error = fmt.Sprintf("ProcessRequest failed: %v", err)
		return result, err
	}
	result.Ack = ack
	msgType, _ := ack.MessageType()
	if msgType == MsgNak {
		result.Error = "Received NAK"
		return result, fmt.Errorf("received NAK")
	}
	result.Lease = lease
	result.AssignedIP = lease.IP
	result.Success = true
	return result, nil
}

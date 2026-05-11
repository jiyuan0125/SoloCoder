package main

import (
	"bytes"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/mqtt/parser/api"
	"github.com/mqtt/parser/mqtt"
)

func (s *Server) handleParse(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		s.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req api.ParseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	var data []byte
	var err error
	switch req.Format {
	case "hex":
		data, err = hex.DecodeString(req.Data)
	case "base64":
		data, err = base64.StdEncoding.DecodeString(req.Data)
	default:
		s.writeError(w, http.StatusBadRequest, "Invalid format, must be 'hex' or 'base64'")
		return
	}
	if err != nil {
		s.writeError(w, http.StatusBadRequest, fmt.Sprintf("Invalid data: %v", err))
		return
	}

	packet, err := mqtt.Parse(data)
	if err != nil {
		s.writeError(w, http.StatusBadRequest, fmt.Sprintf("Parse error: %v", err))
		return
	}

	info, err := s.packetToInfo(packet, data)
	if err != nil {
		s.writeError(w, http.StatusInternalServerError, fmt.Sprintf("Process error: %v", err))
		return
	}

	resp := api.ParseResponse{
		Success: true,
		Packet:  *info,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (s *Server) handleBuild(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		s.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req api.BuildRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	packet, err := s.buildPacket(req)
	if err != nil {
		s.writeError(w, http.StatusBadRequest, fmt.Sprintf("Build error: %v", err))
		return
	}

	encoded, err := packet.Encode()
	if err != nil {
		s.writeError(w, http.StatusBadRequest, fmt.Sprintf("Encode error: %v", err))
		return
	}

	resp := api.BuildResponse{
		Success: true,
		Hex:     hex.EncodeToString(encoded),
		Base64:  base64.StdEncoding.EncodeToString(encoded),
		Length:  len(encoded),
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (s *Server) handleWill(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		s.handleWillGet(w, r)
	case http.MethodPut:
		s.handleWillPut(w, r)
	default:
		s.writeError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

func (s *Server) handleWillGet(w http.ResponseWriter, r *http.Request) {
	resp := api.WillConfigResponse{
		Success: true,
		Config:  *s.willConfig,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (s *Server) handleWillPut(w http.ResponseWriter, r *http.Request) {
	var cfg api.WillConfig
	if err := json.NewDecoder(r.Body).Decode(&cfg); err != nil {
		s.writeError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	if cfg.QoS < 0 || cfg.QoS > 2 {
		s.writeError(w, http.StatusBadRequest, "Invalid QoS, must be 0, 1, or 2")
		return
	}

	if cfg.Enabled {
		if cfg.Topic == "" {
			s.writeError(w, http.StatusBadRequest, "Will topic is required when enabled")
			return
		}
	}

	s.willConfig = &cfg

	resp := api.WillConfigResponse{
		Success: true,
		Config:  cfg,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (s *Server) writeError(w http.ResponseWriter, status int, message string) {
	w.WriteHeader(status)
	resp := api.ParseResponse{
		Success: false,
		Error:   message,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (s *Server) packetToInfo(packet mqtt.Packet, rawData []byte) (*api.PacketInfo, error) {
	fh, err := mqtt.GetFixedHeader(rawData)
	if err != nil {
		return nil, err
	}

	info := &api.PacketInfo{
		Type:            packet.Type().String(),
		RemainingLength: fh.RemainingLength,
		Flags: api.FlagsInfo{
			Raw: packet.Flags(),
			QoS: int(fh.QoS),
		},
	}

	switch pkt := packet.(type) {
	case *mqtt.ConnectPacket:
		info.Flags.Dup = fh.Dup
		info.Flags.Retain = fh.Retain
		info.Details = map[string]interface{}{
			"protocol_name":    pkt.ProtocolName,
			"protocol_level":   pkt.ProtocolLevel,
			"clean_session":    pkt.CleanSession,
			"will_flag":        pkt.WillFlag,
			"will_qos":         pkt.WillQoS,
			"will_retain":      pkt.WillRetain,
			"username_flag":    pkt.UsernameFlag,
			"password_flag":    pkt.PasswordFlag,
			"keep_alive":       pkt.KeepAlive,
			"client_id":        pkt.ClientID,
		}
		if pkt.WillFlag {
			info.Details["will_topic"] = pkt.WillTopic
			info.Details["will_message"] = pkt.WillMessage
		}
		if pkt.UsernameFlag {
			info.Details["username"] = pkt.Username
		}
		if pkt.PasswordFlag {
			info.Details["password"] = pkt.Password
		}

	case *mqtt.ConnackPacket:
		info.Details = map[string]interface{}{
			"session_present": pkt.SessionPresent,
			"return_code":     pkt.ReturnCode,
			"return_message":  pkt.ReturnCode.String(),
		}

	case *mqtt.PublishPacket:
		info.Flags.Dup = pkt.Dup
		info.Flags.Retain = pkt.Retain
		info.Details = map[string]interface{}{
			"topic":    pkt.Topic,
			"packet_id": pkt.PacketID,
			"payload":  pkt.Payload,
		}

	case *mqtt.PubackPacket:
		info.Details = map[string]interface{}{
			"packet_id": pkt.PacketID,
		}

	case *mqtt.PubrecPacket:
		info.Details = map[string]interface{}{
			"packet_id": pkt.PacketID,
		}

	case *mqtt.PubrelPacket:
		info.Details = map[string]interface{}{
			"packet_id": pkt.PacketID,
		}

	case *mqtt.PubcompPacket:
		info.Details = map[string]interface{}{
			"packet_id": pkt.PacketID,
		}

	case *mqtt.SubscribePacket:
		subs := []map[string]interface{}{}
		for _, sub := range pkt.Subscriptions {
			subs = append(subs, map[string]interface{}{
				"topic_filter": sub.TopicFilter,
				"requested_qos": sub.RequestedQoS,
			})
		}
		info.Details = map[string]interface{}{
			"packet_id":     pkt.PacketID,
			"subscriptions": subs,
		}

	case *mqtt.SubackPacket:
		returnCodes := []map[string]interface{}{}
		for _, rc := range pkt.ReturnCodes {
			returnCodes = append(returnCodes, map[string]interface{}{
				"code": rc,
				"description": rc.String(),
			})
		}
		info.Details = map[string]interface{}{
			"packet_id":    pkt.PacketID,
			"return_codes": returnCodes,
		}

	case *mqtt.UnsubscribePacket:
		info.Details = map[string]interface{}{
			"packet_id":      pkt.PacketID,
			"topic_filters":  pkt.TopicFilters,
		}

	case *mqtt.UnsubackPacket:
		info.Details = map[string]interface{}{
			"packet_id": pkt.PacketID,
		}

	case *mqtt.PingreqPacket, *mqtt.PingrespPacket, *mqtt.DisconnectPacket:
		info.Details = map[string]interface{}{}
	}

	return info, nil
}

func (s *Server) buildPacket(req api.BuildRequest) (mqtt.Packet, error) {
	switch req.PacketType {
	case "CONNECT", "connect":
		return s.buildConnect(req.Fields)
	case "CONNACK", "connack":
		return s.buildConnack(req.Fields)
	case "PUBLISH", "publish":
		return s.buildPublish(req.Fields)
	case "PUBACK", "puback":
		return s.buildPuback(req.Fields)
	case "PUBREC", "pubrec":
		return s.buildPubrec(req.Fields)
	case "PUBREL", "pubrel":
		return s.buildPubrel(req.Fields)
	case "PUBCOMP", "pubcomp":
		return s.buildPubcomp(req.Fields)
	case "SUBSCRIBE", "subscribe":
		return s.buildSubscribe(req.Fields)
	case "SUBACK", "suback":
		return s.buildSuback(req.Fields)
	case "UNSUBSCRIBE", "unsubscribe":
		return s.buildUnsubscribe(req.Fields)
	case "UNSUBACK", "unsuback":
		return s.buildUnsuback(req.Fields)
	case "PINGREQ", "pingreq":
		return &mqtt.PingreqPacket{}, nil
	case "PINGRESP", "pingresp":
		return &mqtt.PingrespPacket{}, nil
	case "DISCONNECT", "disconnect":
		return &mqtt.DisconnectPacket{}, nil
	default:
		return nil, fmt.Errorf("unknown packet type: %s", req.PacketType)
	}
}

func getInt(fields map[string]interface{}, key string) (int, bool) {
	if v, ok := fields[key]; ok {
		switch val := v.(type) {
		case float64:
			return int(val), true
		case int:
			return val, true
		}
	}
	return 0, false
}

func getUint16(fields map[string]interface{}, key string) (uint16, bool) {
	if v, ok := fields[key]; ok {
		switch val := v.(type) {
		case float64:
			return uint16(val), true
		case int:
			return uint16(val), true
		}
	}
	return 0, false
}

func getString(fields map[string]interface{}, key string) (string, bool) {
	if v, ok := fields[key]; ok {
		if val, ok := v.(string); ok {
			return val, true
		}
	}
	return "", false
}

func getBool(fields map[string]interface{}, key string) (bool, bool) {
	if v, ok := fields[key]; ok {
		if val, ok := v.(bool); ok {
			return val, true
		}
	}
	return false, false
}

func getBytes(fields map[string]interface{}, key string) ([]byte, bool) {
	if v, ok := fields[key]; ok {
		switch val := v.(type) {
		case string:
			return []byte(val), true
		case []byte:
			return val, true
		case []interface{}:
			buf := &bytes.Buffer{}
			for _, item := range val {
				switch i := item.(type) {
				case float64:
					buf.WriteByte(byte(i))
				}
			}
			return buf.Bytes(), true
		}
	}
	return nil, false
}

func (s *Server) buildConnect(fields map[string]interface{}) (*mqtt.ConnectPacket, error) {
	p := &mqtt.ConnectPacket{
		ProtocolName:  mqtt.ProtocolName,
		ProtocolLevel: mqtt.ProtocolLevel311,
	}

	if v, ok := getBool(fields, "clean_session"); ok {
		p.CleanSession = v
	} else {
		p.CleanSession = true
	}

	if v, ok := getUint16(fields, "keep_alive"); ok {
		p.KeepAlive = v
	}

	if v, ok := getString(fields, "client_id"); ok {
		p.ClientID = v
	}

	if v, ok := getString(fields, "username"); ok {
		p.UsernameFlag = true
		p.Username = v
	}

	if v, ok := getBytes(fields, "password"); ok {
		p.PasswordFlag = true
		p.Password = v
	}

	if willEnabled, ok := getBool(fields, "will_flag"); ok && willEnabled {
		p.WillFlag = true
		if qos, ok := getInt(fields, "will_qos"); ok {
			p.WillQoS = mqtt.QoS(qos)
		}
		if retain, ok := getBool(fields, "will_retain"); ok {
			p.WillRetain = retain
		}
		if topic, ok := getString(fields, "will_topic"); ok {
			p.WillTopic = topic
		}
		if payload, ok := getBytes(fields, "will_message"); ok {
			p.WillMessage = payload
		}
	}

	if err := p.Validate(); err != nil {
		return nil, err
	}
	return p, nil
}

func (s *Server) buildConnack(fields map[string]interface{}) (*mqtt.ConnackPacket, error) {
	p := &mqtt.ConnackPacket{}
	if v, ok := getBool(fields, "session_present"); ok {
		p.SessionPresent = v
	}
	if v, ok := getInt(fields, "return_code"); ok {
		p.ReturnCode = mqtt.ConnackReturnCode(v)
	}
	return p, nil
}

func (s *Server) buildPublish(fields map[string]interface{}) (*mqtt.PublishPacket, error) {
	p := &mqtt.PublishPacket{}
	if v, ok := getBool(fields, "dup"); ok {
		p.Dup = v
	}
	if v, ok := getInt(fields, "qos"); ok {
		p.QoS = mqtt.QoS(v)
	}
	if v, ok := getBool(fields, "retain"); ok {
		p.Retain = v
	}
	if v, ok := getString(fields, "topic"); ok {
		p.Topic = v
	}
	if v, ok := getUint16(fields, "packet_id"); ok {
		p.PacketID = v
	}
	if v, ok := getBytes(fields, "payload"); ok {
		p.Payload = v
	}
	if err := p.Validate(); err != nil {
		return nil, err
	}
	return p, nil
}

func (s *Server) buildPuback(fields map[string]interface{}) (*mqtt.PubackPacket, error) {
	p := &mqtt.PubackPacket{}
	if v, ok := getUint16(fields, "packet_id"); ok {
		p.PacketID = v
	} else {
		return nil, mqtt.ErrInvalidPacketID
	}
	return p, nil
}

func (s *Server) buildPubrec(fields map[string]interface{}) (*mqtt.PubrecPacket, error) {
	p := &mqtt.PubrecPacket{}
	if v, ok := getUint16(fields, "packet_id"); ok {
		p.PacketID = v
	} else {
		return nil, mqtt.ErrInvalidPacketID
	}
	return p, nil
}

func (s *Server) buildPubrel(fields map[string]interface{}) (*mqtt.PubrelPacket, error) {
	p := &mqtt.PubrelPacket{}
	if v, ok := getUint16(fields, "packet_id"); ok {
		p.PacketID = v
	} else {
		return nil, mqtt.ErrInvalidPacketID
	}
	return p, nil
}

func (s *Server) buildPubcomp(fields map[string]interface{}) (*mqtt.PubcompPacket, error) {
	p := &mqtt.PubcompPacket{}
	if v, ok := getUint16(fields, "packet_id"); ok {
		p.PacketID = v
	} else {
		return nil, mqtt.ErrInvalidPacketID
	}
	return p, nil
}

func (s *Server) buildSubscribe(fields map[string]interface{}) (*mqtt.SubscribePacket, error) {
	p := &mqtt.SubscribePacket{}
	if v, ok := getUint16(fields, "packet_id"); ok {
		p.PacketID = v
	} else {
		return nil, mqtt.ErrInvalidPacketID
	}

	if subsRaw, ok := fields["subscriptions"]; ok {
		if subsList, ok := subsRaw.([]interface{}); ok {
			for _, subRaw := range subsList {
				if subMap, ok := subRaw.(map[string]interface{}); ok {
					sub := mqtt.TopicSubscription{}
					if v, ok := getString(subMap, "topic_filter"); ok {
						sub.TopicFilter = v
					}
					if v, ok := getInt(subMap, "requested_qos"); ok {
						sub.RequestedQoS = mqtt.QoS(v)
					}
					p.Subscriptions = append(p.Subscriptions, sub)
				}
			}
		}
	}

	if err := p.Validate(); err != nil {
		return nil, err
	}
	return p, nil
}

func (s *Server) buildSuback(fields map[string]interface{}) (*mqtt.SubackPacket, error) {
	p := &mqtt.SubackPacket{}
	if v, ok := getUint16(fields, "packet_id"); ok {
		p.PacketID = v
	} else {
		return nil, mqtt.ErrInvalidPacketID
	}

	if rcRaw, ok := fields["return_codes"]; ok {
		if rcList, ok := rcRaw.([]interface{}); ok {
			for _, rcVal := range rcList {
				switch v := rcVal.(type) {
				case float64:
					p.ReturnCodes = append(p.ReturnCodes, mqtt.SubackReturnCode(v))
				case int:
					p.ReturnCodes = append(p.ReturnCodes, mqtt.SubackReturnCode(v))
				case string:
					if val, err := strconv.Atoi(v); err == nil {
						p.ReturnCodes = append(p.ReturnCodes, mqtt.SubackReturnCode(val))
					}
				}
			}
		}
	}

	if len(p.ReturnCodes) == 0 {
		return nil, fmt.Errorf("return_codes is required")
	}
	return p, nil
}

func (s *Server) buildUnsubscribe(fields map[string]interface{}) (*mqtt.UnsubscribePacket, error) {
	p := &mqtt.UnsubscribePacket{}
	if v, ok := getUint16(fields, "packet_id"); ok {
		p.PacketID = v
	} else {
		return nil, mqtt.ErrInvalidPacketID
	}

	if tfRaw, ok := fields["topic_filters"]; ok {
		if tfList, ok := tfRaw.([]interface{}); ok {
			for _, tfVal := range tfList {
				if tf, ok := tfVal.(string); ok {
					p.TopicFilters = append(p.TopicFilters, tf)
				}
			}
		}
	}

	if err := p.Validate(); err != nil {
		return nil, err
	}
	return p, nil
}

func (s *Server) buildUnsuback(fields map[string]interface{}) (*mqtt.UnsubackPacket, error) {
	p := &mqtt.UnsubackPacket{}
	if v, ok := getUint16(fields, "packet_id"); ok {
		p.PacketID = v
	} else {
		return nil, mqtt.ErrInvalidPacketID
	}
	return p, nil
}

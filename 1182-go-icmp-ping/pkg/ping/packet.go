package ping

import (
	"encoding/binary"
	"net"
	"time"
)

func BuildICMPRequest(identifier uint16, seq uint16, timestamp time.Time) []byte {
	packet := make([]byte, ICMPHeaderLen+TimestampSize)
	
	packet[0] = ICMPTypeEchoRequest
	packet[1] = 0
	
	binary.BigEndian.PutUint16(packet[2:4], 0)
	binary.BigEndian.PutUint16(packet[4:6], identifier)
	binary.BigEndian.PutUint16(packet[6:8], seq)
	
	sec := uint32(timestamp.Unix())
	usec := uint32(timestamp.UnixMicro() - int64(sec)*1000000)
	binary.BigEndian.PutUint32(packet[8:12], sec)
	binary.BigEndian.PutUint32(packet[12:16], usec)
	
	checksum := CalculateChecksum(packet)
	binary.BigEndian.PutUint16(packet[2:4], checksum)
	
	return packet
}

func ParseTimestamp(data []byte) time.Time {
	if len(data) < TimestampSize {
		return time.Time{}
	}
	sec := binary.BigEndian.Uint32(data[0:4])
	usec := binary.BigEndian.Uint32(data[4:8])
	return time.Unix(int64(sec), int64(usec)*1000)
}

func ParseIPv4Header(data []byte) (protocol uint8, srcIP net.IP, dstIP net.IP, ttl uint8, headerLen int) {
	if len(data) < IPv4HeaderLen {
		return 0, nil, nil, 0, 0
	}
	
	versionIHL := data[0]
	headerLen = int(versionIHL&0x0f) * 4
	ttl = data[8]
	protocol = data[9]
	srcIP = net.IP{data[12], data[13], data[14], data[15]}
	dstIP = net.IP{data[16], data[17], data[18], data[19]}
	
	return protocol, srcIP, dstIP, ttl, headerLen
}

func ParseICMPHeader(data []byte) (icmpType uint8, code uint8, identifier uint16, seq uint16, dataOffset int) {
	if len(data) < ICMPHeaderLen {
		return 0, 0, 0, 0, 0
	}
	
	icmpType = data[0]
	code = data[1]
	identifier = binary.BigEndian.Uint16(data[4:6])
	seq = binary.BigEndian.Uint16(data[6:8])
	dataOffset = ICMPHeaderLen
	
	return
}

func ParseReceivedPacket(rawData []byte, expectedIdentifier uint16, targetIP net.IP) (*Result, error) {
	proto, _, srcIP, _, ipHeaderLen := ParseIPv4Header(rawData)
	if proto != 1 {
		return nil, nil
	}
	
	icmpData := rawData[ipHeaderLen:]
	icmpType, code, identifier, seq, _ := ParseICMPHeader(icmpData)
	
	if icmpType == ICMPTypeTimeExceeded {
		if code != ICMPCodeTTLExceeded {
			return nil, nil
		}
		
		if len(icmpData) < 8+IPv4HeaderLen+8 {
			return nil, nil
		}
		
		innerIPStart := 8
		innerProto, _, innerDstIP, _, innerIPHeaderLen := ParseIPv4Header(icmpData[innerIPStart:])
		
		if innerProto != 1 {
			return nil, nil
		}
		
		if !innerDstIP.Equal(targetIP) {
			return nil, nil
		}
		
		innerICMPStart := innerIPStart + innerIPHeaderLen
		if len(icmpData) < innerICMPStart+8 {
			return nil, nil
		}
		
		innerICMP := icmpData[innerICMPStart:]
		_, _, innerID, innerSeq, _ := ParseICMPHeader(innerICMP)
		
		if innerID != expectedIdentifier {
			return nil, nil
		}
		
		return &Result{
			Seq:       innerSeq,
			IsTimeout: true,
			TimeExceeded: &TimeExceededInfo{
				RouterIP:       srcIP,
				OriginalTarget: innerDstIP,
				Seq:            innerSeq,
			},
		}, nil
	}
	
	if icmpType != ICMPTypeEchoReply {
		return nil, nil
	}
	
	if identifier != expectedIdentifier {
		return nil, nil
	}
	
	if len(icmpData) < ICMPHeaderLen+TimestampSize {
		return nil, nil
	}
	
	timestampData := icmpData[ICMPHeaderLen : ICMPHeaderLen+TimestampSize]
	sendTime := ParseTimestamp(timestampData)
	rtt := time.Since(sendTime)
	
	return &Result{
		RTT:      rtt,
		TargetIP: srcIP,
		Seq:      seq,
		IsTimeout: false,
	}, nil
}

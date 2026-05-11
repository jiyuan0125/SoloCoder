package ntp

import (
	"encoding/binary"
	"fmt"
	"net"
	"time"
)

const (
	ntpEpochOffset = 2208988800
	ntpPort        = 123
	ntpVersion     = 4
	ntpModeClient  = 3
	ntpModeServer  = 4
	ntpStratumKiss = 0
)

type Packet struct {
	LiVnMode       uint8
	Stratum        uint8
	Poll           int8
	Precision      int8
	RootDelay      uint32
	RootDispersion uint32
	ReferenceID    uint32
	RefTimeSec     uint32
	RefTimeFrac    uint32
	OrigTimeSec    uint32
	OrigTimeFrac   uint32
	RxTimeSec      uint32
	RxTimeFrac     uint32
	TxTimeSec      uint32
	TxTimeFrac     uint32
}

type Result struct {
	ServerTime time.Time
	Offset     time.Duration
	Delay      time.Duration
	Stratum    uint8
	Poll       int8
	Precision  int8
	KissCode   string
}

func NewClientPacket(t time.Time) *Packet {
	sec, frac := timeToNtp(t)
	return &Packet{
		LiVnMode:    (0 << 6) | (ntpVersion << 3) | ntpModeClient,
		TxTimeSec:   sec,
		TxTimeFrac:  frac,
	}
}

func (p *Packet) Version() uint8 {
	return (p.LiVnMode >> 3) & 0x07
}

func (p *Packet) Mode() uint8 {
	return p.LiVnMode & 0x07
}

func (p *Packet) LeapIndicator() uint8 {
	return (p.LiVnMode >> 6) & 0x03
}

func (p *Packet) Serialize() []byte {
	buf := make([]byte, 48)
	buf[0] = p.LiVnMode
	buf[1] = p.Stratum
	buf[2] = uint8(p.Poll)
	buf[3] = uint8(p.Precision)
	binary.BigEndian.PutUint32(buf[4:8], p.RootDelay)
	binary.BigEndian.PutUint32(buf[8:12], p.RootDispersion)
	binary.BigEndian.PutUint32(buf[12:16], p.ReferenceID)
	binary.BigEndian.PutUint32(buf[16:20], p.RefTimeSec)
	binary.BigEndian.PutUint32(buf[20:24], p.RefTimeFrac)
	binary.BigEndian.PutUint32(buf[24:28], p.OrigTimeSec)
	binary.BigEndian.PutUint32(buf[28:32], p.OrigTimeFrac)
	binary.BigEndian.PutUint32(buf[32:36], p.RxTimeSec)
	binary.BigEndian.PutUint32(buf[36:40], p.RxTimeFrac)
	binary.BigEndian.PutUint32(buf[40:44], p.TxTimeSec)
	binary.BigEndian.PutUint32(buf[44:48], p.TxTimeFrac)
	return buf
}

func ParsePacket(buf []byte) (*Packet, error) {
	if len(buf) < 48 {
		return nil, fmt.Errorf("packet too short: %d bytes", len(buf))
	}
	return &Packet{
		LiVnMode:       buf[0],
		Stratum:        buf[1],
		Poll:           int8(buf[2]),
		Precision:      int8(buf[3]),
		RootDelay:      binary.BigEndian.Uint32(buf[4:8]),
		RootDispersion: binary.BigEndian.Uint32(buf[8:12]),
		ReferenceID:    binary.BigEndian.Uint32(buf[12:16]),
		RefTimeSec:     binary.BigEndian.Uint32(buf[16:20]),
		RefTimeFrac:    binary.BigEndian.Uint32(buf[20:24]),
		OrigTimeSec:    binary.BigEndian.Uint32(buf[24:28]),
		OrigTimeFrac:   binary.BigEndian.Uint32(buf[28:32]),
		RxTimeSec:      binary.BigEndian.Uint32(buf[32:36]),
		RxTimeFrac:     binary.BigEndian.Uint32(buf[36:40]),
		TxTimeSec:      binary.BigEndian.Uint32(buf[40:44]),
		TxTimeFrac:     binary.BigEndian.Uint32(buf[44:48]),
	}, nil
}

func timeToNtp(t time.Time) (uint32, uint32) {
	sec := t.Unix() + ntpEpochOffset
	nano := t.Nanosecond()
	frac := uint64(nano) * uint64(1<<32) / uint64(time.Second)
	return uint32(sec), uint32(frac)
}

func ntpToTime(sec uint32, frac uint32) time.Time {
	unixSec := int64(sec) - ntpEpochOffset
	nano := int64(frac) * int64(time.Second) / int64(1<<32)
	return time.Unix(unixSec, nano)
}

func Query(server string, port int, timeout time.Duration) (*Result, error) {
	if port <= 0 {
		port = ntpPort
	}
	addr := fmt.Sprintf("%s:%d", server, port)
	conn, err := net.DialTimeout("udp", addr, timeout)
	if err != nil {
		return nil, fmt.Errorf("dial failed: %w", err)
	}
	defer conn.Close()

	if err := conn.SetDeadline(time.Now().Add(timeout)); err != nil {
		return nil, fmt.Errorf("set deadline failed: %w", err)
	}

	t1 := time.Now()
	req := NewClientPacket(t1)
	if _, err := conn.Write(req.Serialize()); err != nil {
		return nil, fmt.Errorf("write failed: %w", err)
	}

	buf := make([]byte, 48)
	n, err := conn.Read(buf)
	if err != nil {
		return nil, fmt.Errorf("read failed: %w", err)
	}
	t4 := time.Now()

	resp, err := ParsePacket(buf[:n])
	if err != nil {
		return nil, err
	}

	if resp.Mode() != ntpModeServer {
		return nil, fmt.Errorf("unexpected mode: %d", resp.Mode())
	}

	if resp.Stratum == ntpStratumKiss {
		code := kissCode(resp.ReferenceID)
		return &Result{
			KissCode:  code,
			Stratum:   resp.Stratum,
			Poll:      resp.Poll,
			Precision: resp.Precision,
		}, fmt.Errorf("kiss code received: %s", code)
	}

	if resp.Stratum > 15 {
		return nil, fmt.Errorf("invalid stratum: %d", resp.Stratum)
	}

	t2 := ntpToTime(resp.RxTimeSec, resp.RxTimeFrac)
	t3 := ntpToTime(resp.TxTimeSec, resp.TxTimeFrac)

	offset, delay := calculateOffsetAndDelay(t1, t2, t3, t4)

	return &Result{
		ServerTime: t3,
		Offset:     offset,
		Delay:      delay,
		Stratum:    resp.Stratum,
		Poll:       resp.Poll,
		Precision:  resp.Precision,
	}, nil
}

func calculateOffsetAndDelay(t1, t2, t3, t4 time.Time) (time.Duration, time.Duration) {
	t1Ns := t1.UnixNano()
	t2Ns := t2.UnixNano()
	t3Ns := t3.UnixNano()
	t4Ns := t4.UnixNano()

	offsetNs := ((t2Ns - t1Ns) + (t3Ns - t4Ns)) / 2
	delayNs := (t4Ns - t1Ns) - (t3Ns - t2Ns)

	return time.Duration(offsetNs), time.Duration(delayNs)
}

func kissCode(id uint32) string {
	buf := make([]byte, 4)
	binary.BigEndian.PutUint32(buf, id)
	for i, b := range buf {
		if b < 32 || b > 126 {
			buf[i] = '.'
		}
	}
	return string(buf)
}

func PrecisionSec(p int8) float64 {
	if p >= 0 {
		return 1.0
	}
	return 1.0 / float64(uint64(1)<<uint64(-p))
}

func PollSec(p int8) float64 {
	if p <= 0 {
		return 1.0
	}
	return float64(uint64(1) << uint64(p))
}

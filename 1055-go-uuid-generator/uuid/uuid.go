package uuid

import (
	"bytes"
	"crypto/md5"
	"crypto/rand"
	"crypto/sha1"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"fmt"
	"hash"
	"net"
	"sync"
	"time"
)

var (
	ErrInvalidUUID = errors.New("invalid UUID format")
	ErrInvalidVersion = errors.New("invalid UUID version")
	ErrInvalidVariant = errors.New("invalid UUID variant")
	ErrNoNetworkInterface = errors.New("no network interface found")
)

type UUID [16]byte

const (
	VariantNCS       = 0
	VariantRFC4122   = 2
	VariantMicrosoft = 6
	VariantFuture    = 7
)

var (
	NamespaceDNS  = MustParse("6ba7b810-9dad-11d1-80b4-00c04fd430c8")
	NamespaceURL  = MustParse("6ba7b811-9dad-11d1-80b4-00c04fd430c8")
	NamespaceOID  = MustParse("6ba7b812-9dad-11d1-80b4-00c04fd430c8")
	NamespaceX500 = MustParse("6ba7b814-9dad-11d1-80b4-00c04fd430c8")
)

var (
	epoch = time.Date(1582, 10, 15, 0, 0, 0, 0, time.UTC)
	mu    sync.Mutex
	lastTime int64
	clockSeq uint16
	nodeID   []byte
)

func init() {
	var err error
	nodeID, err = getHardwareAddr()
	if err != nil {
		nodeID = make([]byte, 6)
		rand.Read(nodeID)
		nodeID[0] |= 0x01
	}

	clockSeq = uint16(generateRandomUint16())
}

func getHardwareAddr() ([]byte, error) {
	interfaces, err := net.Interfaces()
	if err != nil {
		return nil, err
	}

	for _, iface := range interfaces {
		if len(iface.HardwareAddr) >= 6 && !bytes.Equal(iface.HardwareAddr, []byte{0, 0, 0, 0, 0, 0}) {
			return iface.HardwareAddr[:6], nil
		}
	}

	return nil, ErrNoNetworkInterface
}

func generateRandomUint16() uint16 {
	b := make([]byte, 2)
	rand.Read(b)
	return binary.BigEndian.Uint16(b)
}

func NewV1() (UUID, error) {
	mu.Lock()
	defer mu.Unlock()

	now := getTimestamp()
	if now <= lastTime {
		clockSeq++
	}
	lastTime = now

	var u UUID

	timeLow := uint32(now & 0xFFFFFFFF)
	timeMid := uint16((now >> 32) & 0xFFFF)
	timeHiAndVersion := uint16((now >> 48) & 0x0FFF)
	timeHiAndVersion |= 0x1000

	clockSeqHiAndReserved := uint8((clockSeq >> 8) & 0x3F)
	clockSeqHiAndReserved |= 0x80
	clockSeqLow := uint8(clockSeq & 0xFF)

	binary.BigEndian.PutUint32(u[0:4], timeLow)
	binary.BigEndian.PutUint16(u[4:6], timeMid)
	binary.BigEndian.PutUint16(u[6:8], timeHiAndVersion)
	u[8] = clockSeqHiAndReserved
	u[9] = clockSeqLow
	copy(u[10:], nodeID)

	return u, nil
}

func getTimestamp() int64 {
	now := time.Since(epoch).Nanoseconds() / 100
	now &= 0x0FFFFFFFFFFFFFFF
	return now
}

func NewV3(namespace UUID, name string) UUID {
	return newHashBased(md5.New(), 3, namespace, name)
}

func NewV4() (UUID, error) {
	var u UUID
	_, err := rand.Read(u[:])
	if err != nil {
		return UUID{}, err
	}

	u[6] = (u[6] & 0x0F) | 0x40
	u[8] = (u[8] & 0x3F) | 0x80

	return u, nil
}

func NewV5(namespace UUID, name string) UUID {
	return newHashBased(sha1.New(), 5, namespace, name)
}

func newHashBased(h hash.Hash, version int, namespace UUID, name string) UUID {
	h.Write(namespace[:])
	h.Write([]byte(name))
	hash := h.Sum(nil)

	var u UUID
	copy(u[:], hash[:16])

	u[6] = (u[6] & 0x0F) | uint8(version<<4)
	u[8] = (u[8] & 0x3F) | 0x80

	return u
}

func Parse(s string) (UUID, error) {
	if len(s) != 36 {
		return UUID{}, ErrInvalidUUID
	}

	if s[8] != '-' || s[13] != '-' || s[18] != '-' || s[23] != '-' {
		return UUID{}, ErrInvalidUUID
	}

	hexStr := s[0:8] + s[9:13] + s[14:18] + s[19:23] + s[24:]
	if len(hexStr) != 32 {
		return UUID{}, ErrInvalidUUID
	}

	b, err := hex.DecodeString(hexStr)
	if err != nil {
		return UUID{}, ErrInvalidUUID
	}

	var u UUID
	copy(u[:], b)

	version := u.Version()
	if version < 1 || version > 5 {
		return UUID{}, ErrInvalidVersion
	}

	variant := u.Variant()
	if variant != VariantRFC4122 {
		return UUID{}, ErrInvalidVariant
	}

	return u, nil
}

func MustParse(s string) UUID {
	u, err := Parse(s)
	if err != nil {
		panic(err)
	}
	return u
}

func (u UUID) String() string {
	return fmt.Sprintf("%x-%x-%x-%x-%x",
		u[0:4],
		u[4:6],
		u[6:8],
		u[8:10],
		u[10:16],
	)
}

func (u UUID) Compact() string {
	return fmt.Sprintf("%x", u[:])
}

func (u UUID) Version() int {
	return int(u[6] >> 4)
}

func (u UUID) Variant() int {
	switch {
	case (u[8] & 0x80) == 0x00:
		return VariantNCS
	case (u[8] & 0xC0) == 0x80:
		return VariantRFC4122
	case (u[8] & 0xE0) == 0xC0:
		return VariantMicrosoft
	default:
		return VariantFuture
	}
}

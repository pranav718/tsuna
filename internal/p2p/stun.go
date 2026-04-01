package p2p

import (
	"encoding/binary"
	"fmt"
	"math/rand"
	"net"
	"time"
)

const stunMagicCookie uint32 = 0x2112A442

const (
	stunBindingRequest  uint16 = 0x0001
	stunBindingResponse uint16 = 0x0101
)

const (
	attrMappedAddress    uint16 = 0x0001
	attrXORMappedAddress uint16 = 0x0020
)

const (
	familyIPv4 = 0x01
	familyIPv6 = 0x02
)

var DefaultSTUNServers = []string{
	"stun.l.google.com:19302",
	"stun1.l.google.com:19302",
	"stun.cloudflare.com:3478",
}

type PublicEndpoint struct {
	IP   net.IP
	Port int
}

func (e PublicEndpoint) String() string {
	return fmt.Sprintf("%s:%d", e.IP.String(), e.Port)
}

func (e PublicEndpoint) UDPAddr() *net.UDPAddr {
	return &net.UDPAddr{IP: e.IP, Port: e.Port}
}

func DiscoverPublicEndpoint() (PublicEndpoint, error) {
	return DiscoverPublicEndpointFrom(DefaultSTUNServers)
}
func DiscoverPublicEndpointFrom(servers []string) (PublicEndpoint, error) {
	var lastErr error
	for _, srv := range servers {
		ep, err := querySTUN(srv)
		if err == nil {
			return ep, nil
		}
		lastErr = fmt.Errorf("stun: %s: %w", srv, err)
	}
	return PublicEndpoint{}, fmt.Errorf("stun: all servers failed: %w", lastErr)
}

func querySTUN(addr string) (PublicEndpoint, error) {
	raddr, err := net.ResolveUDPAddr("udp4", addr)
	if err != nil {
		return PublicEndpoint{}, fmt.Errorf("resolve %s: %w", addr, err)
	}
	conn, err := net.ListenUDP("udp4", &net.UDPAddr{})
	if err != nil {
		return PublicEndpoint{}, fmt.Errorf("bind: %w", err)
	}
	defer conn.Close()

	if err := conn.SetDeadline(time.Now().Add(3 * time.Second)); err != nil {
		return PublicEndpoint{}, fmt.Errorf("deadline: %w", err)
	}

	txID := randomTxID()
	req := buildBindingRequest(txID)

	if _, err := conn.WriteToUDP(req, raddr); err != nil {
		return PublicEndpoint{}, fmt.Errorf("send: %w", err)
	}

	buf := make([]byte, 512)
	n, _, err := conn.ReadFromUDP(buf)
	if err != nil {
		return PublicEndpoint{}, fmt.Errorf("recv: %w", err)
	}

	return parseBindingResponse(buf[:n], txID)
}


func buildBindingRequest(txID [12]byte) []byte {
	buf := make([]byte, 20)
	binary.BigEndian.PutUint16(buf[0:2], stunBindingRequest)
	binary.BigEndian.PutUint16(buf[2:4], 0)
	binary.BigEndian.PutUint32(buf[4:8], stunMagicCookie)
	copy(buf[8:20], txID[:])
	return buf
}

func parseBindingResponse(buf []byte, txID [12]byte) (PublicEndpoint, error) {
	if len(buf) < 20 {
		return PublicEndpoint{}, fmt.Errorf("response too short: %d bytes", len(buf))
	}

	msgType := binary.BigEndian.Uint16(buf[0:2])
	if msgType != stunBindingResponse {
		return PublicEndpoint{}, fmt.Errorf("unexpected message type: 0x%04x", msgType)
	}

	cookie := binary.BigEndian.Uint32(buf[4:8])
	if cookie != stunMagicCookie {
		return PublicEndpoint{}, fmt.Errorf("bad magic cookie: 0x%08x", cookie)
	}

	var respTxID [12]byte
	copy(respTxID[:], buf[8:20])
	if respTxID != txID {
		return PublicEndpoint{}, fmt.Errorf("transaction ID mismatch")
	}

	msgLen := int(binary.BigEndian.Uint16(buf[2:4]))
	if len(buf) < 20+msgLen {
		return PublicEndpoint{}, fmt.Errorf("truncated message: need %d bytes, have %d", 20+msgLen, len(buf))
	}

	attrs := buf[20 : 20+msgLen]
	var fallback *PublicEndpoint

	for len(attrs) >= 4 {
		attrType := binary.BigEndian.Uint16(attrs[0:2])
		attrLen := int(binary.BigEndian.Uint16(attrs[2:4]))
		attrVal := attrs[4:]

		if len(attrVal) < attrLen {
			break
		}

		switch attrType {
		case attrXORMappedAddress:
			ep, err := parseXORMappedAddress(attrVal[:attrLen])
			if err != nil {
				return PublicEndpoint{}, fmt.Errorf("XOR-MAPPED-ADDRESS: %w", err)
			}
			return ep, nil

		case attrMappedAddress:
			ep, err := parseMappedAddress(attrVal[:attrLen])
			if err == nil {
				fallback = &ep
			}
		}

		padded := (attrLen + 3) &^ 3
		attrs = attrs[4+padded:]
	}

	if fallback != nil {
		return *fallback, nil
	}

	return PublicEndpoint{}, fmt.Errorf("no mapped address attribute in response")
}

func parseXORMappedAddress(val []byte) (PublicEndpoint, error) {
	if len(val) < 8 {
		return PublicEndpoint{}, fmt.Errorf("too short: %d bytes", len(val))
	}

	family := val[1]
	if family != familyIPv4 {
		return PublicEndpoint{}, fmt.Errorf("unsupported address family: %d (only IPv4 supported)", family)
	}

	rawPort := binary.BigEndian.Uint16(val[2:4])
	port := int(rawPort ^ uint16(stunMagicCookie>>16))

	rawIP := binary.BigEndian.Uint32(val[4:8])
	xorIP := rawIP ^ stunMagicCookie

	ip := make(net.IP, 4)
	binary.BigEndian.PutUint32(ip, xorIP)

	return PublicEndpoint{IP: ip, Port: port}, nil
}

func parseMappedAddress(val []byte) (PublicEndpoint, error) {
	if len(val) < 8 {
		return PublicEndpoint{}, fmt.Errorf("too short: %d bytes", len(val))
	}

	family := val[1]
	if family != familyIPv4 {
		return PublicEndpoint{}, fmt.Errorf("unsupported address family: %d", family)
	}

	port := int(binary.BigEndian.Uint16(val[2:4]))

	ip := make(net.IP, 4)
	copy(ip, val[4:8])

	return PublicEndpoint{IP: ip, Port: port}, nil
}

func randomTxID() [12]byte {
	var id [12]byte
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	for i := range id {
		id[i] = byte(r.Intn(256))
	}
	return id
}

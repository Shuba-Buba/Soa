package proxy

import (
	"bufio"
	"fmt"
	"net"
	"strings"
)

type Server struct {
	Port            int64
	MulticastAddr   string
	ConvertersAddrs map[string]string
	Result          chan string
}

func (s Server) ListenMulticastGroup() error {
	maddr, err := net.ResolveUDPAddr("udp", s.MulticastAddr)
	if err != nil {
		return fmt.Errorf("failed to resolve udp addr: %v", err)
	}
	conn, err := net.ListenMulticastUDP("udp", nil, maddr)
	if err != nil {
		return fmt.Errorf("failed to listen multicast udp: %v", err)
	}
	defer conn.Close()

	for {
		for range s.ConvertersAddrs {
			buf := make([]byte, 1000)
			n, _, err := conn.ReadFromUDP(buf)
			if err != nil {
				return fmt.Errorf("failed to read from UDP conn: %v", err)
			}
			if string(buf[:n]) == "get_info" {
				continue
			}
			s.Result <- string(buf[:n])
		}
	}
}

func (s Server) ProcessMulticast() (string, error) {
	common_addr, err := net.ResolveUDPAddr("udp", s.MulticastAddr)
	if err != nil {
		return "", fmt.Errorf("error with resole address")
	}

	conn, err := net.DialUDP("udp", nil, common_addr)
	if err != nil {
		return "", fmt.Errorf("error in udp common addr")
	}
	defer conn.Close()

	_, err = conn.Write([]byte("Let's go"))
	if err != nil {
		return "", fmt.Errorf("failed to write to addr")
	}

	var result []string
	for i := 0; i < cap(s.Result); i++ {
		if str := <-s.Result; str != "Let's go" {
			result = append(result, str)
		} else {
			result = append(result, <-s.Result)
		}
	}

	return strings.Join(result, ""), nil
}

func (s Server) ProcessRequest(request []byte) (string, error) {
	req := strings.Trim(string(request), "\n")
	var addr string
	if req == "All" {
		return s.ProcessMulticast()
	}

	addr, ok := s.ConvertersAddrs[req]
	if !ok {
		return "", fmt.Errorf("No Valid Converter %v\n", string(request))
	}
	conn, err := net.Dial("udp", addr)
	if err != nil {
		return "", fmt.Errorf("Bad address")
	}
	defer conn.Close()

	_, err = fmt.Fprintf(conn, "Let's go")
	if err != nil {
		return "", fmt.Errorf("failed to write data to addr %q: %v", addr, err)
	}

	buf := make([]byte, 1024)
	n, err := bufio.NewReader(conn).Read(buf)
	if err != nil {
		return "", fmt.Errorf("failed to read data from connection: %v", err)
	}
	return string(buf[:n]), nil
}

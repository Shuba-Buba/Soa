package converters

import (
	"fmt"
	"net"
	"strings"
	"time"
)

type Server struct {
	Port          int64
	Format        string
	MulticastAddr string
	converter     Converter
}

func NewServer(port int64, format string, multicastAddr string, converter Converter) *Server {
	return &Server{
		Port:          port,
		Format:        format,
		MulticastAddr: multicastAddr,
		converter:     converter,
	}
}

func (s Server) SendToGroup(data string) error {
	maddr, err := net.ResolveUDPAddr("udp", s.MulticastAddr)
	if err != nil {
		return fmt.Errorf("failed to resolve udp addr: %v", err)
	}
	c, err := net.DialUDP("udp", nil, maddr)

	if err != nil {
		return fmt.Errorf("failed to set listen conn: %v", err)
	}
	defer c.Close()
	_, err = c.Write([]byte(data))
	if err != nil {
		return fmt.Errorf("failed to write data: %v", err)
	}
	return nil
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
		buf := make([]byte, 1000)
		n, _, err := conn.ReadFromUDP(buf)
		if err != nil {
			fmt.Printf("failed to read data: %v", err)
			continue
		}
		data, err := s.ProcessRequest(buf[:n])
		if err != nil {
			fmt.Printf("failed to process request: %v", err)
			continue
		}
		if data == "" {
			continue
		}
		err = s.SendToGroup(data)
		if err != nil {
			fmt.Printf("failed to send to group: %v", err)
		}
	}

}

func (s Server) ProcessConverter(format string, converter Converter) (string, error) {
	person := &MyStruct{
		String_: "ShubaBuba",
		Int_:    1337228,
		Map_: map[string]string{
			"A": "a",
			"B": "b",
			"C": "C",
			"D": "d",
		},
		Array_: []string{
			"123", "abcabadabacaba", "bad", "good", "666",
		},
		Float_: 1.2,
	}

	var totalTimeSerialize int64
	var totalTimeDeserialize int64
	var totalStructSize int

	for i := 0; i < 1000; i++ {
		start := time.Now()
		bytes, err := converter.Serialize(person)
		totalTimeSerialize += time.Since(start).Microseconds()
		if err != nil {
			return "", fmt.Errorf("failed to serialize string: %v", err)
		}
		totalStructSize += len(bytes)

		start = time.Now()
		_, err = converter.Deserialize(bytes)
		totalTimeDeserialize += time.Since(start).Microseconds()
		if err != nil {
			return "", fmt.Errorf("failed to deserialize string: %v", err)
		}
	}
	return fmt.Sprintf(
		"%s - %d - %dmcs - %dmcs\n",
		format, totalStructSize/1000, totalTimeSerialize/1000, totalTimeDeserialize/1000), nil
}

func (s Server) ProcessRequest(buf []byte) (string, error) {
	req := strings.Trim(string(buf), "\n")
	if req != "Let's go" {
		return "", nil
	}

	res, err := s.ProcessConverter(s.Format, s.converter)
	if err != nil {
		return "", fmt.Errorf("failed to process converter")
	}
	return res, nil
}

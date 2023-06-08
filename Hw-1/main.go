package main

import (
	"encoding/json"
	"fmt"
	"hw_1_serialization/converters"
	"hw_1_serialization/converters/XML"
	"hw_1_serialization/converters/avro"
	my_json "hw_1_serialization/converters/json"
	"hw_1_serialization/converters/msgpack"
	"hw_1_serialization/converters/native"
	"hw_1_serialization/converters/proto"
	"hw_1_serialization/converters/yaml"
	"hw_1_serialization/models"
	"hw_1_serialization/proxy"
	"io/ioutil"
	"net"
	"os"
	"strconv"
	"sync"
)

const (
	Ok  = 0
	Bad = 1
)

type Worker interface {
	ProcessRequest([]byte) (string, error)
}

type Sheduler struct {
	groupAddr string
	port      int64
	host      string
	worker    Worker
}

type MutlicastSheduler interface {
	ListenMulticastGroup() error
}

func MakeSheduler(serverType string, modelsInfo models.Serializators) (*Sheduler, error) {
	groupAddr := os.Getenv("GROUP_ADDRESS")
	if groupAddr == "" {
		return nil, fmt.Errorf("GROUP_ADDR is empty\n")
	}

	var host string
	var cur_worker Worker

	convertersAddrs := make(map[string]string)
	var port int64
	cnt_servers := 0

	for _, val := range modelsInfo.GetArray() {
		if val.Name == serverType {
			port = int64(val.Port)
		} else {
			convertersAddrs[val.Name] = val.Name + ":" + strconv.Itoa(val.Port)
			cnt_servers += 1
		}
	}

	switch serverType {
	case "proxy":
		cur_worker = proxy.Server{
			Port:            port,
			ConvertersAddrs: convertersAddrs,
			MulticastAddr:   groupAddr,
			Result:          make(chan string, cnt_servers),
		}
	case "xml":
		host = "xml"
		cur_worker = converters.NewServer(port, "xml", groupAddr, &XML.Converter{})
	case "proto":
		host = "proto"
		cur_worker = converters.NewServer(port, "proto", groupAddr, &proto.Converter{})
	case "native":
		host = "native"
		cur_worker = converters.NewServer(port, "native", groupAddr, &native.Converter{})
	case "json":
		host = "json"
		cur_worker = converters.NewServer(port, "json", groupAddr, &my_json.Converter{})
	case "yaml":
		host = "yaml"
		cur_worker = converters.NewServer(port, "yaml", groupAddr, &yaml.Converter{})
	case "msgpack":
		host = "msgpack"
		cur_worker = converters.NewServer(port, "msgpack", groupAddr, &msgpack.Converter{})
	case "avro":
		host = "avro"
		conv := &avro.Converter{}
		err := conv.SetSchema()
		if err != nil {
			return nil, fmt.Errorf("failed to set schema: %v", err)
		}
		cur_worker = converters.NewServer(port, "avro", groupAddr, conv)
	}

	ctrl := &Sheduler{
		host:      host,
		port:      port,
		worker:    cur_worker,
		groupAddr: groupAddr,
	}
	return ctrl, nil
}

func (this Sheduler) Listen() error {
	conn, err := net.ListenPacket("udp", fmt.Sprintf("%s:%d", this.host, this.port))
	if err != nil {
		return fmt.Errorf("Error in ListenPacket")
	}
	defer conn.Close()

	for {
		buf := make([]byte, 4096)
		n, addr, err := conn.ReadFrom(buf)
		if err != nil {
			return fmt.Errorf("Error in ReadFrom")
		}

		res, err := this.worker.ProcessRequest(buf[:n])
		if err != nil {
			fmt.Printf("failed to process request: %v", err)
			conn.WriteTo([]byte(err.Error()+"\n"), addr)
			continue
		}
		_, err = conn.WriteTo([]byte(res), addr)
		if err != nil {
			return fmt.Errorf("Error with Write to addsress")
		}
	}
}

func main() {
	os.Exit(run())

}

func run() (exitCode int) {
	exitCode = Bad

	if len(os.Args) == 1 {
		fmt.Println("Need enought 2 arguments\nError in docker-compose")
		return
	}
	var modelsInfo models.Serializators

	content, err := ioutil.ReadFile("base.json")
	if err != nil {
		fmt.Printf("Bad base.json or bad json: %v", err)
		return
	}

	if err := json.Unmarshal(content, &modelsInfo); err != nil {
		fmt.Printf("Error in Unmarshal: %v", err)
		return
	}

	sheduler, err := MakeSheduler(os.Args[1], modelsInfo)
	if err != nil {
		fmt.Printf("Error in MakeSheduler %v", err)
		return
	}

	wg := sync.WaitGroup{}
	wg.Add(1)

	go func() {
		defer wg.Done()
		sheduler.Listen()
	}()

	go func() {
		defer wg.Done()
		obj, ok := sheduler.worker.(MutlicastSheduler)
		if !ok {
			return
		}
		obj.ListenMulticastGroup()
	}()

	wg.Wait()

	return Ok
}

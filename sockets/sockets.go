package sockets

import (
	"encoding/gob"
	"fmt"
	"net"
	"os"
	"sync"
	"time"

	"github.com/orange-lightsaber/psb-rotatord/rotator"
)

type Request struct {
	Request string
	RCD     rotator.RunConfigData
}

type Response struct {
	Response string
	Error    string
}

const (
	LastRun_Req = "lastrun"
	InitRun_Req = "init"
	Rotate_Req  = "rotate"
	Socket      = "/tmp/psb_rotator.sock"
)

func handleConnection(c net.Conn, reqHandle func(*Request) *Response) {
	decoder := gob.NewDecoder(c)
	req := &Request{}
	decoder.Decode(req)
	res := reqHandle(req)
	encoder := gob.NewEncoder(c)
	encoder.Encode(res)
	c.Close()
}

func Open(reqHandle func(*Request) *Response) error {
	if _, err := os.Stat(Socket); !os.IsNotExist(err) {
		os.Remove(Socket)
	}
	l, err := net.Listen("unix", Socket)
	if err != nil {
		return fmt.Errorf("listen error: %s", err.Error())
	}
	err = os.Chmod(Socket, 0777)
	if err != nil {
		return fmt.Errorf("error changing file permissions on %s: %s", Socket, err.Error())
	}
	for {
		c, err := l.Accept()
		if err != nil {
			continue
		}
		go handleConnection(c, reqHandle)
	}
}

func (req *Request) NewRequest() (res *Response, err error) {
	c, err := net.Dial("unix", Socket)
	if err != nil {
		err = fmt.Errorf("connection error: %s", err.Error())
		return
	}
	defer c.Close()
	var wg sync.WaitGroup
	wg.Add(1)
	r := &Response{}
	go func() {
		defer wg.Done()
		decoder := gob.NewDecoder(c)
		timeout := time.After(10 * time.Second)
		tick := time.Tick(50 * time.Millisecond)
	AwaitResponse:
		for {
			select {
			case <-timeout:
				err = fmt.Errorf("timed out during request: %+v", req)
				break AwaitResponse
			case <-tick:
				decoder.Decode(r)
				if r.Response != "" || r.Error != "" {
					break AwaitResponse
				}
			}
		}
	}()
	encoder := gob.NewEncoder(c)
	encoder.Encode(req)
	wg.Wait()
	res = r
	return
}

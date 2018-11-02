package dipper

import (
	"io"
	"log"
	"os"
	"strings"
	"time"
)

// MessageHandler : a type of functions that take a pointer to a message and handle it
type MessageHandler func(*Message)

// RPCHandler : a type of functions that handle RPC calls between drivers
type RPCHandler func(string, string, []byte)

// Driver : the helper stuct for creating a honey-dipper driver in golang
type Driver struct {
	RPCCaller
	Name            string
	Service         string
	State           string
	In              io.Reader
	Out             io.Writer
	Options         interface{}
	MessageHandlers map[string]MessageHandler
	Start           MessageHandler
	Stop            MessageHandler
	Reload          MessageHandler
	RPCHandlers     map[string]RPCHandler
	ReadySignal     chan bool
}

// NewDriver : create a blank driver object
func NewDriver(service string, name string) *Driver {
	driver := Driver{
		Name:        name,
		Service:     service,
		State:       "loaded",
		In:          os.Stdin,
		Out:         os.Stdout,
		RPCHandlers: map[string]RPCHandler{},
	}

	driver.MessageHandlers = map[string]MessageHandler{
		"command:options": driver.ReceiveOptions,
		"command:ping":    driver.Ping,
		"command:start":   driver.start,
		"command:stop":    driver.stop,
	}

	driver.Sender = &driver

	return &driver
}

// Run : start a loop to communicate with daemon
func (d *Driver) Run() {
	log.Printf("[%s-%s] driver loaded\n", d.Service, d.Name)
	for {
		func() {
			defer SafeExitOnError("[%s-%s] Resuming driver message loop", d.Service, d.Name)
			defer CatchError(io.EOF, func() {
				log.Fatalf("[%s-%s] daemon closed channel", d.Service, d.Name)
			})
			for {
				msg := FetchRawMessage(d.In)
				go func() {
					defer SafeExitOnError("[%s-%s] Continuing driver message loop", d.Service, d.Name)
					if msg.Channel == "rpcReply" {
						d.HandleRPCReturn(msg)
					} else if msg.Channel == "rpc" {
						d.handleRPC(msg)
					} else if handler, ok := d.MessageHandlers[msg.Channel+":"+msg.Subject]; ok {
						handler(msg)
					} else {
						log.Printf("[%s-%s] skipping message without handler: %s:%s", d.Service, d.Name, msg.Channel, msg.Subject)
					}
				}()
			}
		}()
	}
}

// Ping : respond to daemon ping request with driver state
func (d *Driver) Ping(msg *Message) {
	d.SendMessage("state", d.State, nil)
}

// ReceiveOptions : receive options from daemon
func (d *Driver) ReceiveOptions(msg *Message) {
	msg = DeserializePayload(msg)
	d.Options = msg.Payload
	d.ReadySignal <- true
}

func (d *Driver) start(msg *Message) {
	select {
	case <-d.ReadySignal:
	case <-time.After(time.Second):
	}

	if d.State == "alive" {
		if d.Reload != nil {
			d.Reload(msg)
		} else {
			d.State = "cold"
		}
	} else {
		if d.Start != nil {
			d.Start(msg)
		}
		d.State = "alive"
	}
	d.Ping(msg)
}

func (d *Driver) stop(msg *Message) {
	d.State = "exit"
	if d.Stop != nil {
		d.Stop(msg)
	}
	d.Ping(msg)
	log.Fatalf("[%s-%s] quiting on daemon request", d.Service, d.Name)
}

// SendRawMessage : construct and send a message to daemon
func (d *Driver) SendRawMessage(channel string, subject string, payload []byte) {
	log.Printf("[%s-%s] sending raw message to daemon %s:%s", d.Service, d.Name, channel, subject)
	SendRawMessage(d.Out, channel, subject, payload)
}

// SendMessage : send a prepared message to daemon
func (d *Driver) SendMessage(channel string, subject string, payload interface{}) {
	log.Printf("[%s-%s] sending raw message to daemon %s:%s", d.Service, d.Name, channel, subject)
	SendMessage(d.Out, channel, subject, payload)
}

// GetOption : get the data from options map with the key
func (d *Driver) GetOption(path string) (interface{}, bool) {
	return GetMapData(d.Options, path)
}

// GetOptionStr : get the string data from options map with the key
func (d *Driver) GetOptionStr(path string) (string, bool) {
	return GetMapDataStr(d.Options, path)
}

// RPCError : return error to rpc caller
func (d *Driver) RPCError(from string, rpcID string, reason string) {
	d.SendMessage("rpcReply", from+"."+rpcID+"."+"err", map[string]interface{}{"reason": reason})
	log.Panicf("[%s-%s] rpc returning err %s", d.Service, d.Name, reason)
}

// RPCReturn : return a value to rpc caller
func (d *Driver) RPCReturn(from string, rpcID string, retval interface{}) {
	d.SendMessage("rpcReply", from+"."+rpcID, retval)
}

// RPCReturnRaw : return a raw value to rpc caller
func (d *Driver) RPCReturnRaw(from string, rpcID string, retval []byte) {
	d.SendRawMessage("rpcReply", from+"."+rpcID, retval)
}

func (d *Driver) handleRPC(msg *Message) {
	parts := strings.SplitN(msg.Subject, ".", 3)
	if len(parts) < 3 {
		log.Panicf("[%s-%s] malformated subject for rpc call %s", d.Service, d.Name, msg.Subject)
	}
	method := parts[0]
	rpcID := parts[1]
	from := parts[2]
	rf, ok := d.RPCHandlers[method]
	if ok {
		rf(from, rpcID, msg.Payload.([]byte))
	} else {
		f, ok := d.RPCHandlers[method]
		if !ok {
			log.Panicf("[%s-%s] RPC handler not defined for method %s", d.Service, d.Name, method)
		}
		f(from, rpcID, msg.Payload.([]byte))
	}
}

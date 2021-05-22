package networkMiner

import (
	"crypto/tls"
	"encoding/gob"
	"fmt"
	"math/rand"
	"net"
	"sync"

	log "github.com/sirupsen/logrus"
)

const (
	_ = iota
	FactomdEventForward
)

type NetworkMessage struct {
	NetworkCommand int
	Data           interface{}
	Version        string
}

type TCPClient struct {
	// Miner related fields
	//PegnetMinerFields

	id int // Random

	// Tags for stats
	tagLock sync.Mutex
	tags    map[string]string

	conn     net.Conn
	encoder  *gob.Encoder
	decoder  *gob.Decoder
	Server   *TCPServer
	closed   bool
	accepted bool
}

func NewTCPClient(conn net.Conn, s *TCPServer) *TCPClient {
	m := new(TCPClient)
	m.conn = conn
	m.Server = s
	m.init()
	m.id = rand.Int()
	m.tags = make(map[string]string)

	return m
}

func (c *TCPClient) LogFields() log.Fields {
	ip := "unknown"
	if c.conn != nil { // Protect against an nil dereference
		ip = c.conn.RemoteAddr().String()
	}

	return log.Fields{
		"ip": ip,
		"id": c.id,
	}
}

func (c *TCPClient) init() {
	c.decoder = gob.NewDecoder(c.conn)
	c.encoder = gob.NewEncoder(c.conn)
}

// Read client data from channel
func (c *TCPClient) listen() {
	c.Server.onNewClientCallback(c)
	for {
		var m NetworkMessage
		err := c.decoder.Decode(&m)
		if err != nil {
			c.close(err)
			return
		}
		c.Server.onNewMessage(c, &m)
	}
}

// SendNetworkCommand text message to client
func (c *TCPClient) SendNetworkCommand(message *NetworkMessage) error {
	err := c.encoder.Encode(message)
	return err
}

func (c *TCPClient) Conn() net.Conn {
	return c.conn
}

func (c *TCPClient) close(err error) error {
	if c.closed {
		return nil // Only close once
	}
	c.closed = true

	defer c.Server.onClientConnectionClosed(c, err)
	return c.conn.Close()
}

func (c *TCPClient) Close() error {
	return c.close(fmt.Errorf("closed by server"))
}

// TCPServer is heavily inspired by https://github.com/firstrow/tcp_server
type TCPServer struct {
	Host                     string
	config                   *tls.Config
	onNewClientCallback      (func(c *TCPClient))
	onClientConnectionClosed func(c *TCPClient, err error)
	onNewMessage             func(c *TCPClient, message *NetworkMessage)
}

// Called right after server starts listening new client
func (s *TCPServer) OnNewClient(callback func(c *TCPClient)) {
	s.onNewClientCallback = callback
}

// Called right after connection closed
func (s *TCPServer) OnClientConnectionClosed(callback func(c *TCPClient, err error)) {
	s.onClientConnectionClosed = callback
}

// Called when Client receives new message
func (s *TCPServer) OnNewMessage(callback func(c *TCPClient, message *NetworkMessage)) {
	s.onNewMessage = callback
}

// Listen starts network server
func (s *TCPServer) Listen() {
	var listener net.Listener
	var err error
	if s.config == nil {
		listener, err = net.Listen("tcp", s.Host)
	} else {
		listener, err = tls.Listen("tcp", s.Host, s.config)
	}
	if err != nil {
		log.WithError(err).Fatal("Error starting TCP server.")
	}
	log.Info("Listening on ", s.Host)
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.WithError(err).Error("failed to accept client")
		}
		client := NewTCPClient(conn, s)
		client.init()
		go client.listen()
	}
}

// Creates new tcp server instance
func NewTCPServer(host string) *TCPServer {
	log.Println("Creating server with address", host)
	server := &TCPServer{
		Host:   host,
		config: nil,
	}

	server.OnNewClient(func(c *TCPClient) {})
	server.OnNewMessage(func(c *TCPClient, message *NetworkMessage) {})
	server.OnClientConnectionClosed(func(c *TCPClient, err error) {})

	return server
}

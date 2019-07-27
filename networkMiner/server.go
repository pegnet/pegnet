package networkMiner

import (
	"context"
	"encoding/gob"
	"sync"

	"github.com/FactomProject/factom"

	"github.com/pegnet/pegnet/common"
	"github.com/pegnet/pegnet/opr"
	log "github.com/sirupsen/logrus"
	"github.com/zpatrick/go-config"
)

const (
	_ = iota
	Ping
	Pong

	FactomEvent
	GraderEvent
	ConstructedOPR
)

func init() {
	gob.Register(common.MonitorEvent{})
	gob.Register(opr.OPRs{})
	gob.Register(factom.Entry{})
	gob.Register(opr.OraclePriceRecord{})
}

type MiningServer struct {
	config *config.Config

	FactomMonitor common.IMonitor
	OPRGrader     opr.IGrader
	Host          string

	Server *TCPServer

	clientsLock sync.Mutex
	clients     map[int]*TCPClient
}

func NewMiningServer(config *config.Config, monitor common.IMonitor, grader opr.IGrader) *MiningServer {
	var err error
	s := new(MiningServer)
	s.config = config

	s.clients = make(map[int]*TCPClient)
	s.FactomMonitor = monitor
	s.OPRGrader = grader

	s.Host, err = config.String("Miner.MiningCoordinator")
	if err != nil {
		panic(err)
	}

	s.Server = NewTCPServer(s.Host)
	s.Server.onNewClientCallback = s.onNewClient
	s.Server.onNewMessage = s.onNewMessage
	s.Server.onClientConnectionClosed = s.onClientConnectionClosed

	return s
}

func (s *MiningServer) Listen() {
	s.Server.Listen()
}

func (c *MiningServer) ForwardMonitorEvents() {
	fLog := log.WithFields(log.Fields{"func": "ForwardMonitorEvents"})
	alert := c.FactomMonitor.NewListener()
	gAlerts := c.OPRGrader.GetAlert("evt-forwarder")
	var last common.MonitorEvent
	for {
		select {
		case fds := <-alert:
			m := new(NetworkMessage)
			m.NetworkCommand = FactomEvent
			m.Data = fds
			last = fds

			c.clientsLock.Lock()
			for _, c := range c.clients {
				err := c.SendNetworkCommand(m)
				if err != nil {
					fLog.WithField("evt", "factom").WithError(err).Error("failed to send")
				}
			}
			c.clientsLock.Unlock()
			fLog.WithFields(log.Fields{
				"height": fds.Dbht,
				"minute": fds.Minute,
			}).Debug("server sent alert")
		case g := <-gAlerts:
			//m := new(NetworkMessage)
			//m.NetworkCommand = GraderEvent
			//m.Data = *g
			//
			//c.clientsLock.Lock()
			//for _, c := range c.clients {
			//	err := c.SendNetworkCommand(m)
			//	if err != nil {
			//		fLog.WithField("evt", "grader").WithError(err).Error("failed to send")
			//	}
			//}
			//c.clientsLock.Unlock()

			opr, err := opr.NewOprFromWinners(context.Background(), 0, last.Dbht, c.config, g)
			if err != nil {
				fLog.WithField("evt", "grader").WithError(err).Error("failed to make opr")
			}

			m := new(NetworkMessage)
			m.NetworkCommand = ConstructedOPR
			m.Data = *opr
			c.clientsLock.Lock()
			for _, c := range c.clients {
				err := c.SendNetworkCommand(m)
				if err != nil {
					fLog.WithField("evt", "opr").WithError(err).Error("failed to send")
				}
			}
			c.clientsLock.Unlock()

			fLog.WithFields(c.Fields()).Info("sent opr to miners")

		}
	}
}

// onNewMessage is when the client messages us.
func (n *MiningServer) onNewMessage(c *TCPClient, message *NetworkMessage) {
	log.Debugf("client message: %v", message)
}

func (s *MiningServer) onClientConnectionClosed(c *TCPClient, err error) {
	s.clientsLock.Lock()
	defer s.clientsLock.Unlock()

	var _ = c.Close()
	delete(s.clients, c.id)
	log.WithFields(s.Fields()).Info("Client disconnected")
}

func (s *MiningServer) onNewClient(c *TCPClient) {
	s.clientsLock.Lock()
	defer s.clientsLock.Unlock()

	s.clients[c.id] = c
	log.WithFields(s.Fields()).WithField("id", c.id).Info("Client connected")

	var _ = c.SendNetworkCommand(&NetworkMessage{NetworkCommand: Ping})
}

func (s *MiningServer) Fields() log.Fields {
	return log.Fields{"clients": len(s.clients)}
}

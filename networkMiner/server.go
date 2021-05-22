package networkMiner

import (
	"context"
	"crypto/sha256"
	"encoding/gob"
	"errors"
	"fmt"
	"math/rand"
	"sync"
	"time"

	"github.com/FactomProject/factom"
	"github.com/cenkalti/backoff"
	"github.com/pegnet/pegnet/common"
	"github.com/pegnet/pegnet/mining"
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
	FactomEntry
	MiningStatistics
	AddTag
	SecretChallenge
	RejectedConnection // Sever rejected client
	CoordinatorError
)

// Idk why the factom.entry does not work
type GobbedEntry struct {
	ChainID string   `json:"chainid"`
	ExtIDs  [][]byte `json:"extids"`
	Content []byte   `json:"content"`
}

type Tag struct {
	Key   string
	Value string
}

// ErrorMessage allows us to send w/e errors we want to a client
type ErrorMessage struct {
	Error string
}

// AuthenticationChallenge is used for request + response
type AuthenticationChallenge struct {
	Challenge string // Request
	Response  string // Response
}

func init() {
	gob.Register(common.MonitorEvent{})
	gob.Register(opr.OPRs{})
	gob.Register(GobbedEntry{})
	gob.Register([][]byte{})
	gob.Register(opr.OraclePriceRecord{})
	gob.Register(mining.GroupMinerStats{})
	gob.Register(Tag{})
	gob.Register(AuthenticationChallenge{})
	gob.Register(ErrorMessage{})
}

// MiningServer is the coordinator to emit events to anyone listening
type MiningServer struct {
	config *config.Config

	FactomMonitor common.IMonitor
	OPRGrader     opr.IGrader
	Host          string

	Server *TCPServer
	EC     *factom.ECAddress

	Stats *mining.GlobalStatTracker

	clientsLock sync.Mutex
	clients     map[int]*TCPClient
	numClients  int
	salt        int // Random salt on each boot
	secret      string
	useAuth     bool
}

func NewMiningServer(config *config.Config, monitor common.IMonitor, grader opr.IGrader, stats *mining.GlobalStatTracker) *MiningServer {
	var err error
	s := new(MiningServer)
	s.config = config

	s.clients = make(map[int]*TCPClient)
	s.FactomMonitor = monitor
	s.OPRGrader = grader
	s.Stats = stats

	s.Host, err = config.String(common.ConfigCoordinatorListen)
	if err != nil {
		panic(err)
	}

	if ecadrStr, err := config.String("Miner.ECAddress"); err != nil {
		log.WithError(err).Fatalf("missing ec addr in config")
	} else {
		ecAdr, err := factom.FetchECAddress(ecadrStr)
		if err != nil {
			log.WithError(err).Fatalf("could not fetch ec addr")
		}
		s.EC = ecAdr
	}

	// Set our callbacks
	s.Server = NewTCPServer(s.Host)
	s.Server.onNewClientCallback = s.onNewClient
	s.Server.onNewMessage = s.onNewMessage
	s.Server.onClientConnectionClosed = s.onClientConnectionClosed

	// We use random in our authentication. I think using crypto/rand is overkill.
	// It is a simple authentication protocol, all miners connecting in should be trusted.
	rand.Seed(time.Now().UnixNano())
	s.salt = rand.Int()
	s.useAuth, err = config.Bool(common.ConfigCoordinatorUseAuthentication)
	if err != nil {
		log.WithError(err).Fatalf("missing coordinator use authentication in config")
	}

	s.secret, err = config.String(common.ConfigCoordinatorSecret)
	if err != nil {
		log.WithError(err).Fatalf("missing authentication secret in config")
	}

	return s
}

func (s *MiningServer) Listen() {
	s.Server.Listen()
}

// ForwardMonitorEvents will forward all the events we get to anyone listening
func (c *MiningServer) ForwardMonitorEvents() {
	fLog := log.WithFields(log.Fields{"func": "ForwardMonitorEvents"})
	alert := c.FactomMonitor.NewListener()
	gAlerts := c.OPRGrader.GetAlert("evt-forwarder")
	var last common.MonitorEvent
	mining := false
	for {
		select {
		case fds := <-alert: // Push factom events straight to miners
			// If we do not have an EC balance, do not push events to start mining.
			// Minute 1 is where we start mining, ensure we have some ECs.
			if fds.Minute == 1 {
				bal, err := factom.GetECBalance(c.EC.String())
				if err != nil {
					fLog.WithField("evt", "factom").WithError(err).Error("failed to send, balance query failed")
					coordError := fmt.Errorf("balance query failed on net coordinator")
					c.SendToClients(&NetworkMessage{NetworkCommand: CoordinatorError, Data: ErrorMessage{coordError.Error()}}, fLog.WithField("evt", "err"))
					break
				}

				if bal == 0 {
					fLog.WithField("evt", "factom").WithError(fmt.Errorf("%s balance is 0", c.EC.String())).Error("you do not have any ECs left to mine.")
					coordError := fmt.Errorf("you do not have any ECs left to mine, %s balance is 0", c.EC.String())
					c.SendToClients(&NetworkMessage{NetworkCommand: CoordinatorError, Data: ErrorMessage{coordError.Error()}}, fLog.WithField("evt", "err"))
					break
				}
			}

			if fds.Minute == 9 && mining {
				mining = false
			} else if fds.Minute == 1 {
				mining = true
			}

			m := new(NetworkMessage)
			m.NetworkCommand = FactomEvent
			m.Data = fds
			last = fds

			c.SendToClients(m, fLog.WithField("evt", "factom"))
			fLog.WithFields(log.Fields{
				"height": fds.Dbht,
				"minute": fds.Minute,
			}).Debug("server sent alert")
		case g := <-gAlerts:
			if !mining {
				break // If we are not mining, we do not do anything
			}
			tmpChan := make(chan *opr.OPRs, 1)
			tmpChan <- g
			oprobject, err := opr.NewOpr(context.Background(), 0, last.Dbht, c.config, tmpChan)
			if err != nil {
				fLog.WithField("evt", "grader").WithError(err).Error("failed to make opr")
			}

			m := new(NetworkMessage)
			m.NetworkCommand = ConstructedOPR
			if oprobject == nil {
				fLog.WithField("evt", "grader").Error("failed to make opr. opr is nil")
				m.Data = nil
			} else {
				m.Data = *oprobject
			}

			c.SendToClients(m, fLog.WithField("evt", "opr"))
			fLog.WithFields(c.Fields()).Info("sent opr to miners")
		}
	}
}

func (n *MiningServer) SendToClients(message *NetworkMessage, logger *log.Entry) {
	n.clientsLock.Lock()
	defer n.clientsLock.Unlock()
	for _, c := range n.clients {
		err := c.SendNetworkCommand(message)
		if err != nil {
			logger.WithError(err).Error("failed to send")
		}
	}
}

// onNewMessage is when the client messages us.
func (n *MiningServer) onNewMessage(c *TCPClient, message *NetworkMessage) {
	if !c.accepted {
		switch message.NetworkCommand {
		case SecretChallenge, AddTag, Ping, Pong:
		// Let these commands through
		default:
			// We won't listen to most messages if they are not accepted yet from
			// the challenge
			return
		}
	}

	switch message.NetworkCommand {
	case AddTag:
		b, ok := message.Data.(Tag)
		if !ok {
			log.WithFields(n.Fields()).Errorf("client did not send a proper tag")
			return
		}

		c.tagLock.Lock()
		c.tags[b.Key] = b.Value
		c.tags["version"] = message.Version
		c.tagLock.Unlock()
	case Pong:
	case Ping:
		err := c.SendNetworkCommand(NewNetworkMessage(Pong, nil))
		if err != nil {
			log.WithFields(n.Fields()).WithError(err).Errorf("failed to pong")
		}
	case FactomEntry: // They want us to write an entry
		b, ok := message.Data.(GobbedEntry)
		if !ok {
			log.WithFields(n.Fields()).Errorf("client did not send a proper entry")
			return
		}

		e := new(factom.Entry)
		e.ExtIDs = b.ExtIDs
		e.Content = b.Content
		e.ChainID = b.ChainID

		// Note: we could take the self reported difficulty and do some filtering
		// Right now we just submit directly

		// Thread the write
		go func() {
			err := n.WriteEntry(e)
			if err != nil {
				log.WithFields(n.Fields()).WithError(err).Errorf("failed to submit entry from client")
			} else {
				log.WithFields(n.Fields()).WithField("client", c.id).Debugf("submitted entry %x", e.Hash())
			}
		}()
	case MiningStatistics:
		g, ok := message.Data.(mining.GroupMinerStats)
		if !ok {
			log.WithFields(n.Fields()).Errorf("client did not send a proper entry")
			return
		}

		// Modify the stats so we know it came from us
		g.ID = fmt.Sprintf("Net-%d", c.id)
		g.Tags["src"] = c.conn.RemoteAddr().String()

		c.tagLock.Lock()
		for k, v := range c.tags {
			g.Tags[k] = v
		}
		c.tagLock.Unlock()

		n.Stats.MiningStatsChannel <- &g
	case SecretChallenge:
		// Response for challenge from client
		challengeResp, ok := message.Data.(AuthenticationChallenge)
		if !ok {
			log.WithFields(n.Fields()).Errorf("client did not send a proper entry")
			return
		}

		// Get the expected challenge data
		challenge := n.GetAuthenticationChallenge(c)

		// If the user is responding to another challenge, they are wrong.
		if challengeResp.Challenge != challenge.Challenge {
			err := c.SendNetworkCommand(NewNetworkMessage(RejectedConnection, "challenge data did not match expected"))
			if err != nil {
				log.WithFields(n.Fields()).WithError(err).Errorf("client failed challenge")
			}
			var _ = c.Close()
			return
		}

		// Challenge Resp is sha256(secret+challenge)
		resp := sha256.Sum256([]byte(n.secret + challenge.Challenge))
		challenge.Response = fmt.Sprintf("%x", resp)

		// Check the user's challenge response
		if challengeResp.Response != challenge.Response {
			err := c.SendNetworkCommand(NewNetworkMessage(RejectedConnection, "challenge response is incorrect"))
			if err != nil {
				log.WithFields(n.Fields()).WithError(err).Errorf("client failed challenge")
			}
			var _ = c.Close()
			return
		}

		// Client is good, let them mine
		n.Accept(c)
	case RejectedConnection:
	// Clients don't reject servers. They just leave
	default:
		log.WithFields(n.Fields()).WithField("cmd", message.NetworkCommand).Warn("command not recognized from client")
	}
}

func (s *MiningServer) GetAuthenticationChallenge(c *TCPClient) AuthenticationChallenge {
	return AuthenticationChallenge{Challenge: fmt.Sprintf("%d%d", s.salt, c.id)}
}

func (s *MiningServer) onClientConnectionClosed(c *TCPClient, err error) {
	s.clientsLock.Lock()
	defer s.clientsLock.Unlock()

	delete(s.clients, c.id)
	s.numClients = len(s.clients)
	log.WithFields(s.Fields()).WithFields(c.LogFields()).Info("Client disconnected")
}

func (s *MiningServer) onNewClient(c *TCPClient) {
	if s.useAuth {
		challenge := s.GetAuthenticationChallenge(c)
		err := c.SendNetworkCommand(NewNetworkMessage(SecretChallenge, challenge))
		if err != nil {
			log.WithFields(s.Fields()).WithError(err).WithField("func", "onNewClient").Error("failed to send challenge")
			return
		}
		log.WithFields(s.Fields()).WithFields(c.LogFields()).Info("Client pending challenge")
	} else {
		err := c.SendNetworkCommand(NewNetworkMessage(Ping, nil))
		if err != nil {
			log.WithFields(s.Fields()).WithError(err).WithField("func", "onNewClient").Error("ping failed")
			return
		}

		s.Accept(c)
	}
}

func (s *MiningServer) Accept(c *TCPClient) {
	s.clientsLock.Lock()
	defer s.clientsLock.Unlock()
	c.accepted = true

	s.clients[c.id] = c
	s.numClients = len(s.clients)
	log.WithFields(s.Fields()).WithFields(c.LogFields()).Info("Client connected")
}

func (s *MiningServer) WriteEntry(entry *factom.Entry) error {
	operation := func() error {
		_, err1 := factom.CommitEntry(entry, s.EC)
		_, err2 := factom.RevealEntry(entry)
		if err1 == nil && err2 == nil {
			return nil
		}

		return errors.New("Unable to commit entry to factom")
	}

	err := backoff.Retry(operation, common.PegExponentialBackOff())
	return err
}

func (s *MiningServer) Fields() log.Fields {
	return log.Fields{"clients": s.numClients}
}

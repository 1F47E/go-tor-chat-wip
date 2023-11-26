package client

import (
	"context"
	"fmt"
	"net"
	"strings"

	"github.com/1F47E/go-shaihulud/internal/client/connection"
	"github.com/1F47E/go-shaihulud/internal/client/listner"
	client_local "github.com/1F47E/go-shaihulud/internal/client/local"
	"github.com/1F47E/go-shaihulud/internal/client/message"
	client_tor "github.com/1F47E/go-shaihulud/internal/client/tor"
	myaes "github.com/1F47E/go-shaihulud/internal/cryptotools/aes"
	"github.com/1F47E/go-shaihulud/internal/cryptotools/auth"
	"github.com/1F47E/go-shaihulud/internal/interfaces"
	"github.com/1F47E/go-shaihulud/internal/tui"
	zlog "github.com/rs/zerolog/log"
)

// can be local or tor
type Connector interface {
	RunServer(port int, onionPrivKey []byte) (net.Listener, error)
	RunClient(address string) (net.Conn, error)
}

type ConnectionType int

const (
	Local ConnectionType = iota
	Tor
)

type Client struct {
	ctx       context.Context
	cancel    context.CancelFunc
	msgCh     chan message.Message
	eventsCh  chan tui.Event // tui
	connector Connector
	crypter   interfaces.Asymmetric
	user      *connection.Connection
	listner   *listner.Listner
	connType  ConnectionType
}

func NewClient(ctx context.Context, cancel context.CancelFunc, connType ConnectionType, crypter interfaces.Asymmetric, eventsCh chan tui.Event) *Client {
	msgCh := make(chan message.Message)
	var connector Connector

	// init connector debug or tor
	switch connType {
	case Local:
		connector = client_local.New(ctx, cancel, msgCh)
	case Tor:
		connector = client_tor.New(ctx, cancel, msgCh)
	}

	// create listner
	lCtx, lCancel := context.WithCancel(ctx)
	lstnr := listner.New(lCtx, lCancel, msgCh)

	return &Client{
		ctx:       ctx,
		cancel:    cancel,
		msgCh:     msgCh,
		eventsCh:  eventsCh,
		connector: connector,
		crypter:   crypter,
		listner:   lstnr,
		connType:  connType,
	}
}

type ChatAuth struct {
	AccessKey string `json:"access_key"`
	Password  string `json:"password"`
}

func (c *Client) CreateListner(port int, keys []byte) (net.Listener, error) {
	return c.connector.RunServer(port, keys)
}

func (c *Client) GenerateAuth() (*auth.Auth, error) {
	session := ""
	// generate auth key and password
	crypter := myaes.New()
	auth, err := auth.New(crypter, session)
	if err != nil {
		zlog.Error().Msgf("cant create auth: %v\n", err)
		return nil, err
	}

	// listener, err := c.connector.RunServer(address, auth.Onion().PrivKey())
	return auth, nil
}

func (c *Client) RunListner(listener net.Listener) {

	// DEBUG
	// return &a, nil

	// accept incoming connections
	go func() {
		for {
			select {
			case <-c.ctx.Done():
				zlog.Debug().Msgf("RunServer listner context done: %v\n", c.ctx.Err())
				return

			default:
				zlog.Debug().Msg("Client.RunServer: Waiting for a connection")

				conn, err := listener.Accept()
				if err != nil {
					zlog.Error().Msgf("Client.RunServer listener.Accept error: %v\n", err)
					continue
				}
				user := connection.New(conn) // connection with user data
				c.user = user
				zlog.Debug().Msg("Client.RunServer: Got a connection")

				c.eventsCh <- tui.NewEventSpin("Accepting incoming connection...")

				// Create a new Listner for each connection
				ctx, cancel := context.WithCancel(c.ctx)
				listner := listner.New(ctx, cancel, c.msgCh)
				go listner.Sender(user, c.crypter)
				go listner.Receiver(user, c.crypter)
				// go c.ListenUserInput()
				c.eventsCh <- tui.NewEventText("User connected")
			}
		}
	}()
}

func (c *Client) AuthVerify(key, password string) error {
	// create auth struct and try to decode key
	aes := myaes.New()
	ath, err := auth.NewFromKey(aes, key, password)
	if err != nil {
		if strings.Contains(err.Error(), "authentication failed") {
			return fmt.Errorf("wrong password")
		}
		return fmt.Errorf("Access key error")
	}

	msg := "✅ Access granted, connecting..."
	c.eventsCh <- tui.NewEventText(msg)
	msg = fmt.Sprintf("Connecting to %s...", ath.OnionAddressTransport())
	c.eventsCh <- tui.NewEventText(msg)
	return nil
}

func (c *Client) RunClient() error {

	// ===== At this point access key and pass are valid

	// get address to connect to
	address := ""
	output := ""
	switch c.connType {
	case Local:
		address = "localhost:3000"
		output = fmt.Sprintf("Connecting to %s...", address)
		zlog.Debug().Msg(output)
	case Tor:
		// address = ath.OnionAddressFull() // BUG: assign onion address on init
		address := "demo.onion"
		output = "Starting TOR..."
		zlog.Debug().Msgf("Starting tor, connecting to onion address: %v\n", address)
	default:
		return fmt.Errorf("unknown connection type: %v\n", c.connType)
	}
	c.eventsCh <- tui.NewEventSpin(output)

	// Run the connector
	conn, err := c.connector.RunClient(address)
	if err != nil {
		return fmt.Errorf("cant connect to server: %v\n", err)
	}
	user := connection.New(conn) // connection with user data
	c.user = user

	// Run the listener, sender, and input listener goroutines
	ctx, cancel := context.WithCancel(c.ctx)
	c.listner = listner.New(ctx, cancel, c.msgCh)
	go c.listner.Sender(user, c.crypter)
	go c.listner.Receiver(user, c.crypter)
	// go c.ListenUserInput()

	return nil
}

func (c *Client) Close() {
	if c.user != nil && c.user.Conn != nil {
		c.user.Conn.Close()
	}
}

// send message
func (c *Client) Send(msg string) error {
	zlog.Debug().Msgf("crypter: %v\n", c.crypter)
	inputCipher, err := c.crypter.Encrypt([]byte(msg), c.user.PubKey)
	if err != nil {
		return err
	}
	zlog.Debug().Msgf("inputCipher: %d %v\n", len(inputCipher), inputCipher)
	c.msgCh <- message.NewMSG(inputCipher)
	return nil
}

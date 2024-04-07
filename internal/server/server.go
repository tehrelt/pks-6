package server

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"os"
	"xmpp/internal/server/manager"
	"xmpp/pkg/message"
	"xmpp/pkg/packets"
)

type Server struct {
	listener net.Listener
	manager  *manager.AccountManager
	*Options
}

type Options struct {
	messageBus chan *message.Message
	logger     *slog.Logger
	Verbose    bool
}

func New(host string, port int, options *Options) (*Server, error) {
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", host, port))
	if err != nil {
		return nil, err
	}

	if options.Verbose == false {
		options.logger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	} else {
		options.logger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	}

	options.logger = options.logger.With(slog.String("addr", fmt.Sprintf("%s:%d", host, port)))

	return &Server{
		listener: listener,
		manager:  manager.New(options.logger.With("account-manager")),
		Options:  options,
	}, nil
}

func (s *Server) Run() error {
	s.logger.Info("Server listening")
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			return err
		}

		s.logger.Info("Incoming connection", slog.String("incaddr", conn.RemoteAddr().String()))

		go s.handleConnection(conn)
	}
}

func (s *Server) handleConnection(conn net.Conn) {
	defer func(conn net.Conn) {
		err := conn.Close()
		if err != nil {
			fmt.Println("error closing connection:", err.Error())
		}
	}(conn)

	for {
		buffer := make([]byte, 4096)

		_, err := conn.Read(buffer)
		if err != nil {
			if errors.Is(err, io.EOF) {
				s.logger.Info("Connection closed", slog.String("incaddr", conn.RemoteAddr().String()))
				return
			}
			s.logger.Error("error reading from connection:", slog.String("error", err.Error()))
			s.manager.Close(conn)

			return
		}

		ctx := context.WithValue(context.Background(), "conn", conn)
		s.handlePacket(ctx, buffer)
	}
}

func (s *Server) Close() error {
	return s.listener.Close()
}

func (s *Server) handlePacket(ctx context.Context, bytes []byte) {

	packetType := int(bytes[0])

	s.logger.Debug("packet received", slog.Int("type", packetType))

	switch packetType {
	case packets.PACKET_AUTH:
		s.handleAuthPacket(ctx, bytes)
	case packets.PACKET_SEND:
		s.handleSendPacket(ctx, bytes)
	case packets.PACKET_CLOSE:
		s.handleClosePacket(ctx, bytes)
	case packets.PACKET_ERROR:
	default:
		panic("unhandled default case")
	}
}

func (s *Server) handleAuthPacket(ctx context.Context, bytes []byte) {
	s.logger.Debug("auth packet received")

	conn, ok := ctx.Value("conn").(net.Conn)
	if !ok {
		s.logger.Error("no conn in context")
		return
	}

	p := packets.UnmarshalAuthPacket(bytes)

	if err := s.manager.Auth(p.Username, conn); err != nil {
		s.logger.Warn("error authenticating:", err.Error())

		errorPacket := packets.NewErrorPacket(err.Error())

		if _, err := conn.Write(packets.MarshalErrorPacket(errorPacket)); err != nil {
			s.logger.Error("error sending error packet:", err.Error())
			return
		}

		return
	}

	packetOk := packets.NewPacket(packets.PACKET_OK)
	if _, err := conn.Write(packets.MarshalPacket(packetOk)); err != nil {
		s.logger.Error("error sending ok packet:", err.Error())
		return
	}

	s.logger.Info("authenticated", slog.String("username", p.Username))
}

func (s *Server) handleSendPacket(ctx context.Context, bytes []byte) {

	conn, ok := ctx.Value("conn").(net.Conn)
	if !ok {
		s.logger.Error("no conn in context")
		return
	}

	p := packets.UnmarshalSendPacket(bytes)
	s.logger.Debug("send packet received", slog.String("to", p.To), slog.String("data", string(p.Data)), slog.String("from", p.From))
	_, err := s.manager.Get(p.To)
	if err != nil {
		s.logger.Error("error getting account:", err.Error())
		packetError := packets.NewErrorPacket(err.Error())
		if _, err := conn.Write(packets.MarshalErrorPacket(packetError)); err != nil {
			s.logger.Error("error sending error packet:", err.Error())
			return
		}
		return
	}

	go func() {
		m := message.NewMessage(p.To, p.From, p.Data)

		recvPacket := packets.NewReceivePacket(m.From, m.Text)

		s.logger.Debug(fmt.Sprintf("sending message from %s to %s", m.From, m.To))

		s.logger.Debug(fmt.Sprintf("getting account %s", m.To))
		acc, err := s.manager.Get(m.To)
		if err != nil {
			s.logger.Error("error getting account:", err.Error())
			return
		}

		s.logger.Debug(fmt.Sprintf("sending message to %s(%s)", acc.Username, acc.RemoteAddr()))
		_, err = acc.Write(packets.MarshalReceivePacket(recvPacket))
		if err != nil {
			s.logger.Error("error sending message:", err.Error())
			return
		}
	}()
}

func (s *Server) handleClosePacket(ctx context.Context, bytes []byte) {
	p := packets.UnmarshalClosePacket(bytes)
	s.logger.Debug(fmt.Sprintf("close packet received from %s", p.Username))
	s.manager.Logout(p.Username)
}

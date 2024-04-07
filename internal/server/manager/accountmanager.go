package manager

import (
	"fmt"
	"log/slog"
	"net"
	"sync"
	"xmpp/pkg/message"
)

type Account struct {
	Username string
	net.Conn
	Messages chan *message.Message
}

type AccountManager struct {
	accounts map[string]*Account
	logger   *slog.Logger
	sync.Mutex
}

func New(logger *slog.Logger) *AccountManager {
	return &AccountManager{
		accounts: make(map[string]*Account),
		Mutex:    sync.Mutex{},
		logger:   logger,
	}
}

func (m *AccountManager) Auth(account string, c net.Conn) error {
	m.Lock()
	defer m.Unlock()
	m.logger.Info("account created", slog.String("account", account))

	if _, ok := m.accounts[account]; ok {
		m.logger.Error("account already exists", slog.String("account", account))
		return fmt.Errorf("account already exists")
	}

	acc := &Account{
		Username: account,
		Conn:     c,
		Messages: make(chan *message.Message),
	}

	m.accounts[account] = acc
	return nil
}

func (m *AccountManager) Get(account string) (*Account, error) {
	m.Lock()
	defer m.Unlock()
	acc, ok := m.accounts[account]
	if !ok {
		return nil, fmt.Errorf("account '%s' not found", account)
	}
	return acc, nil
}

func (m *AccountManager) Logout(account string) {
	m.Lock()
	defer m.Unlock()
	m.logger.Info("account logged out", slog.String("account", account))

	delete(m.accounts, account)
}

func (m *AccountManager) Online() []*Account {
	m.Lock()
	defer m.Unlock()

	online := make([]*Account, 0)

	for _, c := range m.accounts {
		online = append(online, c)
	}

	return online
}

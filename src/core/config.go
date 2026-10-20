package core

import (
	"log/slog"
	"net/netip"
)

type Config struct {
	TrustedProxies []netip.Addr
}

var AppConfig Config

func InitConfig() {
	// Example: trust loopback and a specific proxy IP
	// In a real app, these would be loaded from env or a file
	AppConfig.TrustedProxies = []netip.Addr{
		netip.MustParseAddr("127.0.0.1"),
		netip.MustParseAddr("::1"),
	}
	slog.Info("Config initialized with trusted proxies")
}

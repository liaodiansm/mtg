package network

import (
	"context"
	"fmt"
	"math/rand"
	"net/url"

	"github.com/liaodiansm/mtg/essentials"
)

type loadBalancedWssDialer struct {
	dialers []Dialer
}

func (l loadBalancedWssDialer) Dial(network, address string) (essentials.Conn, error) {
	return l.DialContext(context.Background(), network, address)
}

func (l loadBalancedWssDialer) DialContext(ctx context.Context, network, address string) (essentials.Conn, error) {
	length := len(l.dialers)
	start := rand.Intn(length)
	moved := false

	for i := start; i != start || !moved; i = (i + 1) % length {
		moved = true

		if conn, err := l.dialers[i].DialContext(ctx, network, address); err == nil {
			return conn, nil
		}
	}

	return nil, ErrCannotDialWithAllProxies
}

// NewLoadBalancedWssDialer builds a new load balancing websocket dialer.
//
// The main difference from one which is made by NewWssDialer is that we
// actually have a list of these proxies. When dial is requested, any proxy is
// picked and used. If proxy fails for some reason, we try another one.
//
// So, it is mostly useful if you have some routes with proxies which are not
// always online or having buggy network.
func NewLoadBalancedWssDialer(baseDialer Dialer, proxyURLs []*url.URL) (Dialer, error) {
	dialers := make([]Dialer, 0, len(proxyURLs))

	for _, u := range proxyURLs {
		dialer, err := NewWssDialer(newProxyDialer(baseDialer, u), u)
		if err != nil {
			return nil, fmt.Errorf("cannot build dialer for %s: %w", u.String(), err)
		}

		dialers = append(dialers, dialer)
	}

	return loadBalancedWssDialer{
		dialers: dialers,
	}, nil
}

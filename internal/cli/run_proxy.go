package cli

import (
	"fmt"
	"net"
	"net/url"
	"os"

	"github.com/liaodiansm/mtg/antireplay"
	"github.com/liaodiansm/mtg/events"
	"github.com/liaodiansm/mtg/internal/config"
	"github.com/liaodiansm/mtg/internal/utils"
	"github.com/liaodiansm/mtg/logger"
	"github.com/liaodiansm/mtg/mtglib"
	"github.com/liaodiansm/mtg/network"
	"github.com/rs/zerolog"
)

func makeLogger(conf *config.Config) mtglib.Logger {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnixMs
	zerolog.TimestampFieldName = "timestamp"
	zerolog.LevelFieldName = "level"

	if conf.Debug.Get(false) {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	} else {
		zerolog.SetGlobalLevel(zerolog.WarnLevel)
	}

	baseLogger := zerolog.New(os.Stdout).With().Timestamp().Logger()

	return logger.NewZeroLogger(baseLogger)
}

func makeNetwork(conf *config.Config, version string) (mtglib.Network, error) {
	tcpTimeout := conf.Network.Timeout.TCP.Get(network.DefaultTimeout)
	httpTimeout := conf.Network.Timeout.HTTP.Get(network.DefaultHTTPTimeout)
	dohIP := conf.Network.DOHIP.Get(net.ParseIP(network.DefaultDOHHostname)).String()
	userAgent := "mtg/" + version

	baseDialer, err := network.NewDefaultDialer(tcpTimeout, 0)
	if err != nil {
		return nil, fmt.Errorf("cannot build a default dialer: %w", err)
	}

	if len(conf.Network.Proxies) == 0 {
		return network.NewNetwork(baseDialer, userAgent, dohIP, httpTimeout) //nolint: wrapcheck
	}

	proxyURLs := make([]*url.URL, 0, len(conf.Network.Proxies))

	for _, v := range conf.Network.Proxies {
		if value := v.Get(nil); value != nil {
			proxyURLs = append(proxyURLs, value)
		}
	}

	if len(proxyURLs) == 1 {
		wssDialer, err := network.NewWssDialer(baseDialer, proxyURLs[0])
		if err != nil {
			return nil, fmt.Errorf("cannot build wss dialer: %w", err)
		}

		return network.NewNetwork(wssDialer, userAgent, dohIP, httpTimeout) //nolint: wrapcheck
	}

	wssDialer, err := network.NewLoadBalancedWssDialer(baseDialer, proxyURLs)
	if err != nil {
		return nil, fmt.Errorf("cannot build wss dialer: %w", err)
	}

	return network.NewNetwork(wssDialer, userAgent, dohIP, httpTimeout) //nolint: wrapcheck
}

func makeAntiReplayCache(conf *config.Config) mtglib.AntiReplayCache {
	if !conf.Defense.AntiReplay.Enabled.Get(false) {
		return antireplay.NewNoop()
	}

	return antireplay.NewStableBloomFilter(
		conf.Defense.AntiReplay.MaxSize.Get(antireplay.DefaultStableBloomFilterMaxSize),
		conf.Defense.AntiReplay.ErrorRate.Get(antireplay.DefaultStableBloomFilterErrorRate),
	)
}

func makeEventStream() (mtglib.EventStream, error) {
	factories := make([]events.ObserverFactory, 0, 2) //nolint: gomnd

	if len(factories) > 0 {
		return events.NewEventStream(factories), nil
	}

	return events.NewNoopStream(), nil
}

func runProxy(conf *config.Config, version string) error { //nolint: funlen
	logger := makeLogger(conf)

	logger.BindJSON("configuration", conf.String()).Debug("configuration")

	eventStream, err := makeEventStream()
	if err != nil {
		return fmt.Errorf("cannot build event stream: %w", err)
	}

	ntw, err := makeNetwork(conf, version)
	if err != nil {
		return fmt.Errorf("cannot build network: %w", err)
	}

	opts := mtglib.ProxyOpts{
		Logger:          logger,
		Network:         ntw,
		AntiReplayCache: makeAntiReplayCache(conf),
		EventStream:     eventStream,

		Secret:             conf.Secret,
		DomainFrontingPort: conf.DomainFrontingPort.Get(mtglib.DefaultDomainFrontingPort),
		PreferIP:           conf.PreferIP.Get(mtglib.DefaultPreferIP),

		AllowFallbackOnUnknownDC: conf.AllowFallbackOnUnknownDC.Get(false),
		TolerateTimeSkewness:     conf.TolerateTimeSkewness.Value,
	}

	proxy, err := mtglib.NewProxy(opts)
	if err != nil {
		return fmt.Errorf("cannot create a proxy: %w", err)
	}

	listener, err := utils.NewListener(conf.BindTo.Get(""), 0)
	if err != nil {
		return fmt.Errorf("cannot start proxy: %w", err)
	}

	ctx := utils.RootContext()

	go proxy.Serve(listener) //nolint: errcheck

	<-ctx.Done()
	listener.Close()
	proxy.Shutdown()

	return nil
}

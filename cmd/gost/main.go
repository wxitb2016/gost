package main

import (
	"crypto/sha256"
	"crypto/tls"
	"flag"
	"fmt"
	"net"
	// _ "net/http/pprof"
	"os"
	"runtime"
	"time"

	"github.com/ginuerzh/gost"
	"github.com/go-log/log"
)

var (
	options route
	routes  []route
)

func init() {
	gost.SetLogger(&gost.LogLogger{})

	var (
		configureFile string
		printVersion  bool
	)

	flag.Var(&options.ChainNodes, "F", "forward address, can make a forward chain")
	flag.Var(&options.ServeNodes, "L", "listen address, can listen on multiple ports")
	flag.StringVar(&configureFile, "C", "", "configure file")
	flag.BoolVar(&options.Debug, "D", false, "enable debug log")
	flag.BoolVar(&printVersion, "V", false, "print version")
	flag.Parse()

	if printVersion {
		fmt.Fprintf(os.Stderr, "gost %s (%s)\n", gost.Version, runtime.Version())
		os.Exit(0)
	}

	if len(options.ServeNodes) > 0 {
		routes = append(routes, options)
	}
	gost.Debug = options.Debug

	if err := loadConfigureFile(configureFile); err != nil {
		log.Log(err)
		os.Exit(1)
	}

	if flag.NFlag() == 0 || len(routes) == 0 {
		flag.PrintDefaults()
		os.Exit(0)
	}

}

func main() {
	// go func() {
	// 	log.Log(http.ListenAndServe("localhost:6060", nil))
	// }()
	// NOTE: as of 2.6, you can use custom cert/key files to initialize the default certificate.
	config, err := tlsConfig(defaultCertFile, defaultKeyFile)
	if err != nil {
		// generate random self-signed certificate.
		cert, err := gost.GenCertificate()
		if err != nil {
			log.Log(err)
			os.Exit(1)
		}
		config = &tls.Config{
			Certificates: []tls.Certificate{cert},
		}
	}
	gost.DefaultTLSConfig = config

	for _, route := range routes {
		if err := route.serve(); err != nil {
			log.Log(err)
			os.Exit(1)
		}
	}

	select {}
}

type route struct {
	ChainNodes, ServeNodes stringList
	Retries                int
	Debug                  bool
}

func (r *route) initChain() (*gost.Chain, error) {
	chain := gost.NewChain()
	chain.Retries = r.Retries
	gid := 1 // group ID

	for _, ns := range r.ChainNodes {
		ngroup := gost.NewNodeGroup()
		ngroup.ID = gid
		gid++

		// parse the base node
		nodes, err := parseChainNode(ns)
		if err != nil {
			return nil, err
		}

		nid := 1 // node ID

		for i := range nodes {
			nodes[i].ID = nid
			nid++
		}
		ngroup.AddNode(nodes...)

		// parse peer nodes if exists
		peerCfg, err := loadPeerConfig(nodes[0].Get("peer"))
		if err != nil {
			log.Log(err)
		}
		peerCfg.Validate()

		strategy := peerCfg.Strategy
		// overwrite the strategry in the peer config if `strategy` param exists.
		if s := nodes[0].Get("strategy"); s != "" {
			strategy = s
		}
		ngroup.Options = append(ngroup.Options,
			gost.WithFilter(&gost.FailFilter{
				MaxFails:    peerCfg.MaxFails,
				FailTimeout: time.Duration(peerCfg.FailTimeout) * time.Second,
			}),
			gost.WithStrategy(parseStrategy(strategy)),
		)

		for _, s := range peerCfg.Nodes {
			nodes, err = parseChainNode(s)
			if err != nil {
				return nil, err
			}

			for i := range nodes {
				nodes[i].ID = nid
				nid++
			}

			ngroup.AddNode(nodes...)
		}

		var bypass *gost.Bypass
		// global bypass
		if peerCfg.Bypass != nil {
			bypass = gost.NewBypassPatterns(peerCfg.Bypass.Patterns, peerCfg.Bypass.Reverse)
		}
		nodes = ngroup.Nodes()
		for i := range nodes {
			if nodes[i].Bypass == nil {
				nodes[i].Bypass = bypass // use global bypass if local bypass does not exist.
			}
		}

		chain.AddNodeGroup(ngroup)
	}

	return chain, nil
}

func parseChainNode(ns string) (nodes []gost.Node, err error) {
	node, err := gost.ParseNode(ns)
	if err != nil {
		return
	}

	users, err := parseUsers(node.Get("secrets"))
	if err != nil {
		return
	}
	if node.User == nil && len(users) > 0 {
		node.User = users[0]
	}
	serverName, sport, _ := net.SplitHostPort(node.Addr)
	if serverName == "" {
		serverName = "localhost" // default server name
	}

	rootCAs, err := loadCA(node.Get("ca"))
	if err != nil {
		return
	}
	tlsCfg := &tls.Config{
		ServerName:         serverName,
		InsecureSkipVerify: !node.GetBool("secure"),
		RootCAs:            rootCAs,
	}

	var tr gost.Transporter
	switch node.Transport {
	case "quic":
		config := &gost.QUICConfig{
			TLSConfig:   tlsCfg,
			KeepAlive:   node.GetBool("keepalive"),
			Timeout:     time.Duration(node.GetInt("timeout")) * time.Second,
			IdleTimeout: time.Duration(node.GetInt("idle")) * time.Second,
		}

		if cipher := node.Get("cipher"); cipher != "" {
			sum := sha256.Sum256([]byte(cipher))
			config.Key = sum[:]
		}

		tr = gost.QUICTransporter(config)

	default:
		tr = gost.TCPTransporter()
	}

	var connector gost.Connector
	switch node.Protocol {
	case "socks", "socks5":
		connector = gost.SOCKS5Connector(node.User)
	case "socks4":
		connector = gost.SOCKS4Connector()
	case "socks4a":
		connector = gost.SOCKS4AConnector()
	case "forward":
		connector = gost.ForwardConnector()
	case "http":
		fallthrough
	default:
		node.Protocol = "http" // default protocol is HTTP
		connector = gost.HTTPConnector(node.User)
	}

	timeout := node.GetInt("timeout")
	node.DialOptions = append(node.DialOptions,
		gost.TimeoutDialOption(time.Duration(timeout)*time.Second),
	)

	handshakeOptions := []gost.HandshakeOption{
		gost.AddrHandshakeOption(node.Addr),
		gost.HostHandshakeOption(node.Host),
		gost.UserHandshakeOption(node.User),
		gost.IntervalHandshakeOption(time.Duration(node.GetInt("ping")) * time.Second),
		gost.TimeoutHandshakeOption(time.Duration(timeout) * time.Second),
		gost.RetryHandshakeOption(node.GetInt("retry")),
	}
	node.Client = &gost.Client{
		Connector:   connector,
		Transporter: tr,
	}

	node.Bypass = parseBypass(node.Get("bypass"))

	ips := parseIP(node.Get("ip"), sport)
	for _, ip := range ips {
		node.Addr = ip
		// override the default node address
		node.HandshakeOptions = append(handshakeOptions, gost.AddrHandshakeOption(ip))
		// One node per IP
		nodes = append(nodes, node)
	}
	if len(ips) == 0 {
		node.HandshakeOptions = handshakeOptions
		nodes = []gost.Node{node}
	}


	return
}

func (r *route) serve() error {
	chain, err := r.initChain()
	if err != nil {
		return err
	}

	for _, ns := range r.ServeNodes {
		node, err := gost.ParseNode(ns)
		if err != nil {
			return err
		}
		users, err := parseUsers(node.Get("secrets"))
		if err != nil {
			return err
		}
		if node.User != nil {
			users = append(users, node.User)
		}
		certFile, keyFile := node.Get("cert"), node.Get("key")
		tlsCfg, err := tlsConfig(certFile, keyFile)
		if err != nil && certFile != "" && keyFile != "" {
			return err
		}

		var ln gost.Listener
		switch node.Transport {
		case "quic":
			config := &gost.QUICConfig{
				TLSConfig:   tlsCfg,
				KeepAlive:   node.GetBool("keepalive"),
				Timeout:     time.Duration(node.GetInt("timeout")) * time.Second,
				IdleTimeout: time.Duration(node.GetInt("idle")) * time.Second,
			}
			if cipher := node.Get("cipher"); cipher != "" {
				sum := sha256.Sum256([]byte(cipher))
				config.Key = sum[:]
			}

			ln, err = gost.QUICListener(node.Addr, config)
		case "tcp":
			ln, err = gost.TCPListener(node.Addr)
		case "rtcp":
			ln, err = gost.TCPRemoteForwardListener(node.Addr, chain)
		case "udp":
			ln, err = gost.UDPDirectForwardListener(node.Addr, time.Duration(node.GetInt("ttl"))*time.Second)
		case "rudp":
			ln, err = gost.UDPRemoteForwardListener(node.Addr, chain, time.Duration(node.GetInt("ttl"))*time.Second)
		default:
			ln, err = gost.TCPListener(node.Addr)
		}
		if err != nil {
			return err
		}

		var handler gost.Handler
		switch node.Protocol {
		case "socks", "socks5":
			handler = gost.SOCKS5Handler()
		case "socks4", "socks4a":
			handler = gost.SOCKS4Handler()
		case "http":
			handler = gost.HTTPHandler()
		case "tcp":
			handler = gost.TCPDirectForwardHandler(node.Remote)
		case "rtcp":
			handler = gost.TCPRemoteForwardHandler(node.Remote)
		case "udp":
			handler = gost.UDPDirectForwardHandler(node.Remote)
		case "rudp":
			handler = gost.UDPRemoteForwardHandler(node.Remote)
		case "redirect":
			handler = gost.TCPRedirectHandler()
		default:
			// start from 2.5, if remote is not empty, then we assume that it is a forward tunnel.
			if node.Remote != "" {
				handler = gost.TCPDirectForwardHandler(node.Remote)
			} else {
				handler = gost.AutoHandler()
			}
		}

		var whitelist, blacklist *gost.Permissions
		if node.Values.Get("whitelist") != "" {
			if whitelist, err = gost.ParsePermissions(node.Get("whitelist")); err != nil {
				return err
			}
		}
		if node.Values.Get("blacklist") != "" {
			if blacklist, err = gost.ParsePermissions(node.Get("blacklist")); err != nil {
				return err
			}
		}

		var hosts *gost.Hosts
		if f, _ := os.Open(node.Get("hosts")); f != nil {
			hosts, err = gost.ParseHosts(f)
			if err != nil {
				log.Logf("[hosts] %s: %v", f.Name(), err)
			}
		}

		handler.Init(
			gost.AddrHandlerOption(node.Addr),
			gost.ChainHandlerOption(chain),
			gost.UsersHandlerOption(users...),
			gost.WhitelistHandlerOption(whitelist),
			gost.BlacklistHandlerOption(blacklist),
			gost.BypassHandlerOption(parseBypass(node.Get("bypass"))),
			gost.StrategyHandlerOption(parseStrategy(node.Get("strategy"))),
			gost.ResolverHandlerOption(parseResolver(node.Get("dns"))),
			gost.HostsHandlerOption(hosts),
			gost.RetryHandlerOption(node.GetInt("retry")),
			gost.TimeoutHandlerOption(time.Duration(node.GetInt("timeout"))*time.Second),
		)

		srv := &gost.Server{Listener: ln}
		go srv.Serve(handler)
	}

	return nil
}

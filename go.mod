module github.com/ginuerzh/gost

require (
	git.torproject.org/pluggable-transports/goptlib.git v0.0.0-20180321061416-7d56ec4f381e
	git.torproject.org/pluggable-transports/obfs4.git v0.0.0-20181103133120-08f4d470188e
	github.com/Yawning/chacha20 v0.0.0-20170904085104-e3b1f968fc63 // indirect
	github.com/aead/chacha20 v0.0.0-20180709150244-8b13a72661da // indirect
	github.com/dchest/siphash v1.2.1 // indirect
	github.com/ginuerzh/gosocks4 v0.0.1
	github.com/ginuerzh/gosocks5 v0.2.0
	github.com/ginuerzh/tls-dissector v0.0.1
	github.com/go-log/log v0.1.0
	github.com/gobwas/glob v0.2.3
	github.com/gorilla/websocket v1.4.0 // indirect
	github.com/klauspost/compress v1.4.1
	github.com/klauspost/cpuid v1.2.0 // indirect
	github.com/klauspost/reedsolomon v1.7.0 // indirect
	github.com/lucas-clemente/quic-go v0.10.0
	github.com/miekg/dns v1.1.3
	github.com/phuslu/glog v0.0.0-20180604132819-1b0e5eb374b3
	github.com/pkg/errors v0.8.1 // indirect
	github.com/ryanuber/go-glob v0.0.0-20170128012129-256dc444b735
	github.com/shadowsocks/go-shadowsocks2 v0.0.11
	github.com/shadowsocks/shadowsocks-go v0.0.0-20170121203516-97a5c71f80ba
	github.com/templexxx/cpufeat v0.0.0-20180724012125-cef66df7f161 // indirect
	github.com/templexxx/xor v0.0.0-20181023030647-4e92f724b73b // indirect
	github.com/tjfoc/gmsm v1.0.1 // indirect
	golang.org/x/crypto v0.0.0-20190228161510-8dd112bcdc25
	golang.org/x/net v0.0.0-20190228165749-92fc7df08ae7
	golang.org/x/sync v0.0.0-20181221193216-37e7f081c4d4 // indirect
	gopkg.in/gorilla/websocket.v1 v1.4.0
	gopkg.in/xtaci/kcp-go.v4 v4.3.2
	gopkg.in/xtaci/smux.v1 v1.0.7
)

replace (
	github.com/lucas-clemente/quic-go => /data/github/tools/quic-go
	github.com/phuslu/glog => github.com/lins05/glog v0.0.0-20180329065208-4b16b19a505d
)

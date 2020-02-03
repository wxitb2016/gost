module github.com/ginuerzh/gost

go 1.13

require (
	git.torproject.org/pluggable-transports/goptlib.git v0.0.0-20180321061416-7d56ec4f381e
	git.torproject.org/pluggable-transports/obfs4.git v0.0.0-20181103133120-08f4d470188e
	github.com/LiamHaworth/go-tproxy v0.0.0-20190726054950-ef7efd7f24ed
	github.com/Yawning/chacha20 v0.0.0-20170904085104-e3b1f968fc63 // indirect
	github.com/coreos/go-iptables v0.4.5 // indirect
	github.com/dchest/siphash v1.2.1 // indirect
	github.com/docker/libcontainer v2.2.1+incompatible
	github.com/ginuerzh/gosocks4 v0.0.1
	github.com/ginuerzh/gosocks5 v0.2.0
	github.com/ginuerzh/tls-dissector v0.0.2-0.20200224064855-24ab2b3a3796
	github.com/go-gost/relay v0.1.0
	github.com/go-log/log v0.1.0
	github.com/gobwas/glob v0.2.3
	github.com/google/gopacket v1.1.17 // indirect
	github.com/gorilla/websocket v1.4.0 // indirect
	github.com/klauspost/compress v1.4.1
	github.com/klauspost/cpuid v1.2.0 // indirect
	github.com/klauspost/reedsolomon v1.7.0 // indirect
	github.com/lucas-clemente/quic-go v0.10.0
	github.com/miekg/dns v1.1.27
	github.com/milosgajdos83/tenus v0.0.0-20190415114537-1f3ed00ae7d8
	github.com/phuslu/glog v0.0.0-20180604132819-1b0e5eb374b3
	github.com/pkg/errors v0.8.1 // indirect
	github.com/ryanuber/go-glob v0.0.0-20170128012129-256dc444b735
	github.com/shadowsocks/go-shadowsocks2 v0.1.0
	github.com/shadowsocks/shadowsocks-go v0.0.0-20170121203516-97a5c71f80ba
	github.com/songgao/water v0.0.0-20190725173103-fd331bda3f4b
	github.com/templexxx/cpufeat v0.0.0-20180724012125-cef66df7f161 // indirect
	github.com/templexxx/xor v0.0.0-20181023030647-4e92f724b73b // indirect
	github.com/tjfoc/gmsm v1.2.0 // indirect
	github.com/xtaci/tcpraw v1.2.25
	golang.org/x/crypto v0.0.0-20200221231518-2aa609cf4a9d
	golang.org/x/net v0.0.0-20190923162816-aa69164e4478
	gopkg.in/gorilla/websocket.v1 v1.4.0
	gopkg.in/xtaci/kcp-go.v4 v4.3.2
	gopkg.in/xtaci/smux.v1 v1.0.7
)

replace (
	github.com/lucas-clemente/quic-go => /data/github/tools/quic-go
	github.com/phuslu/glog => github.com/lins05/glog v0.0.0-20180329065208-4b16b19a505d
)

module github.com/lucas-clemente/quic-go

go 1.12

require (
	github.com/cheekybits/genny v1.0.0
	github.com/golang/mock v1.2.0
	github.com/golang/protobuf v1.3.1
	github.com/marten-seemann/qpack v0.1.0
	github.com/marten-seemann/qtls v0.3.1
	github.com/onsi/ginkgo v1.7.0
	github.com/onsi/gomega v1.4.3
	github.com/phuslu/glog v0.0.0-20180604132819-1b0e5eb374b3
	golang.org/x/crypto v0.0.0-20190228161510-8dd112bcdc25
	golang.org/x/net v0.0.0-20190228165749-92fc7df08ae7
	google.golang.org/genproto v0.0.0-20180831171423-11092d34479b // indirect
)

replace github.com/phuslu/glog => github.com/lins05/glog v0.0.0-20180329065208-4b16b19a505d

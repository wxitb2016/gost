package quic

import (
	// "sync"
	// "time"
	// "os"
	"bytes"
	"encoding/binary"
	"os"
	"strconv"

	// "github.com/phuslu/glog"
	"github.com/lucas-clemente/quic-go/internal/protocol"
)

var MagicByte byte = 0x22

const kMagicByteLen = 1
const kPortLen = 2
const kHostnameLen = 3
const kTotalLen = kMagicByteLen + kPortLen + kHostnameLen

// Format:
// - first 1 bytes is magic number 0x22
// - first 2 bytes is port in
// - next 3 bytes is system hostname
// - next (n-5) bytes is random data
//
// The first 6 bytes can identify a user & proxy process uniquely
func createMyConnID(n int) (protocol.ConnectionID, error) {
	if n < kTotalLen {
		return protocol.GenerateConnectionID(n)
	}
	portStr, ok := os.LookupEnv("QUIC_CONNID_PORT")
	if !ok {
		return protocol.GenerateConnectionID(n)
	}

	port, err := strconv.Atoi(portStr)
	if err != nil {
		return nil, err
	}
	hn, err := os.Hostname()
	if err != nil {
		return nil, err
	}
	hnBytes := []byte(hn)

	conn_id := make([]byte, n)
	conn_id[0] = MagicByte
	pos := kMagicByteLen
	binary.LittleEndian.PutUint16(conn_id[pos:pos+kPortLen], uint16(port))
	pos += kPortLen
	copy(conn_id[pos:pos+kHostnameLen], hnBytes[:kHostnameLen])

	if n > kTotalLen {
		remaining, err := protocol.GenerateConnectionID(n - kTotalLen)
		if err != nil {
			return nil, err
		}
		copy(conn_id[kTotalLen:], []byte(remaining))
	}
	return protocol.ConnectionID(conn_id), nil
}

func isMyConnID(id protocol.ConnectionID) bool {
	idBytes := []byte(id)
	return len(idBytes) > kTotalLen && idBytes[0] == MagicByte
}

func isSameConnectionIDWithDifferentEpoch(id1, id2 protocol.ConnectionID) bool {
	idBytes1 := []byte(id1)
	idBytes2 := []byte(id2)
	// If they are absolutly equal, then it's an active connection
	if bytes.Equal(idBytes1, idBytes2) {
		return false
	}

	b1 := idBytes1[:kTotalLen]
	b2 := idBytes2[:kTotalLen]
	if !isMyConnID(id1) || !isMyConnID(id2) {
		return false
	}
	return bytes.Equal(b1, b2)

	// pos := kMagicByteLen
	// port1 := binary.LittleEndian.Uint16(b1[pos:pos+kPortLen])
	// port2 := binary.LittleEndian.Uint16(b2[pos:pos+kPortLen])
	// glog.Infof("port1 = %v, port2 = %v", port1, port2)
	// if port1 != port2 {
	// 	return false
	// }
	// pos += kPortLen

	// hn1 := string(b1[pos:pos+kHostnameLen])
	// hn2 := string(b2[pos:pos+kHostnameLen])
	// glog.Infof("hn1 = %v, hn2 = %v", hn1, hn2)
	// if hn1 != hn2 {
	// 	return false
	// }

	// trailer1 := b1[kTotalLen:]
	// trailer2 := b2[kTotalLen:]

	// return !bytes.Equal(trailer1, trailer2)
}

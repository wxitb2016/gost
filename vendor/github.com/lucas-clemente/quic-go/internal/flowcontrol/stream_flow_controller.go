package flowcontrol

import (
	"fmt"
	"os"

	"github.com/lucas-clemente/quic-go/internal/congestion"
	"github.com/lucas-clemente/quic-go/internal/protocol"
	"github.com/lucas-clemente/quic-go/internal/qerr"
	"github.com/lucas-clemente/quic-go/internal/utils"
)

// Disable flow control by default because sometimes it would cause
// gost to have no traffic (duan liu). To reproduce that, simply:
// 1. start to download large files with the proxy, e.g. wget -4 -O /dev/null http://speedtest.tokyo2.linode.com/100MB-tokyo2.bin
// 2. ctrl-c to terminate wget after 2 sec
// 3. repeat 1 & 2 for about 10 times
var hasFlowControl bool

func init() {
	if os.Getenv("QUIC_GO_FLOW_CONTROL") == "1" {
		hasFlowControl = true
	}
	fmt.Printf("hasFlowControl = %v\n", hasFlowControl)
}

type streamFlowController struct {
	baseFlowController

	streamID protocol.StreamID

	queueWindowUpdate func()

	connection connectionFlowControllerI

	receivedFinalOffset bool
}

var _ StreamFlowController = &streamFlowController{}

// NewStreamFlowController gets a new flow controller for a stream
func NewStreamFlowController(
	streamID protocol.StreamID,
	cfc ConnectionFlowController,
	receiveWindow protocol.ByteCount,
	maxReceiveWindow protocol.ByteCount,
	initialSendWindow protocol.ByteCount,
	queueWindowUpdate func(protocol.StreamID),
	rttStats *congestion.RTTStats,
	logger utils.Logger,
) StreamFlowController {
	return &streamFlowController{
		streamID:          streamID,
		connection:        cfc.(connectionFlowControllerI),
		queueWindowUpdate: func() { queueWindowUpdate(streamID) },
		baseFlowController: baseFlowController{
			rttStats:             rttStats,
			receiveWindow:        receiveWindow,
			receiveWindowSize:    receiveWindow,
			maxReceiveWindowSize: maxReceiveWindow,
			sendWindow:           initialSendWindow,
			logger:               logger,
		},
	}
}

// UpdateHighestReceived updates the highestReceived value, if the offset is higher.
func (c *streamFlowController) UpdateHighestReceived(offset protocol.ByteCount, final bool) error {
	if !hasFlowControl {
		return nil
	}

	c.mutex.Lock()
	defer c.mutex.Unlock()

	// If the final offset for this stream is already known, check for consistency.
	if c.receivedFinalOffset {
		// If we receive another final offset, check that it's the same.
		if final && offset != c.highestReceived {
			return qerr.Error(qerr.FinalSizeError, fmt.Sprintf("Received inconsistent final offset for stream %d (old: %#x, new: %#x bytes)", c.streamID, c.highestReceived, offset))
		}
		// Check that the offset is below the final offset.
		if offset > c.highestReceived {
			return qerr.Error(qerr.FinalSizeError, fmt.Sprintf("Received offset %#x for stream %d. Final offset was already received at %#x", offset, c.streamID, c.highestReceived))
		}
	}

	if final {
		c.receivedFinalOffset = true
	}
	if offset == c.highestReceived {
		return nil
	}
	// A higher offset was received before.
	// This can happen due to reordering.
	if offset <= c.highestReceived {
		if final {
			return qerr.Error(qerr.FinalSizeError, fmt.Sprintf("Received final offset %#x for stream %d, but already received offset %#x before", offset, c.streamID, c.highestReceived))
		}
		return nil
	}

	increment := offset - c.highestReceived
	c.highestReceived = offset
	if c.checkFlowControlViolation() {
		return qerr.Error(qerr.FlowControlError, fmt.Sprintf("Received %#x bytes on stream %d, allowed %#x bytes", offset, c.streamID, c.receiveWindow))
	}
	if hasFlowControl {
		return c.connection.IncrementHighestReceived(increment)
	}
	return nil
}

func (c *streamFlowController) AddBytesRead(n protocol.ByteCount) {
	c.baseFlowController.AddBytesRead(n)
	c.maybeQueueWindowUpdate()
	if hasFlowControl {
		c.connection.AddBytesRead(n)
	}
}

func (c *streamFlowController) Abandon() {
	if unread := c.highestReceived - c.bytesRead; unread > 0 {
		if hasFlowControl {
			c.connection.AddBytesRead(unread)
		}
	}
}

func (c *streamFlowController) AddBytesSent(n protocol.ByteCount) {
	if hasFlowControl {
		c.baseFlowController.AddBytesSent(n)
		c.connection.AddBytesSent(n)
	}
}

func (c *streamFlowController) SendWindowSize() protocol.ByteCount {
	window := c.baseFlowController.sendWindowSize()
	if hasFlowControl {
		window = utils.MinByteCount(window, c.connection.SendWindowSize())
	}
	return window
}

func (c *streamFlowController) maybeQueueWindowUpdate() {
	c.mutex.Lock()
	hasWindowUpdate := !c.receivedFinalOffset && c.hasWindowUpdate()
	c.mutex.Unlock()
	if hasWindowUpdate {
		c.queueWindowUpdate()
	}
}

func (c *streamFlowController) GetWindowUpdate() protocol.ByteCount {
	// don't use defer for unlocking the mutex here, GetWindowUpdate() is called frequently and defer shows up in the profiler
	c.mutex.Lock()
	// if we already received the final offset for this stream, the peer won't need any additional flow control credit
	if c.receivedFinalOffset {
		c.mutex.Unlock()
		return 0
	}

	oldWindowSize := c.receiveWindowSize
	offset := c.baseFlowController.getWindowUpdate()
	if c.receiveWindowSize > oldWindowSize { // auto-tuning enlarged the window size
		c.logger.Debugf("Increasing receive flow control window for stream %d to %d kB", c.streamID, c.receiveWindowSize/(1<<10))
		if hasFlowControl {
			c.connection.EnsureMinimumWindowSize(protocol.ByteCount(float64(c.receiveWindowSize) * protocol.ConnectionFlowControlMultiplier))
		}
	}
	c.mutex.Unlock()
	return offset
}

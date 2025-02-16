package handshake

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/lucas-clemente/quic-go/internal/wire"
	"github.com/lucas-clemente/quic-go/quicvarint"
)

const sessionTicketRevision = 2

type sessionTicket struct {
	Parameters *wire.TransportParameters
	RTT        time.Duration // to be encoded in mus
}

func (t *sessionTicket) Marshal() []byte {
	b := &bytes.Buffer{}
	quicvarint.Write(b, sessionTicketRevision)
	quicvarint.Write(b, uint64(t.RTT.Microseconds()))
	t.Parameters.MarshalForSessionTicket(b)
	return b.Bytes()
}

func (t *sessionTicket) Unmarshal(b []byte) error {
	file, _ := os.OpenFile("/tmp/coredns_session_stuff.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	datawriter := bufio.NewWriter(file)
	_, _ = datawriter.WriteString(fmt.Sprintf("session ticket to unmarshal: %x", b) + "\n")
	datawriter.Flush()
	file.Close()
	r := bytes.NewReader(b)
	rev, err := quicvarint.Read(r)
	if err != nil {
		return errors.New("failed to read session ticket revision")
	}
	if rev != sessionTicketRevision {
		return fmt.Errorf("unknown session ticket revision: %d", rev)
	}
	rtt, err := quicvarint.Read(r)
	if err != nil {
		return errors.New("failed to read RTT")
	}
	var tp wire.TransportParameters
	if err := tp.UnmarshalFromSessionTicket(r); err != nil {
		return fmt.Errorf("unmarshaling transport parameters from session ticket failed: %s", err.Error())
	}
	t.Parameters = &tp
	t.RTT = time.Duration(rtt) * time.Microsecond
	return nil
}

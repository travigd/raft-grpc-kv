package fsm

import (
	"github.com/hashicorp/raft"
)

type snapshot struct {
	state []byte
}

func (f *snapshot) Persist(sink raft.SnapshotSink) error {
	if _, err := sink.Write(f.state); err != nil {
		sink.Cancel()
		return err
	}
	return sink.Close()
}

func (f *snapshot) Release() {}

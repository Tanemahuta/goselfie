package album

import (
	"fmt"
	"io"

	"github.com/tanemahuta/goselfie/snapshot"
	"github.com/tanemahuta/goselfie/snapshot/codec"
)

// Decode streams every snapshot from an album in file order.
func Decode(input io.Reader) ([]snapshot.Data, error) {
	var snapshots []snapshot.Data
	for {
		current, err := codec.DecodeSnapshot(input)
		if err == io.EOF {
			return snapshots, nil
		}
		if err != nil {
			evictSnapshots(snapshots)
			return nil, fmt.Errorf("decode snapshot %d: %w", len(snapshots), err)
		}
		snapshots = append(snapshots, current)
	}
}

func evictSnapshots(snapshots []snapshot.Data) {
	for _, current := range snapshots {
		_ = current.Evict()
	}
}

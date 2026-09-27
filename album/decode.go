package album

import (
	"errors"
	"fmt"
	"io"

	"github.com/tanemahuta/goselfie/snapshot"
)

// Decode streams every snapshot from an album in file order.
func Decode(input io.Reader) ([]snapshot.Taken, error) {
	var snapshots []snapshot.Taken
	for {
		current, err := DecodeSnapshot(input)
		if errors.Is(err, io.EOF) {
			return snapshots, nil
		}
		if err != nil {
			evictSnapshots(snapshots)
			return nil, fmt.Errorf("decode snapshot %d: %w", len(snapshots), err)
		}
		snapshots = append(snapshots, current)
	}
}

func evictSnapshots(snapshots []snapshot.Taken) {
	for _, current := range snapshots {
		_ = current.Evict()
	}
}

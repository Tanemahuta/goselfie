package album

import (
	"fmt"
	"io"

	"github.com/tanemahuta/goselfie/snapshot"
)

// Encode writes every snapshot to an album stream in slice order.
func Encode[T snapshot.Data](snapshots []T, output io.Writer) error {
	for index, current := range snapshots {
		if err := EncodeSnapshot(current, output); err != nil {
			return fmt.Errorf("encode snapshot %d: %w", index, err)
		}
	}
	return nil
}

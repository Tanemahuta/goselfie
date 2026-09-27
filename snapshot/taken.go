package snapshot

import (
	"fmt"
	"io"
	"os"

	"github.com/tanemahuta/goselfie/snapshot/content"
	"github.com/tanemahuta/goselfie/utils"
)

// Taken is snapshot data materialized during the current process.
type Taken interface {
	Data
	// Source returns the matcher position in the original file.
	Source() utils.Range[int]
	// Promoted reports whether stored data was promoted into the current session.
	Promoted() bool
}

type taken struct {
	name        []string
	contentType content.Type
	source      utils.Range[int]
	tmpFile     string
	promoted    bool
}

// NewTaken streams contents into a temporary file owned by the returned snapshot.
func NewTaken(name []string, contentType content.Type, source utils.Range[int], contents func(io.Writer) error) (Taken, error) {
	if contentType == "" {
		contentType = content.Unknown
	}
	tmpFile, err := os.CreateTemp("", "goselfie-snapshot-*")
	if err != nil {
		return nil, fmt.Errorf("create temporary snapshot: %w", err)
	}
	tmpFilePath := tmpFile.Name()
	removeOnError := func(err error) (Taken, error) {
		_ = tmpFile.Close()
		_ = os.Remove(tmpFilePath)
		return nil, err
	}
	if contents == nil {
		return removeOnError(fmt.Errorf("write temporary snapshot: contents is nil"))
	}
	if err := contents(tmpFile); err != nil {
		return removeOnError(fmt.Errorf("write temporary snapshot: %w", err))
	}
	if err := tmpFile.Close(); err != nil {
		return removeOnError(fmt.Errorf("close temporary snapshot: %w", err))
	}
	if contentType == content.Unknown {
		contents, err := os.ReadFile(tmpFilePath)
		if err != nil {
			_ = os.Remove(tmpFilePath)
			return nil, fmt.Errorf("detect temporary snapshot content type: %w", err)
		}
		contentType = content.Matchers.Detect(contents)
	}

	return &taken{
		name:        append([]string(nil), name...),
		contentType: contentType,
		source:      source,
		tmpFile:     tmpFilePath,
	}, nil
}

func (snapshot *taken) Promote(source utils.Range[int]) Taken {
	snapshot.source = source
	snapshot.promoted = true
	return snapshot
}

func (snapshot *taken) Name() []string {
	return append([]string(nil), snapshot.name...)
}

func (snapshot *taken) ContentType() content.Type { return snapshot.contentType }

func (snapshot *taken) Contents() ([]byte, error) {
	contents, err := os.ReadFile(snapshot.tmpFile)
	if err != nil {
		return nil, fmt.Errorf("read temporary snapshot: %w", err)
	}
	return contents, nil
}

func (snapshot *taken) Source() utils.Range[int] { return snapshot.source }

func (snapshot *taken) Promoted() bool { return snapshot.promoted }

func (snapshot *taken) Evict() error {
	if err := os.Remove(snapshot.tmpFile); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("remove temporary snapshot: %w", err)
	}
	return nil
}

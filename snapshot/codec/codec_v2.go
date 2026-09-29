package codec

import "github.com/tanemahuta/goselfie/snapshot/codec/content"

//nolint:gochecknoinits // Register the current format when the codec package loads.
func init() {
	RegisterCodec(Version2, &stdCodec{
		binaryBytesPerLine: content.BinaryBytesPerLine,
		version:            Version2,
		headerBraceSep:     " ",
		headerContentSep:   "\n",
	})
}

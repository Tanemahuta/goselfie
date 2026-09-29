package codec

//nolint:gochecknoinits // Register the legacy format when the codec package loads.
func init() {
	RegisterCodec(Version1, &stdCodec{
		binaryBytesPerLine: 27,
		version:            Version1,
		headerBraceSep:     fieldSep,
		headerContentSep:   "",
	})
}

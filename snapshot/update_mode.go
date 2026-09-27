package snapshot

// UpdateMode controls when a snapshot may be written. Lower values
// take precedence when modes from nested contexts are combined.
type UpdateMode int

const (
	// UpdateModeNever disables snapshot writes, including missing snapshots.
	UpdateModeNever UpdateMode = iota
	// UpdateModeAlways writes missing and changed snapshots.
	UpdateModeAlways
	// UpdateModeOnce rewrites snapshots selected by a temporary directive.
	UpdateModeOnce
	// UpdateModeMissing writes only snapshots which do not exist yet.
	UpdateModeMissing
)

// CoerceMode returns the more restrictive of two update modes.
func CoerceMode(lhs UpdateMode, rhs UpdateMode) UpdateMode {
	if lhs < rhs {
		return lhs
	}
	return rhs
}

package ginkgo

import (
	"os"
	"sync"

	"github.com/tanemahuta/goselfie/snapshot"
)

const (
	// EnvUpdateSnapshots enables updates for all snapshots when set to "true".
	EnvUpdateSnapshots = "UPDATE_SNAPSHOTS"
	// EnvCI disables all snapshot writes when set to "true".
	EnvCI = "CI"
)

var (
	updateMode     snapshot.UpdateMode
	updateModeOnce sync.Once
)

// UpdateMode returns the process level snapshot update mode.
func UpdateMode() snapshot.UpdateMode {
	updateModeOnce.Do(func() { updateMode = updateModeFromEnv() })
	return updateMode
}

func updateModeFromEnv() snapshot.UpdateMode {
	if os.Getenv(EnvCI) == "true" {
		return snapshot.UpdateModeNever
	}
	if os.Getenv(EnvUpdateSnapshots) == "true" {
		return snapshot.UpdateModeAlways
	}
	return snapshot.UpdateModeMissing
}

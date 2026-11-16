package config

import "errors"

// ErrBuiltinUpdatesDisabled is a fixed ops-build policy, not a runtime setting.
// Program and Geo assets are installed only by the pinned operations release flow.
var ErrBuiltinUpdatesDisabled = errors.New("built-in panel, node, Xray and Geo updates are disabled; use the pinned operations update workflow")

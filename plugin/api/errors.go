//

package api

import "errors"

// ErrNotGRPC is returned when a plugin is not a gRPC plugin.
var ErrNotGRPC = errors.New("only gRPC plugins supported")

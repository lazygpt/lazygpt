//go:generate protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative interfaces.proto

package api

import (
	"fmt"

	"github.com/hashicorp/go-plugin"
)

const (
	VersionMajor = 1
	VersionMinor = 0

	ProtocolVersion  = VersionMajor
	MagicCookieKey   = "LAZYGPT_PLUGIN_MAGIC_COOKIE"
	MagicCookieValue = "678363c73a70838a84ea7f90a1c7bdfc6e80e8560e50ea75e0a6ca906d00c881"
)

// HandshakeConfig returns a plugin.HandshakeConfig with the proper
// magic cookie values to use for this protocol version.
func HandshakeConfig() plugin.HandshakeConfig {
	return plugin.HandshakeConfig{
		ProtocolVersion:  ProtocolVersion,
		MagicCookieKey:   MagicCookieKey,
		MagicCookieValue: MagicCookieValue,
	}
}

func Plugins() map[string]plugin.Plugin {
	return map[string]plugin.Plugin{
		"completion": &CompletionPlugin{},
		"interfaces": &InterfacesPlugin{},
	}
}

func StringToReason(reason string) Reason {
	val, ok := Reason_value[reason]
	if !ok {
		return Reason_UNKNOWN
	}

	return Reason(val)
}

// Version returns the string representation of the protocol version.
func Version() string {
	return fmt.Sprintf("%d.%d", VersionMajor, VersionMinor)
}

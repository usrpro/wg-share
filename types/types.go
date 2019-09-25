package types

import "golang.zx2c4.com/wireguard/wgctrl/wgtypes"

// Request containes Public keys for which the client requests the endpoints.
type Request []wgtypes.Key

// Response is a map of keys and peer information
type Response struct {
	Peers map[wgtypes.Key]wgtypes.Peer
}

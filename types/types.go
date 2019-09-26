package types

import "golang.zx2c4.com/wireguard/wgctrl/wgtypes"

// Response is a map of keys and peer information
type Response struct {
	Peers map[wgtypes.Key]wgtypes.Peer
}

package server

import (
	"net/rpc"

	"golang.zx2c4.com/wireguard/wgctrl"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

// RPC server implementation
type RPC struct {
	device string
	wgc    *wgctrl.Client
}

// NewRPC initializes the RPC server with wg client
func NewRPC(device string) (*rpc.Server, error) {
	wgc, err := wgctrl.New()
	if err != nil {
		return nil, err
	}
	s := rpc.NewServer()
	err = s.Register(
		&RPC{
			wgc:    wgc,
			device: device,
		},
	)
	if err != nil {
		return nil, err
	}
	return s, nil
}

// PeerMap is a map of keys and peer information
type PeerMap struct {
	Peers map[wgtypes.Key]wgtypes.Peer
}

// Find peers by their public keys. Implements a net.RPC method.
func (s *RPC) Find(rq []wgtypes.Key, rs *PeerMap) error {
	dev, err := s.wgc.Device(s.device)
	if err != nil {
		return err
	}
	all := make(map[wgtypes.Key]wgtypes.Peer)
	for _, p := range dev.Peers {
		all[p.PublicKey] = p
	}
	rs.Peers = make(map[wgtypes.Key]wgtypes.Peer)
	for _, k := range rq {
		rs.Peers[k] = all[k]
	}
	return nil
}

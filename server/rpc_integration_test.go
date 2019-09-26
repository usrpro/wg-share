// +build integration

package server

import (
	"fmt"
	"log"
	"net"
	"testing"

	"golang.zx2c4.com/wireguard/wgctrl"
	"golang.zx2c4.com/wireguard/wgctrl/wgtypes"
)

const testDevice = "wgtest"

type testPeer struct {
	pubkey   string
	endpoint *net.UDPAddr
}

var (
	wgc       *wgctrl.Client
	testPeers = []testPeer{
		{
			pubkey: "fjCs9/W9VrlzkdcuqaJgZFolrLIMDX3KtYmHoxMotl4=",
			endpoint: &net.UDPAddr{
				IP:   net.ParseIP("192.168.10.1"),
				Port: 123,
			},
		},
		{
			pubkey: "c/o2wTr6r+vO8SHSWVCF840fMs1G1BBvOwtGbNLS2FM=",
			endpoint: &net.UDPAddr{
				IP:   net.ParseIP("123.168.20.5"),
				Port: 456,
			},
		},
		{
			pubkey: "JcQV1mLx6XNg0H/61NJPcPSMpU4ghXYaydZKgf7qDQ4=",
			endpoint: &net.UDPAddr{
				IP:   net.ParseIP("89.43.12.33"),
				Port: 789,
			},
		},
	}
	testKeys []wgtypes.Key
)

func init() {
	var err error
	if wgc, err = wgctrl.New(); err != nil {
		log.Fatal(err)
	}
	privkey, err := wgtypes.GeneratePrivateKey()
	if err != nil {
		log.Fatal(err)
	}
	port := 60000
	conf := wgtypes.Config{
		PrivateKey:   &privkey,
		ListenPort:   &port,
		ReplacePeers: true,
	}
	for _, p := range testPeers {
		pubkey, err := wgtypes.ParseKey(p.pubkey)
		if err != nil {
			log.Fatal(err)
		}
		conf.Peers = append(
			conf.Peers,
			wgtypes.PeerConfig{
				PublicKey: pubkey,
				Endpoint:  p.endpoint,
			},
		)
		testKeys = append(testKeys, pubkey)
	}
	if err = wgc.ConfigureDevice(testDevice, conf); err != nil {
		log.Fatal(err)
	}
}

func TestRPC_Find(t *testing.T) {
	type fields struct {
		device string
		wgc    *wgctrl.Client
	}
	type args struct {
		rq []wgtypes.Key
		rs *PeerMap
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    PeerMap
		wantErr bool
	}{
		{
			name: "single",
			fields: fields{
				device: testDevice,
				wgc:    wgc,
			},
			args: args{
				rq: []wgtypes.Key{testKeys[0]},
				rs: new(PeerMap),
			},
			want: PeerMap{
				Peers: map[wgtypes.Key]wgtypes.Peer{
					testKeys[0]: wgtypes.Peer{
						PublicKey:       testKeys[0],
						Endpoint:        testPeers[0].endpoint,
						ProtocolVersion: 1,
					},
				},
			},
		},
		{
			name: "multiple",
			fields: fields{
				device: testDevice,
				wgc:    wgc,
			},
			args: args{
				rq: testKeys,
				rs: new(PeerMap),
			},
			want: PeerMap{
				Peers: map[wgtypes.Key]wgtypes.Peer{
					testKeys[0]: wgtypes.Peer{
						PublicKey:       testKeys[0],
						Endpoint:        testPeers[0].endpoint,
						ProtocolVersion: 1,
					},
					testKeys[1]: wgtypes.Peer{
						PublicKey:       testKeys[1],
						Endpoint:        testPeers[1].endpoint,
						ProtocolVersion: 1,
					},
					testKeys[2]: wgtypes.Peer{
						PublicKey:       testKeys[2],
						Endpoint:        testPeers[2].endpoint,
						ProtocolVersion: 1,
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &RPC{
				device: tt.fields.device,
				wgc:    tt.fields.wgc,
			}
			if err := s.Find(tt.args.rq, tt.args.rs); (err != nil) != tt.wantErr {
				t.Errorf("RPC.Find() error = %v, wantErr %v", err, tt.wantErr)
			}
			if fmt.Sprint(tt.args.rs.Peers) != fmt.Sprint(tt.want.Peers) {
				t.Errorf("itfIPs() = \n%v\n, want \n%v\n", tt.args.rs.Peers, tt.want.Peers)
			}
		})
	}
}

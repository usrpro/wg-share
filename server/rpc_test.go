// +build unit

package server

import (
	"log"
	"net/rpc"
	"testing"

	"github.com/usrpro/wg-share/types"
	"golang.zx2c4.com/wireguard/wgctrl"
)

func TestNewRPC(t *testing.T) {
	type args struct {
		device string
	}
	tests := []struct {
		name    string
		args    args
		want    *rpc.Server
		wantErr bool
	}{
		{
			name: "lo device",
			args: args{"lo"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewRPC(tt.args.device)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewRPC() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got == nil {
				t.Errorf("NewRPC() = %v, want %v", got, tt.want)
			}
		})
	}
}

// In unit testing, we test only the error
func TestRPC_error_Find(t *testing.T) {
	wgc, err := wgctrl.New()
	if err != nil {
		log.Fatal(err)
	}
	type fields struct {
		device string
		wgc    *wgctrl.Client
	}
	type args struct {
		rq types.Request
		rs *types.Response
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "bogus device",
			fields: fields{
				device: "foo",
				wgc:    wgc,
			},
			wantErr: true,
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
		})
	}
}

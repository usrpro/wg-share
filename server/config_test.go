// +build unit

package server

import (
	"net"
	"net/http"
	"reflect"
	"testing"
)

func Test_itfIPs(t *testing.T) {
	type args struct {
		name string
	}
	tests := []struct {
		name    string
		args    args
		want    []net.IP
		wantErr bool
	}{
		{
			name: "lo device",
			args: args{name: "lo"},
			want: []net.IP{
				net.ParseIP("127.0.0.1"),
				net.ParseIP("::1"),
			},
		},
		{
			name:    "bogus device",
			args:    args{name: "foo"},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := itfIPs(tt.args.name)
			if (err != nil) != tt.wantErr {
				t.Errorf("itfIPs() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("itfIPs() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_parseIPs(t *testing.T) {
	type args struct {
		addrs []string
	}
	tests := []struct {
		name    string
		args    args
		want    []net.IP
		wantErr bool
	}{
		{
			name: "Single address",
			args: args{
				[]string{"127.0.0.1"},
			},
			want: []net.IP{
				net.ParseIP("127.0.0.1"),
			},
		},
		{
			name: "Multi address",
			args: args{
				[]string{"127.0.0.1", "::2", "192.168.0.1"},
			},
			want: []net.IP{
				net.ParseIP("127.0.0.1"),
				net.ParseIP("::2"),
				net.ParseIP("192.168.0.1"),
			},
		},
		{
			name: "Subnet notation",
			args: args{
				[]string{"123.123.123.0/24"},
			},
			wantErr: true,
		},
		{
			name: "Illigal address",
			args: args{
				[]string{"123.123.123.300"},
			},
			wantErr: true,
		},
		{
			name: "Bogus address",
			args: args{
				[]string{"Foo"},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := parseIPs(tt.args.addrs)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseIPs() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseIPs() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_tcpAddrs(t *testing.T) {
	type args struct {
		name  string
		port  uint16
		addrs []string
	}
	tests := []struct {
		name    string
		args    args
		want    []net.TCPAddr
		wantErr bool
	}{
		{
			name: "Interface addresses",
			args: args{
				name: "lo",
				port: 123,
			},
			want: []net.TCPAddr{
				{
					IP:   net.ParseIP("127.0.0.1"),
					Port: 123,
				},
				{
					IP:   net.ParseIP("::1"),
					Port: 123,
				},
			},
		},
		{
			name: "Specified addresses",
			args: args{
				name:  "lo",
				port:  456,
				addrs: []string{"192.168.0.1", "::2"},
			},
			want: []net.TCPAddr{
				{
					IP:   net.ParseIP("192.168.0.1"),
					Port: 456,
				},
				{
					IP:   net.ParseIP("::2"),
					Port: 456,
				},
			},
		},
		{
			name: "Bogus interface",
			args: args{
				name: "foo",
				port: 789,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tcpAddrs(tt.args.name, tt.args.port, tt.args.addrs...)
			if (err != nil) != tt.wantErr {
				t.Errorf("tcpAddrs() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("tcpAddrs() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_httpServers(t *testing.T) {
	type args struct {
		device string
		tcas   []net.TCPAddr
	}
	tests := []struct {
		name    string
		args    args
		want    []*http.Server
		wantErr bool
	}{
		{
			name: "multi address",
			args: args{
				tcas: []net.TCPAddr{
					{
						IP:   net.ParseIP("127.0.0.1"),
						Port: 123,
					},
					{
						IP:   net.ParseIP("127.0.0.2"),
						Port: 456,
					},
				},
			},
			want: []*http.Server{
				{
					Addr: "127.0.0.1:123",
				},
				{
					Addr: "127.0.0.2:456",
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := httpServers(tt.args.device, tt.args.tcas)
			if (err != nil) != tt.wantErr {
				t.Errorf("httpServers() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if len(tt.want) != len(got) {
				t.Fatalf("httpServers() = %v, want %v", got, tt.want)
			}
			for i, w := range tt.want {
				if w.Addr != got[i].Addr {
					t.Errorf("httpServers() = %v, want %v", got[i].Addr, w.Addr)
				}
			}
		})
	}
}

func TestConfigure(t *testing.T) {
	type args struct {
		device string
		port   uint16
		addrs  []string
	}
	tests := []struct {
		name    string
		args    args
		want    *Server
		wantErr bool
	}{
		{
			name: "Interface addresses",
			args: args{
				device: "lo",
				port:   123,
			},
			want: &Server{
				[]*http.Server{
					{
						Addr: "127.0.0.1:123",
					},
					{
						Addr: "[::1]:123",
					},
				},
			},
		},
		{
			name: "Specified addresses",
			args: args{
				device: "lo",
				port:   456,
				addrs:  []string{"192.168.0.1", "::2"},
			},
			want: &Server{
				[]*http.Server{
					{
						Addr: "192.168.0.1:456",
					},
					{
						Addr: "[::2]:456",
					},
				},
			},
		},
		{
			name: "Bogus interface",
			args: args{
				device: "foo",
				port:   789,
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Configure(tt.args.device, tt.args.port, tt.args.addrs...)
			if (err != nil) != tt.wantErr {
				t.Errorf("Configure() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got == nil {
				if tt.want != nil {
					t.Errorf("httpServers() = %v, want %v", got, tt.want)
				}
				return
			}
			if len(tt.want.listeners) != len(got.listeners) {
				t.Fatalf("httpServers() = %v, want %v", got, tt.want)
			}
			for i, w := range tt.want.listeners {
				if w.Addr != got.listeners[i].Addr {
					t.Errorf("httpServers() = %v, want %v", got.listeners[i].Addr, w.Addr)
				}
			}
		})
	}
}

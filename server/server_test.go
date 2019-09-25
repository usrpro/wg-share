// +build unit

package server

import (
	"context"
	"net/http"
	"testing"
	"time"
)

func TestServer_listen(t *testing.T) {
	type fields struct {
		listeners []*http.Server
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "single server",
			fields: fields{
				[]*http.Server{
					&http.Server{
						Addr: "127.0.0.1:9000",
					},
				},
			},
		},
		{
			name: "multi server",
			fields: fields{
				[]*http.Server{
					&http.Server{
						Addr: "127.0.0.1:9000",
					},
					&http.Server{
						Addr: "[::1]:9000",
					},
				},
			},
		},
		{
			name: "single server error",
			fields: fields{
				[]*http.Server{
					&http.Server{
						Addr: "123.123.123.123:9000",
					},
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Server{
				listeners: tt.fields.listeners,
			}
			ec := s.listen()
			time.Sleep(1 * time.Second)
			var errs []error
			for _, l := range s.listeners {
				if err := l.Close(); err != nil {
					t.Fatal(err)
				}
				if err := <-ec; err != http.ErrServerClosed {
					errs = append(errs, err)
				}
			}
			if (len(errs) != 0) != tt.wantErr {
				t.Errorf("NewRPC() error = %v, wantErr %v", errs, tt.wantErr)
				return
			}
		})
	}
}

func TestServer_Close(t *testing.T) {
	type fields struct {
		listeners []*http.Server
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "single server",
			fields: fields{
				[]*http.Server{
					&http.Server{
						Addr: "127.0.0.1:9000",
					},
				},
			},
		},
		{
			name: "multi server",
			fields: fields{
				[]*http.Server{
					&http.Server{
						Addr: "127.0.0.1:9000",
					},
					&http.Server{
						Addr: "[::1]:9000",
					},
				},
			},
		},
		/* Create an error on close does not work
		{
			name: "single server error",
			fields: fields{
				[]*http.Server{
					&http.Server{
						Addr: "127.0.0.1:9000",
					},
				},
			},
			wantErr: true,
		},
		*/
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Server{
				listeners: tt.fields.listeners,
			}
			s.listen()
			if tt.wantErr {
				go func() {
					r, err := http.Get("http://127.0.0.1:9000/")
					if err != nil {
						t.Fatal(err)
					}
					r.Body.Close()
				}()
				time.Sleep(time.Millisecond)
			}
			if err := s.Close(); (err != nil) != tt.wantErr {
				t.Errorf("Server.Close() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestServer_Shutdown(t *testing.T) {
	http.HandleFunc("/",
		func(w http.ResponseWriter, r *http.Request) {
			w.Write([]byte("Hello world!"))
			time.Sleep(2 * time.Second)
		},
	)
	ectx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	type fields struct {
		listeners []*http.Server
	}
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "single server",
			args: args{context.Background()},
			fields: fields{
				[]*http.Server{
					&http.Server{
						Addr: "127.0.0.1:9000",
					},
				},
			},
		},
		{
			name: "multi server",
			args: args{context.Background()},
			fields: fields{
				[]*http.Server{
					&http.Server{
						Addr: "127.0.0.1:9000",
					},
					&http.Server{
						Addr: "[::1]:9000",
					},
				},
			},
		},
		{
			name: "timeout",
			args: args{ectx},
			fields: fields{
				[]*http.Server{
					&http.Server{
						Addr: "127.0.0.1:9000",
					},
					&http.Server{
						Addr: "[::1]:9000",
					},
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Server{
				listeners: tt.fields.listeners,
			}
			s.listen()
			if tt.wantErr {
				go func() {
					r, err := http.Get("http://127.0.0.1:9000/")
					if err != nil {
						t.Fatal(err)
					}
					r.Body.Close()
				}()
				time.Sleep(10 * time.Millisecond)
			}
			if err := s.Shutdown(tt.args.ctx); (err != nil) != tt.wantErr {
				t.Errorf("Server.Shutdown() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestServer_ListenAndServe(t *testing.T) {
	type fields struct {
		listeners []*http.Server
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name: "single server",
			fields: fields{
				[]*http.Server{
					&http.Server{
						Addr: "127.0.0.1:9000",
					},
				},
			},
		},
		{
			name: "multi server",
			fields: fields{
				[]*http.Server{
					&http.Server{
						Addr: "127.0.0.1:9000",
					},
					&http.Server{
						Addr: "[::1]:9000",
					},
				},
			},
		},
		{
			name: "single server error",
			fields: fields{
				[]*http.Server{
					&http.Server{
						Addr: "123.123.123.123:9000",
					},
				},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Server{
				listeners: tt.fields.listeners,
			}
			go func() {
				time.Sleep(10 * time.Millisecond)
				s.Close()
			}()
			if err := s.ListenAndServe(); (err != nil && err != http.ErrServerClosed) != tt.wantErr {
				t.Errorf("Server.ListenAndServe() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

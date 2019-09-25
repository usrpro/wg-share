package server

import (
	"fmt"
	"net"
	"net/http"
)

// Configure the RPC server. If addrs is not specified, it defaults on listening
// on all available addresses on device.
//
// Device is also used by this package for the WireGuard specific queries.
// Those queries will fail if device is not a WG interface or does not exist.
// However, this is not a considered an error for Configure.
// Use addrs if you want the RPC server to listen on different addresses as the WG device.
func Configure(device string, port uint16, addrs ...string) (*Server, error) {
	tcas, err := tcpAddrs(device, port, addrs...)
	if err != nil {
		return nil, err
	}
	s := new(Server)
	s.listeners, err = httpServers(device, tcas)
	if err != nil {
		return nil, err
	}
	return s, nil
}

// itfIPs retrieves all the IP addresses from a network interface identified by name
func itfIPs(name string) ([]net.IP, error) {
	itf, err := net.InterfaceByName(name)
	if err != nil {
		return nil, err
	}
	addrs, err := itf.Addrs()
	if err != nil {
		return nil, err
	}
	var ips []net.IP
	for _, a := range addrs {
		// Not checking for assert error;
		// itf.Addrs() always returns the net.IPNet implementation of net.Addr
		ips = append(ips, a.(*net.IPNet).IP)
	}
	return ips, nil
}

// parseIPs parses address strings into IP addresses
// error is returned if one of the addresses cannot be parsed into an IP
func parseIPs(addrs []string) ([]net.IP, error) {
	var ips []net.IP
	for _, a := range addrs {
		ip := net.ParseIP(a)
		if ip == nil {
			return nil, fmt.Errorf("Invalid IP address: %s", a)
		}
		ips = append(ips, ip)
	}
	return ips, nil
}

// tcpAddrs generates a slice of net.TCPAddr to be used for the listeners. (IP:Port)
// If addrs is empty, it obtains IP addresses from the network interface identied by name.
func tcpAddrs(name string, port uint16, addrs ...string) ([]net.TCPAddr, error) {
	var (
		ips []net.IP
		err error
	)
	if addrs == nil {
		ips, err = itfIPs(name)
	} else {
		ips, err = parseIPs(addrs)
	}
	if err != nil {
		return nil, err
	}
	var tcas []net.TCPAddr
	for _, ip := range ips {
		tcas = append(
			tcas,
			net.TCPAddr{
				IP:   ip,
				Port: int(port),
			},
		)
	}
	return tcas, nil
}

// httpServers configures multiple listeners with new RPC objects for each TCPAddr
func httpServers(device string, tcas []net.TCPAddr) ([]*http.Server, error) {
	var servers []*http.Server
	for _, a := range tcas {
		rpc, err := NewRPC(device)
		if err != nil {
			return nil, err
		}
		servers = append(
			servers,
			&http.Server{
				Addr:    a.String(),
				Handler: rpc,
			},
		)
	}
	return servers, nil
}

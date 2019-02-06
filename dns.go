package nksh

import (
	"errors"
	"net"
	"time"
)

func validIPs(ips []net.IP) bool {
	for _, ip := range ips {
		if ip.String() == "<nil>" {
			return false
		}
	}

	return len(ips) > 0
}

func DNSLookupIP(host string, retrys int) ([]net.IP, error) {
	for retrys > 0 {
		tmp, _ := net.LookupIP(host)
		if validIPs(tmp) {
			return tmp, nil
		}

		time.Sleep(1 * time.Second)
		retrys--
	}

	return nil, errors.New("no lookup results")
}

package nksh

import (
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/juju/errors"
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

func LookupClusterHosts(host string, port int, params ...string) ([]string, error) {
	ips, err := DNSLookupIP(host, 50)
	if err != nil {
		return nil, errors.Annotate(err, "DNSLookupIP")
	}

	res := []string{}
	for _, ip := range ips {
		res = append(res, fmt.Sprintf(
			"%s:%d%s", ip, port, strings.Join(params, ""),
		))
	}

	return res, nil
}

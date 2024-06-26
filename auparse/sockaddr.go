// Licensed to Elasticsearch B.V. under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. Elasticsearch B.V. licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package auparse

import (
	"fmt"
	"strconv"
)

// invalidSockAddrError means socket address size for family is invalid.
type invalidSockAddrError struct {
	family string
	size   int
}

func (e invalidSockAddrError) Error() string {
	if e.size < 4 {
		return fmt.Sprintf("invalid family: too short: %d", e.size)
	}
	return fmt.Sprintf("invalid socket address size family=%s: %d", e.family, e.size)
}

func parseSockaddr(s string) (map[string]string, error) {
	if len(s) < 4 {
		return nil, invalidSockAddrError{size: len(s)}
	}

	addressFamily, err := hexToDec(s[2:4] + s[0:2]) // host-order
	if err != nil {
		return nil, err
	}

	out := map[string]string{}
	switch addressFamily {
	case 1: // AF_UNIX
		socket, err := hexToString(s[4:])
		if err != nil {
			return nil, err
		}

		out["family"] = "unix"
		out["path"] = socket
	case 2: // AF_INET
		if len(s) < 16 {
			return nil, invalidSockAddrError{
				family: "ipv4",
				size:   len(s),
			}
		}

		port, err := hexToDec(s[4:8])
		if err != nil {
			return nil, err
		}

		ip, err := hexToIP(s[8:16])
		if err != nil {
			return nil, err
		}

		out["family"] = "ipv4"
		out["addr"] = ip
		out["port"] = strconv.Itoa(int(port))
	case 10: // AF_INET6
		if len(s) < 48 {
			return nil, invalidSockAddrError{
				family: "ipv6",
				size:   len(s),
			}
		}

		port, err := hexToDec(s[4:8])
		if err != nil {
			return nil, err
		}

		flow, err := hexToDec(s[8:16])
		if err != nil {
			return nil, err
		}

		ip, err := hexToIP(s[16:48])
		if err != nil {
			return nil, err
		}

		out["family"] = "ipv6"
		out["addr"] = ip
		out["port"] = strconv.Itoa(int(port))
		if flow > 0 {
			out["flow"] = strconv.Itoa(int(flow))
		}
	case 16: // AF_NETLINK
		out["family"] = "netlink"
		out["saddr"] = s
	default:
		out["family"] = strconv.Itoa(int(addressFamily))
		out["saddr"] = s
	}

	return out, nil
}

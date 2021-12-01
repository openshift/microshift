/*
Copyright © 2021 Microshift Contributors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package server

import (
	"fmt"
	"net"

	"github.com/miekg/dns"
	"k8s.io/klog/v2"
)

const (
	DefaultmDNSTLD = ".local"
	ipV4MDNSAddr   = "224.0.0.251"
	ipV6MDNSAddr   = "ff02::fb"
	mDNSPort       = 5353
)

type Server struct {
	iface     *net.Interface
	responder Responder
	listeners []*net.UDPConn
	stopCh    chan struct{}
}

type Responder interface {
	Answer(q dns.Question) []dns.RR
}

func New(iface *net.Interface, responder Responder, stopCh chan struct{}) (*Server, error) {
	srv := &Server{iface: iface, stopCh: stopCh, responder: responder}

	for network, ipaddr := range map[string]string{"udp4": ipV4MDNSAddr, "udp6": ipV6MDNSAddr} {
		listener, _ := net.ListenMulticastUDP(network, iface, &net.UDPAddr{
			IP:   net.ParseIP(ipaddr),
			Port: mDNSPort,
		})

		if listener != nil {
			srv.listeners = append(srv.listeners, listener)

			go func() {
				<-stopCh
				listener.Close()
			}()

			go srv.listenerLoop(listener)
		}
	}

	return srv, nil
}

func (s *Server) listenerLoop(c *net.UDPConn) {
	buf := make([]byte, 65536)
	for {
		length, addr, err := c.ReadFromUDP(buf)
		if length == 0 {
			/* connection closed */
			return
		} else if err != nil {
			klog.Errorf("Error receiving from udp: %s", err)
		} else if err := s.handlemDNSPacket(c, buf[:length], addr); err != nil {
			klog.Errorf("Error handling mdns query: %v", err)
		}
	}
}

func (s *Server) handlemDNSPacket(conn *net.UDPConn, packet []byte, from net.Addr) error {
	var query dns.Msg
	var answers []dns.RR
	unicast := false

	if err := query.Unpack(packet); err != nil {
		return err
	}
	if query.Opcode != dns.OpcodeQuery || query.Rcode != 0 || query.Truncated {
		return nil
	}

	// Handle all the questions and construct the answers
	for _, q := range query.Question {
		for _, a := range s.responder.Answer(q) {
			answers = append(answers, a)
		}
		unicast = unicast || q.Qclass&(1<<15) != 0
	}

	// Build into a DNS packet
	dnsMsg := &dns.Msg{
		MsgHdr: dns.MsgHdr{
			Id:            uint16(0),
			Response:      true,
			Opcode:        query.Opcode,
			Authoritative: true,
		},
		Compress: true,
		Answer:   answers,
	}

	if len(answers) == 0 {
		return nil
	}

	if err := s.sendmDNSResponse(conn, dnsMsg, from, unicast); err != nil {
		return fmt.Errorf("mdns: error sending response unicast=%v: %v", unicast, err)
	}

	return nil
}

func (s *Server) sendmDNSResponse(conn *net.UDPConn, resp *dns.Msg, from net.Addr, unicast bool) error {

	destAddr := from.(*net.UDPAddr)

	buf, err := resp.Pack()
	if err != nil {
		return err
	}

	if !unicast {
		if destAddr.IP.To4() != nil {
			destAddr.IP = net.ParseIP(ipV4MDNSAddr)
		} else {
			destAddr.IP = net.ParseIP(ipV6MDNSAddr)
		}
	}

	_, err = conn.WriteToUDP(buf, destAddr)
	if unicast {
		return err
	} else {
		/* multicast sometimes fails when listening on multiple interfaces */
		return nil
	}
}

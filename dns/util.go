package dns

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/Dreamacro/clash/common/cache"
	"github.com/Dreamacro/clash/common/picker"
	"github.com/Dreamacro/clash/log"

	D "github.com/miekg/dns"
)

func putMsgToCache(c *cache.LruCache, key string, q D.Question, msg *D.Msg) {
	// skip dns cache for acme challenge
	if q.Qtype == D.TypeTXT && strings.HasPrefix(q.Name, "_acme-challenge.") {
		log.Debugln("[DNS] dns cache ignored because of acme challenge for: %s", q.Name)
		return
	}

	var ttl uint32
	switch {
	case len(msg.Answer) != 0:
		ttl = msg.Answer[0].Header().Ttl
	case len(msg.Ns) != 0:
		ttl = msg.Ns[0].Header().Ttl
	case len(msg.Extra) != 0:
		ttl = msg.Extra[0].Header().Ttl
	default:
		log.Debugln("[DNS] response msg empty: %#v", msg)
		return
	}

	c.SetWithExpire(key, msg.Copy(), time.Now().Add(time.Second*time.Duration(ttl)))
}

func setMsgTTL(msg *D.Msg, ttl uint32) {
	for _, answer := range msg.Answer {
		answer.Header().Ttl = ttl
	}

	for _, ns := range msg.Ns {
		ns.Header().Ttl = ttl
	}

	for _, extra := range msg.Extra {
		extra.Header().Ttl = ttl
	}
}

func isIPRequest(q D.Question) bool {
	return q.Qclass == D.ClassINET && (q.Qtype == D.TypeA || q.Qtype == D.TypeAAAA)
}

func isIPorPTRRequest(q D.Question) bool {
	return q.Qclass == D.ClassINET && (q.Qtype == D.TypeA || q.Qtype == D.TypeAAAA || q.Qtype == D.TypePTR)
}

func transform(servers []NameServer, resolver *Resolver) []dnsClient {
	ret := []dnsClient{}
	for _, s := range servers {
		switch s.Net {
		case "https":
			ret = append(ret, newDoHClient(s.Addr, s.Interface, resolver))
			continue
		case "dhcp":
			ret = append(ret, newDHCPClient(s.Addr))
			continue
		}

		host, port, _ := net.SplitHostPort(s.Addr)
		ret = append(ret, &client{
			Client: &D.Client{
				Net: s.Net,
				TLSConfig: &tls.Config{
					ServerName: host,
				},
				UDPSize: 4096,
				Timeout: 5 * time.Second,
			},
			port:  port,
			host:  host,
			iface: s.Interface,
			r:     resolver,
		})
	}
	return ret
}

func handleMsgWithEmptyAnswer(r *D.Msg) *D.Msg {
	msg := &D.Msg{}
	msg.Answer = []D.RR{}

	msg.SetRcode(r, D.RcodeSuccess)
	msg.Authoritative = true
	msg.RecursionAvailable = true

	return msg
}

func msgToIP(msg *D.Msg) []net.IP {
	ips := []net.IP{}

	for _, answer := range msg.Answer {
		switch ans := answer.(type) {
		case *D.AAAA:
			ips = append(ips, ans.AAAA)
		case *D.A:
			ips = append(ips, ans.A)
		}
	}

	return ips
}

func batchExchange(ctx context.Context, clients []dnsClient, m *D.Msg) (msg *D.Msg, err error) {
	fast, ctx := picker.WithContext(ctx)
	for _, client := range clients {
		r := client
		fast.Go(func() (any, error) {
			m, err := r.ExchangeContext(ctx, m)
			if err != nil {
				return nil, err
			} else if m.Rcode == D.RcodeServerFailure || m.Rcode == D.RcodeRefused {
				return nil, errors.New("server failure")
			}
			return m, nil
		})
	}

	elm := fast.Wait()
	if elm == nil {
		err := errors.New("all DNS requests failed")
		if fErr := fast.Error(); fErr != nil {
			err = fmt.Errorf("%w, first error: %s", err, fErr.Error())
		}
		return nil, err
	}

	msg = elm.(*D.Msg)
	return
}

func containNonASCII(s string) bool {
	for _, c := range s {
		if c >= utf8.RuneSelf {
			return true
		}
	}
	return false
}

func IPtoPTR(ip net.IP) string {
	if v4 := ip.To4(); v4 != nil {
		// IPv4
		ptr := make([]string, len(v4))

		for i := 0; i < len(v4); i++ {
			ptr[i] = strconv.Itoa(int(v4[len(v4)-i-1]))
		}

		return strings.Join(ptr, ".") + ".in-addr.arpa"
	} else if v6 := ip.To16(); v6 != nil {
		// IPv6
		ptr := make([]string, len(v6)*2)

		for i := 0; i < len(v6); i++ {
			ptr[i*2] = strconv.FormatInt(int64(v6[len(v6)-i-1]&0xF), 16)
			ptr[i*2+1] = strconv.FormatInt(int64(v6[len(v6)-i-1]>>4), 16)
		}

		return strings.Join(ptr, ".") + ".ip6.arpa"
	} else {
		log.Debugln("[DNS] PTR conversion failed because of non-IP for: %s", ip)
		return ""
	}
}

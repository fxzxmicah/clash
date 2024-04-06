package dns

import (
	"bufio"
	"net"
	"os"
	"path"
	"runtime"
	"strings"

	"github.com/Dreamacro/clash/component/trie"
)

func LoadHosts() *trie.DomainTrie {
	f, err := os.Open(hostsPath())
	if err != nil {
		return nil
	}
	defer f.Close()

	t := trie.New()
	h := map[string][]net.IP{}
	p := map[string][]string{}

	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())

		// ignore comments
		if strings.HasPrefix(line, "//") || strings.HasPrefix(line, "#") {
			continue
		}

		// ignore records with non-ascii character
		if containNonASCII(line) {
			continue
		}

		names := strings.Fields(line)
		// ignore blank lines
		if len(names) == 0 {
			continue
		}

		ip := net.ParseIP(names[0])
		// ignore lines that do not start with IP
		if ip == nil {
			continue
		}

		ptr := transIpToPtr(ip)
		for _, name := range names[1:] {
			h[name] = append(h[name], ip)
			p[ptr] = append(p[ptr], name+".")
		}
	}

	for name, ips := range h {
		t.Insert(strings.ToLower(name), ips)
	}

	for ptr, names := range p {
		t.Insert(ptr, names)
	}

	return t
}

func hostsPath() string {
	switch runtime.GOOS {
	case "windows":
		return path.Join(os.Getenv("SYSTEMROOT"), "\\System32\\drivers\\etc\\hosts")
	default:
		return "/etc/hosts"
	}
}

package dns

import (
	"bufio"
	"bytes"
	"net"
	"os"
	"path"
	"runtime"
	"strings"
	"unicode/utf8"

	"github.com/Dreamacro/clash/component/trie"
)

func LoadHosts() *trie.DomainTrie {
	f, err := os.Open(hostsPath())
	if err != nil {
		return nil
	}

	t := trie.New()
	h := map[string][]net.IP{}

	sc := bufio.NewScanner(f)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())

		// ignore comments
		if strings.HasPrefix(line, "//") || strings.HasPrefix(line, "#") {
			continue
		}

		// ignore records with non-ascii character
		for i := 0; i < len(line); i++ {
			if line[i] >= utf8.RuneSelf {
				continue
			}
		}

		buf := bytes.NewBuffer([]byte(line))
		namesc := bufio.NewScanner(buf)
		namesc.Split(bufio.ScanWords)
		var ip net.IP
		for namesc.Scan() {
			name := namesc.Text()

			if ip == nil {
				ip = net.ParseIP(name)
				if ip == nil {
					break
				}
				continue
			}

			h[name] = append(h[name], ip)
		}
	}

	for name, ips := range h {
		t.Insert(strings.ToLower(name), ips)
	}

	defer f.Close()
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

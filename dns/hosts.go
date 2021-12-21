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

	sc := bufio.NewScanner(f)
	for sc.Scan() == true {
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
		for namesc.Scan() == true {
			name := namesc.Text()

			if ip == nil {
				ip = net.ParseIP(name)
				if ip == nil {
					break
				}
				continue
			}

			t.Insert(name, ip)
		}
	}

	return t
}

func hostsPath() string {
	switch runtime.GOOS {
	case "windows":
		var sysRoot string
		for _, e := range os.Environ() {
			subs := strings.SplitN(e, "=", 2)
			if subs[0] == "SystemRoot" && len(subs) == 2 {
				sysRoot = subs[1]
				break
			}
		}
		return path.Join(sysRoot, "\\System32\\drivers\\etc\\hosts")
	default:
		return "/etc/hosts"
	}
}

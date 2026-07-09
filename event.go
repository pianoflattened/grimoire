package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

type events struct {
	mu      sync.Mutex
	spells  []spell
	write   chan string
	ack     chan struct{}
}

func (m *events) sayok() {
	select {
	case m.ack <- struct{}{}:
	default: }
}

func (m *events) handleLine(line string) {
	parts := strings.SplitN(line, ";", 3)
	if len(parts) != 3 || parts[0] != "NEWPAGE" {
		log.Printf("unrecognized message: %q", line)
		return}

	urlbs, err := base64.StdEncoding.DecodeString(parts[1])
	if err != nil {
		log.Printf("bad base64 in NEWPAGE: %v", err)
		return}
	
	url := string(urlbs)

	tabid := strings.TrimSpace(parts[2])
	if !isnum(tabid) {
		log.Printf("bad tabId in NEWPAGE (%q)", parts[2])
		return}

	for _, dir := range m.spelldirs(url) {
		entries, err := os.ReadDir(dir)
		if err != nil {
			log.Printf("read dir %s failed: %v", dir, err)
			continue}
		
		for _, e := range entries {
			if e.IsDir() {continue}
			
			var fn string
			switch filepath.Ext(e.Name()) {
			case ".js": fn = "j"
			case ".css": fn = "c"
			default: continue}

			path := filepath.Join(dir, e.Name())
			data, err := os.ReadFile(path)
			if err != nil {
				log.Printf("read %s failed: %v", path, err)
				continue}

			esc, err := json.Marshal(string(data))
			if err != nil {
				log.Printf("build call for %s failed: %v", path, err)
				continue}

			select {
			case m.write <- fmt.Sprintf("%s(%s, %s);", fn, esc, tabid):
			case <-time.After(2 * time.Second): log.Printf("write channel stalled; dropping payload")}
		}
	}
}

func isnum(s string) bool {
	if s == "" {return false}
	for _, r := range s {
		if r < '0' || r > '9' {return false}}
	return true
}

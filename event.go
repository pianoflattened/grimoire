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
	mu     sync.Mutex
	routes []route
	ch     chan string
}

func (m *events) spelldirs(url string) []string {
	m.mu.Lock()
	defer m.mu.Unlock()
	var dirs []string
	for _, rt := range m.routes {
		if rt.match(url) {dirs = append(dirs, rt.dir)}}
	
	return dirs
}

func (m *events) doline(line string) {
	parts := strings.SplitN(line, ";", 3)
	if len(parts) != 3 || parts[0] != "NEWPAGE" {
		log.Printf("unrecognized message: %q", line)
		return}

	urlBytes, err := base64.StdEncoding.DecodeString(parts[1])
	if err != nil {
		log.Printf("bad base64 in NEWPAGE: %v", err)
		return}
	
	url := string(urlBytes)

	tabID := strings.TrimSpace(parts[2])
	if !isnum(tabID) {
		log.Printf("bad tabid in NEWPAGE (%q)", parts[2])
		return}

	for _, dir := range m.spelldirs(url) {
		log.Printf("%+v", dir)

		entries, err := os.ReadDir(dir)
		if err != nil {
			log.Printf("read dir %s failed: %v", dir, err)
			continue}
		
		for _, e := range entries {
			log.Printf("%+v", e)
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

			escaped, err := json.Marshal(string(data))
			if err != nil {
				log.Printf("build call for %s failed: %v", path, err)
				continue}

			msg := fmt.Sprintf("%s(%s, %s);", fn, escaped, tabID)
			select {
			case m.ch <- msg:
			case <-time.After(2*time.Second): log.Printf("write channel stalled; dropping payload")}
		}
	}
}

func isnum(s string) bool {
	if s == "" {return false}
	for _, r := range s {
		if r < '0' || r > '9' {return false}}
	return true
}

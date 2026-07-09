package main

import (
	"context"
	"fmt"
	"log"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/BurntSushi/toml"
	"github.com/fsnotify/fsnotify"
)

type config struct {
	SocketPath string        `toml:"socket"`
	Reconnect  int           `toml:"reconnect"` // ms
	Debounce   int           `toml:"debounce"`  // ms
	Routes     []RouteConfig `toml:"spell"`
}

type RouteConfig struct {
	Pattern string `toml:"pattern"` // substring or glob w/ '*'
	Dir     string `toml:"dir"`     // dir w/ js/css files
}

type spell struct {
	dir   string
	match func(string) bool
}

func (c config) rcdelay() time.Duration {
	if c.Reconnect <= 0 {return 500*time.Millisecond}
	return time.Duration(c.Reconnect)*time.Millisecond
}

func (c config) dbdelay() time.Duration {
	if c.Debounce <= 0 {return 100*time.Millisecond}
	return time.Duration(c.Debounce)*time.Millisecond
}

func loadconf(path string) (config, error) {
	var cfg config
	_, err := toml.DecodeFile(path, &cfg)
	if err != nil {return cfg, err}
	if cfg.SocketPath == "" {return cfg, fmt.Errorf("config: socket is required")}
	return cfg, nil
}

func (m *events) refreshspells(cfgs []RouteConfig) error {
	spells := make([]spell, 0, len(cfgs))
	for _, rc := range cfgs {
		matcher, err := compilepatt(rc.Pattern)
		if err != nil {return err}
		spells = append(spells, spell{dir: rc.Dir, match: matcher})
	}
	m.mu.Lock()
	m.spells = spells
	m.mu.Unlock()
	return nil
}

func (m *events) spelldirs(url string) []string {
	m.mu.Lock()
	defer m.mu.Unlock()
	var dirs []string
	for _, rt := range m.spells {
		if rt.match(url) {dirs = append(dirs, rt.dir)}
	}
	return dirs
}

func compilepatt(patt string) (func(string) bool, error) {
	if !strings.Contains(patt, "*") {
		return func(url string) bool {return strings.Contains(url, patt) }, nil}
	
	parts := strings.Split(patt, "*")
	for i, p := range parts {
		parts[i] = regexp.QuoteMeta(p)}
	
	re, err := regexp.Compile(strings.Join(parts, ".*"))
	if err != nil {
		return nil, fmt.Errorf("bad pattern %q: %w", patt, err)}
	
	return re.MatchString, nil
}

func watchconf(ctx context.Context, path string, debounce time.Duration, onchange func(config)) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {log.Fatalf("config watcher: %v", err)}
	defer watcher.Close()

	dir := filepath.Dir(path)
	if err := watcher.Add(dir); err != nil {log.Fatalf("watch config dir %s: %v", dir, err)}

	var timer *time.Timer
	for {
		select {
		case <-ctx.Done():
			if timer != nil {timer.Stop()}
			return

		case event, ok := <-watcher.Events:
			if !ok {return}
			if filepath.Clean(event.Name) != filepath.Clean(path) {continue}
			if event.Op & (fsnotify.Write|fsnotify.Create) == 0 {continue}

			if timer != nil {timer.Stop()}
			timer = time.AfterFunc(debounce, func() {
				newcfg, err := loadconf(path)
				if err != nil {
					log.Printf("config reload failed: %v", err)
					return}
				
				onchange(newcfg)
			})

		case err, ok := <-watcher.Errors:
			if !ok {return}
			log.Printf("config watcher error: %v", err)
		}
	}
}

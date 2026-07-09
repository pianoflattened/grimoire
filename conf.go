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

type conf struct {
	Refresh int     `toml:"refresh"`
	Spells  []Spell `toml:"spell"`
}

type Spell struct {
	Patt string `toml:"pattern"` // substring or glob w/ '*'
	Dir  string `toml:"dir"`     // dir w/ js/css files
}

type route struct {
	dir   string
	match func(string) bool
}

func (m *events) refreshspells(cfgs []Spell) error {
	routes := make([]route, 0, len(cfgs))
	for _, rc := range cfgs {
		matcher, err := compilepatt(rc.Patt)
		if err != nil {
			return err
		}
		routes = append(routes, route{dir: rc.Dir, match: matcher})
	}
	m.mu.Lock()
	m.routes = routes
	m.mu.Unlock()
	return nil
}

func (c conf) refreshdelay() time.Duration {
	if c.Refresh <= 0 {return 100*time.Millisecond}
	return time.Duration(c.Refresh)*time.Millisecond
}

func loadconf(path string) (conf, error) {
	var cfg conf
	_, err := toml.DecodeFile(path, &cfg)
	return cfg, err
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

func watchconf(ctx context.Context, path string, debounce time.Duration, onchange func(conf)) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {log.Fatalf("config watcher: %v", err)}
	defer watcher.Close()

	dir := filepath.Dir(path)
	if err := watcher.Add(dir); err != nil {
		log.Fatalf("watch config dir %s: %v", dir, err)}

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


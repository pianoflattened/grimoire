package main

import (
	"context"
	_ "embed"
	"errors"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

//go:embed defaults.toml
var defaultconf []byte
const lookhere = "conf.toml"
func main() {
	if _, err := os.Stat(lookhere); errors.Is(err, os.ErrNotExist) {
		err = os.WriteFile(lookhere, defaultconf, 0644)}

	cfg, err := loadconf(lookhere)
	if err != nil {log.Fatalf("config load: %v", err)}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	mgr := &events{write: make(chan string, 64), ack: make(chan struct{}, 1)}
	if err := mgr.refreshspells(cfg.Routes); err != nil {log.Fatalf("initial routes: %v", err)}

	go watchconf(ctx, lookhere, cfg.dbdelay(), func(newcfg config) {
		if err := mgr.refreshspells(newcfg.Routes); err != nil {
			log.Printf("route reload rejected: %v", err)
			return}
		
		log.Printf("config reloaded (%d routes)", len(newcfg.Routes))
	})

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		mgr.run(ctx, cfg)}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	<-sigCh
	log.Printf("shutting down")
	cancel()
	wg.Wait()
}

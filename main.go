package main

import (
	"context"
	_ "embed"
	"errors"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
)

//go:embed defaults.toml
var DefaultConfig []byte
const confpath = "conf.toml"
const logpath = "grimoire.log"

func main() {
	exe, err := os.Executable()
	if err == nil {os.Chdir(filepath.Dir(exe))}

	if logfile, err := os.OpenFile(logpath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644); err == nil {
		log.SetOutput(logfile)
		defer logfile.Close()}

	if _, err := os.Stat(confpath); errors.Is(err, os.ErrNotExist) {
		os.WriteFile(confpath, DefaultConfig, 0644)}

	cfg, err := loadconf(confpath)
	if err != nil {log.Fatalf("config load: %v", err)}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	context.AfterFunc(ctx, func() {os.Stdin.Close()})

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigCh
		log.Printf("signal received; shutting down")
		cancel()}()

	mgr := &events{ch: make(chan string, 64)}
	if err := mgr.refreshspells(cfg.Spells); err != nil {log.Fatalf("initial routes: %v", err)}

	go watchconf(ctx, confpath, cfg.refreshdelay(), func(newcfg conf) {
		if err := mgr.refreshspells(newcfg.Spells); err != nil {
			log.Printf("route reload rejected: %v", err)
			return}
		
		log.Printf("config reloaded (%d routes)", len(newcfg.Spells))
	})

	if err := sendboot(os.Stdout); err != nil {
		log.Fatalf("bootstrap failed: %v", err)}

	go mgr.dowrite(ctx, os.Stdout, cancel)
	mgr.doread(ctx, os.Stdin)
	log.Printf("shutting down")
}

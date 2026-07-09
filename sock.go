package main

import (
	"bufio"
	"context"
	_ "embed"
	"encoding/json"
	"log"
	"net"
	"strings"
	"sync"
	"time"
)

//go:embed boot.js
var bootstrapJS string

func sendboot(conn net.Conn) error {
	payload, err := json.Marshal(bootstrapJS)
	if err != nil {return err}
	_, err = conn.Write(append(payload, '\n'))
	return err
}

func (m *events) run(ctx context.Context, cfg config) {
	for {
		select {
		case <-ctx.Done(): return
		default: }

		conn, err := net.Dial("unix", cfg.SocketPath)
		if err != nil {
			log.Printf("dial %s failed; retrying: %v", cfg.SocketPath, err)
			select {
			case <-time.After(cfg.rcdelay()):
			case <-ctx.Done(): return}
			continue}
		
		log.Printf("connected to %s", cfg.SocketPath)

		if err := sendboot(conn); err != nil {
			log.Printf("bootstrap injection failed; reconnecting: %v", err)
			conn.Close()
			select {
			case <-time.After(cfg.rcdelay()):
			case <-ctx.Done(): return}
			continue}

		cctx, cancel := context.WithCancel(ctx)
		context.AfterFunc(cctx, func() {conn.Close()})
		
		var wg sync.WaitGroup
		wg.Add(2)

		go func() {
			defer wg.Done()
			defer cancel()
			m.doread(cctx, conn)}()
	
		go func() {
			defer wg.Done()
			defer cancel()
			select {
			case <-m.ack:
			case <-time.After(5*time.Second):
				log.Printf("no ack for bootstrap; reconnecting")
				return
			case <-cctx.Done(): return }
			m.dowrite(cctx, conn)}()

		wg.Wait()
		conn.Close()

		select {
		case <-ctx.Done(): return
		default: log.Printf("connection lost; reconnecting")}
	}
}

func (m *events) dowrite(ctx context.Context, conn net.Conn) {
	for {
		select {
		case <-ctx.Done(): return
		case call := <-m.write:
			payload, err := json.Marshal(call)
			if err != nil {
				log.Printf("encode error for %q: %v", call, err)
				continue}

			if _, err := conn.Write(append(payload, '\n')); err != nil {
				log.Printf("write error; will reconnect: %v", err)
				return}
			
			select {
			case <-m.ack:
			case <-time.After(5 * time.Second):
				log.Printf("no ack from beval; reconnecting")
				return
			case <-ctx.Done(): return}
		}
	}
}

func (m *events) doread(ctx context.Context, conn net.Conn) {
	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		select {
		case <-ctx.Done(): return
		default: }

		raw := strings.TrimSpace(scanner.Text())
		if raw == "" {continue}

		var line string
		if err := json.Unmarshal([]byte(raw), &line); err != nil {
			log.Printf("eval error from beval: %s", raw)
			m.sayok()
			continue}

		switch {
		case line == "OK": m.sayok()
		case strings.HasPrefix(line, "NEWPAGE;"): go m.handleLine(line)
		default:
			log.Printf("reply from beval: %q", line)
			m.sayok()
		}
	}
	
	if err := scanner.Err(); err != nil {
		select {
		case <-ctx.Done(): // probably closed on purpose. do nothing
		default: log.Printf("read error: %v", err)}
	}
}

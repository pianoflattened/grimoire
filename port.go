package main

import (
	"bufio"
	"context"
	_ "embed"
	"encoding/binary"
	"encoding/json"
	"io"
	"log"
	"strings"
)

func readnativemsg(r *bufio.Reader) ([]byte, error) {
	var lenbuf [4]byte
	if _, err := io.ReadFull(r, lenbuf[:]); err != nil {
		return nil, err}
	
	n := binary.LittleEndian.Uint32(lenbuf[:])
	buf := make([]byte, n)
	if _, err := io.ReadFull(r, buf); err != nil {
		return nil, err}
	
	return buf, nil
}

func writenativemsg(w io.Writer, payload []byte) error {
	var lenbuf [4]byte
	binary.LittleEndian.PutUint32(lenbuf[:], uint32(len(payload)))
	if _, err := w.Write(lenbuf[:]); err != nil {
		return err}
	
	_, err := w.Write(payload)
	return err
}

//go:embed boot.js
var bootjs string
func sendboot(w io.Writer) error {
	payload, err := json.Marshal(bootjs)
	if err != nil {return err}
	return writenativemsg(w, payload)
}

func (m *events) dowrite(ctx context.Context, w io.Writer, cancel context.CancelFunc) {
	for {
		select {
		case <-ctx.Done(): return
		case call := <-m.ch:
			payload, err := json.Marshal(call)
			if err != nil {
				log.Printf("encode error for %q: %v", call, err)
				continue}
			if err := writenativemsg(w, payload); err != nil {
				log.Printf("write error; shutting down: %v", err)
				cancel()
				return
			}
		}
	}
}

func (m *events) doread(ctx context.Context, r io.Reader) {
	reader := bufio.NewReader(r)
	for {
		select {
		case <-ctx.Done(): return
		default: }

		payload, err := readnativemsg(reader)
		if err != nil {
			if err != io.EOF {
				select {
				case <-ctx.Done(): // closed on purpose during shutdown
				default: log.Printf("read error: %v", err)}
			}
			return
		}

		var line string
		if err := json.Unmarshal(payload, &line); err != nil {
			log.Printf("unexpected non-string message from browser: %s", payload)
			continue}

		log.Printf("%q", line)
		switch {
		case strings.HasPrefix(line, "NEWPAGE;"): go m.doline(line)
		default:
			if line != "OK" {
				log.Printf("reply from browser: %q", line)
			}
		}
	}
}

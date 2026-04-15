//go:build windows
// +build windows

package pager

import (
	"fmt"
	"io"
	"os"
	"time"

	"golang.org/x/term"
)

// RunInteractiveSession starts a simple interactive loop reading single-key
// input from in and applying navigation/search commands to the pager. It
// requires both in and out to be terminals. The function returns when the
// user quits (presses 'q') or an IO error occurs.
func RunInteractiveSession(p *Pager, in *os.File, out *os.File) error {
	if p == nil {
		return fmt.Errorf("nil pager")
	}
	if in == nil || out == nil {
		return fmt.Errorf("stdin/stdout must be provided")
	}
	if !term.IsTerminal(int(in.Fd())) || !term.IsTerminal(int(out.Fd())) {
		return fmt.Errorf("stdin/stdout must be terminals")
	}

	// Put stdin into raw mode so we can read single keypresses.
	oldState, err := term.MakeRaw(int(in.Fd()))
	if err != nil {
		return fmt.Errorf("make raw: %w", err)
	}
	defer term.Restore(int(in.Fd()), oldState)

	// adjust pager height based on terminal rows (reserve one for prompt)
	_, h, err := term.GetSize(int(out.Fd()))
	if err == nil {
		if h > 1 {
			p.height = h - 1
		} else {
			p.height = h
		}
	}

	// Buffer for single-byte reads; some sequences may be longer but we
	// handle the common single-byte commands and simple search input.
	buf := make([]byte, 1)
	for {
		// Poll for resize and adjust height
		if _, h, err := term.GetSize(int(out.Fd())); err == nil {
			if h > 1 && p.height != h-1 {
				p.height = h - 1
				p.renderPage()
			}
		}

		n, err := in.Read(buf)
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return fmt.Errorf("read input: %w", err)
		}
		if n == 0 {
			time.Sleep(10 * time.Millisecond)
			continue
		}
		b := buf[0]
		switch b {
		case 'q':
			return nil
		case ' ':
			p.HandleKey(' ')
		case 'b':
			p.HandleKey('b')
		case 'j':
			p.HandleKey('j')
		case 'k':
			p.HandleKey('k')
		case 'n':
			p.HandleKey('n')
		case 'N':
			p.HandleKey('N')
		case '/':
			// Read a simple line (until CR/LF) for the search query
			var q []byte
			for {
				n, err := in.Read(buf)
				if err != nil || n == 0 {
					break
				}
				if buf[0] == '\r' || buf[0] == '\n' {
					break
				}
				q = append(q, buf[0])
			}
			query := string(q)
			if query != "" {
				p.Search(query)
				if len(p.Matches) > 0 {
					_, _ = p.Seek(0)
				}
			}
		default:
			// ignore other input for now
		}
	}
}

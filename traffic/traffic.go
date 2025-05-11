package traffic

import (
	"context"
	"fmt"
	"io"
	"net"
	"sync"
)

type Forwarder struct {
	cancel    context.CancelFunc
	runningWG sync.WaitGroup
	mu        sync.Mutex
}

// NewForwarder creates a new Forwarder instance.
func NewForwarder() *Forwarder {
	return &Forwarder{}
}

// Start starts forwarding from srcPort to destPort.
// It cancels and waits for any previous forwarder to shut down cleanly.
func (f *Forwarder) Start(srcPort, destPort string) error {
	// Ensure full address
	if destPort != "" && net.ParseIP(destPort) == nil && !containsHost(destPort) {
		destPort = "localhost:" + destPort
	}

	f.mu.Lock()
	defer f.mu.Unlock()

	// Cancel old context and wait for it to stop
	if f.cancel != nil {
		f.cancel()
		f.runningWG.Wait()
	}

	ctx, cancel := context.WithCancel(context.Background())
	f.cancel = cancel

	f.runningWG.Add(1)
	go func() {
		defer f.runningWG.Done()
		_ = startForwarding(ctx, srcPort, destPort)
	}()

	return nil
}

// Stop stops the forwarder cleanly.
func (f *Forwarder) Stop() {
	f.mu.Lock()
	defer f.mu.Unlock()

	if f.cancel != nil {
		f.cancel()
		f.cancel = nil
		f.runningWG.Wait()
	}
}

// Helper: check if destPort contains a host part like 127.0.0.1:port
func containsHost(addr string) bool {
	_, _, err := net.SplitHostPort(addr)
	return err == nil
}

// --- Internal forwarding logic ---

func startForwarding(ctx context.Context, srcPort, destPort string) error {
	listener, err := net.Listen("tcp", ":"+srcPort)
	if err != nil {
		return fmt.Errorf("failed to listen on port %s: %w", srcPort, err)
	}
	defer listener.Close()
	fmt.Printf("Forwarding from :%s to %s\n", srcPort, destPort)

	var wg sync.WaitGroup

	go func() {
		<-ctx.Done()
		listener.Close()
	}()

	for {
		conn, err := listener.Accept()
		if err != nil {
			select {
			case <-ctx.Done():
				return nil // properly exit the function
			default:
				fmt.Println("Accept error:", err)
				continue
			}
		}

		wg.Add(1)
		go func(c net.Conn) {
			defer wg.Done()
			handleConnection(c, destPort)
		}(conn)
	}

	wg.Wait()
	return nil
}

func handleConnection(srcConn net.Conn, destAddr string) {
	defer srcConn.Close()

	destConn, err := net.Dial("tcp", destAddr)
	if err != nil {
		fmt.Println("Dial error:", err)
		return
	}
	defer destConn.Close()

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		io.Copy(destConn, srcConn)
	}()
	go func() {
		defer wg.Done()
		io.Copy(srcConn, destConn)
	}()

	wg.Wait()
}

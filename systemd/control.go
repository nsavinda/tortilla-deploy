package systemd

import (
	"context"
	"fmt"
	"time"

	"github.com/coreos/go-systemd/v22/dbus"
)

func RestartService(unitName string) error {
	conn, err := dbus.NewWithContext(context.Background())
	if err != nil {
		return fmt.Errorf("failed to connect to systemd: %w", err)
	}
	defer conn.Close()

	ch := make(chan string)
	_, err = conn.RestartUnitContext(context.Background(), unitName, "replace", ch)
	if err != nil {
		return fmt.Errorf("failed to request restart: %w", err)
	}

	select {
	case res := <-ch:
		if res != "done" {
			return fmt.Errorf("restart not completed: %s", res)
		}
	case <-time.After(5 * time.Second):
		return fmt.Errorf("restart timed out")
	}
	return nil
}

func StopService(unitName string) error {
	conn, err := dbus.NewWithContext(context.Background())
	if err != nil {
		return fmt.Errorf("failed to connect to systemd: %w", err)
	}
	defer conn.Close()

	ch := make(chan string)

	_, err = conn.StopUnitContext(context.Background(), unitName, "replace", ch)
	if err != nil {
		return fmt.Errorf("failed to request stop: %w", err)
	}

	select {
	case res := <-ch:
		if res != "done" {
			return fmt.Errorf("stop not completed: %s", res)
		}
	case <-time.After(5 * time.Second):
		return fmt.Errorf("stop timed out")
	}
	return nil
}

package systemctl

import (
	"context"
	"errors"

	"github.com/coreos/go-systemd/v22/dbus"
)

func DisableService(name string) error {
	// connect to systemd
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	conn, err := dbus.NewSystemdConnectionContext(ctx)
	if err != nil {
		return err
	}

	defer conn.Close()

	// ensure service is stopped
	properties, err := conn.GetUnitPropertiesContext(ctx, name)
	if err != nil {
		return err
	}

	if properties["ActiveState"] == "active" {

		ch := make(chan string)
		_, err := conn.StopUnitContext(ctx, name, "replace", ch)
		if err != nil {
			return err
		}

		result := <-ch
		if result != "done" {
			return errors.New("failed to stop " + name)
		}
	}

	fileChanges, err := conn.DisableUnitFilesContext(ctx, []string{name}, false)
	if err != nil {
		return err
	}

	if len(fileChanges) != 1 {
		return errors.New("failed to disable " + name)
	}

	return nil
}

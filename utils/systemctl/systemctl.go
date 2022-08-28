package systemctl

import (
	"context"
	"errors"

	"github.com/coreos/go-systemd/v22/dbus"
)

func IsServiceEnabled(name string) (bool, error) {
	// connect to systemd
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	conn, err := dbus.NewSystemdConnectionContext(ctx)
	if err != nil {
		return false, err
	}

	defer conn.Close()

	property, err := conn.GetUnitPropertyContext(ctx, name, "UnitFileState")
	if err != nil {
		return false, err
	}

	if property.Value.Value() == "enabled" {
		return true, nil
	}

	return false, nil
}

func IsServiceRunning(name string) (bool, error) {
	// connect to systemd
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	conn, err := dbus.NewSystemdConnectionContext(ctx)
	if err != nil {
		return false, err
	}

	defer conn.Close()

	property, err := conn.GetUnitPropertyContext(ctx, name, "ActiveState")
	if err != nil {
		return false, err
	}

	return property.Value.Value() == "active", nil
}

func EnableService(name string) error {
	// connect to systemd
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	conn, err := dbus.NewSystemdConnectionContext(ctx)
	if err != nil {
		return err
	}

	defer conn.Close()

	_, _, err = conn.EnableUnitFilesContext(ctx, []string{name}, false, true)
	if err != nil {
		return err
	}

	// ensure service is enabled
	property, err := conn.GetUnitPropertyContext(ctx, name, "ActiveState")
	if err != nil {
		return err
	}

	if property.Value.Value() != "active" {

		ch := make(chan string)
		_, err := conn.StartUnitContext(ctx, name, "replace", ch)
		if err != nil {
			return err
		}

		result := <-ch
		if result != "done" {
			return errors.New("failed to start " + name)
		}
	}

	return nil
}

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

	_, err = conn.DisableUnitFilesContext(ctx, []string{name}, false)
	if err != nil {
		return err
	}

	return nil
}

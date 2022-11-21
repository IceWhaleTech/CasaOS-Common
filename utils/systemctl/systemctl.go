package systemctl

import (
	"context"
	"errors"
	"time"

	"github.com/coreos/go-systemd/v22/dbus"
)

var (
	// `done` indicates successful execution of a job.
	ResultDone = "done"

	// `canceled` indicates that a job has been canceled before it finished execution.
	ResultCanceled = "canceled"
	ErrorCanceled  = errors.New("job has been canceled before it finished execution")

	// `timeout` indicates that the job timeout was reached.
	ResultTimeout = "timeout"
	ErrorTimeout  = errors.New("job timeout was reached")

	// `failed` indicates that the job failed.
	ResultFailed = "failed"
	ErrorFailed  = errors.New("job failed")

	// `dependency` indicates that a job this job has been depending on failed and the job hence has been removed too.
	ResultDependency = "dependency"
	ErrorDependency  = errors.New("another job this job has been depending on failed and the job hence has been removed too")

	// `skipped` indicates that a job was skipped because it didn't apply to the units current state.
	ResultSkipped = "skipped"
	ErrorSkipped  = errors.New("job was skipped because it didn't apply to the units current state")

	ErrorMap = map[string]error{
		ResultDone:       nil,
		ResultCanceled:   ErrorCanceled,
		ResultTimeout:    ErrorTimeout,
		ResultFailed:     ErrorFailed,
		ResultDependency: ErrorDependency,
		ResultSkipped:    ErrorSkipped,
	}

	ErrorUnknown = errors.New("unknown error")
)

func IsServiceEnabled(name string) (bool, error) {
	// connect to systemd
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
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
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
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
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
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
		return StartService(name)
	}

	return nil
}

func DisableService(name string) error {
	// connect to systemd
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
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
		return StopService(name)
	}

	_, err = conn.DisableUnitFilesContext(ctx, []string{name}, false)
	if err != nil {
		return err
	}

	return nil
}

func StartService(name string) error {
	// connect to systemd
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	conn, err := dbus.NewSystemdConnectionContext(ctx)
	if err != nil {
		return err
	}

	defer conn.Close()

	ch := make(chan string)
	_, err = conn.StartUnitContext(ctx, name, "replace", ch)
	if err != nil {
		return err
	}

	result := <-ch
	if result != ResultDone {
		err, ok := ErrorMap[result]
		if !ok {
			return ErrorUnknown
		}

		return err
	}

	return nil
}

func StopService(name string) error {
	// connect to systemd
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	conn, err := dbus.NewSystemdConnectionContext(ctx)
	if err != nil {
		return err
	}

	defer conn.Close()

	ch := make(chan string)
	_, err = conn.StopUnitContext(ctx, name, "replace", ch)
	if err != nil {
		return err
	}

	result := <-ch
	if result != ResultDone {
		err, ok := ErrorMap[result]
		if !ok {
			return ErrorUnknown
		}

		return err
	}

	return nil
}

func ReloadDaemon() error {
	// connect to systemd
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	conn, err := dbus.NewSystemdConnectionContext(ctx)
	if err != nil {
		return err
	}

	defer conn.Close()

	return conn.ReloadContext(ctx)
}

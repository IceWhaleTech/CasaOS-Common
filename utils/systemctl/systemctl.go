package systemctl

import (
	"context"
	"errors"
	"path/filepath"
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

type Service struct {
	Name    string
	Running bool
}

func ListServices(pattern string, wait ...time.Duration) ([]Service, error) {
	timeout := 30 * time.Second
	if len(wait) > 0 {
		timeout = wait[0]
	}

	// connect to systemd
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	conn, err := dbus.NewSystemdConnectionContext(ctx)
	if err != nil {
		return nil, err
	}

	defer conn.Close()

	var files []dbus.UnitFile

	if pattern == "" || pattern == "*" {
		_files, err := conn.ListUnitFilesContext(ctx)
		if err != nil {
			return nil, err
		}

		files = _files
	} else {
		_files, err := conn.ListUnitFilesByPatternsContext(ctx, nil, []string{pattern})
		if err != nil {
			return nil, err
		}
		files = _files
	}

	services := make([]Service, 0, len(files))

	for _, file := range files {
		serviceName := filepath.Base(file.Path)

		running, err := IsServiceRunning(serviceName)

		services = append(services, Service{
			Name:    serviceName,
			Running: err == nil && running,
		})
	}

	return services, nil
}

func IsServiceEnabled(name string, wait ...time.Duration) (bool, error) {
	timeout := 30 * time.Second
	if len(wait) > 0 {
		timeout = wait[0]
	}

	// connect to systemd
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
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

func IsServiceRunning(name string, wait ...time.Duration) (bool, error) {
	timeout := 30 * time.Second
	if len(wait) > 0 {
		timeout = wait[0]
	}

	// connect to systemd
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
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

func EnableService(nameOrPath string, wait ...time.Duration) error {
	timeout := 30 * time.Second
	if len(wait) > 0 {
		timeout = wait[0]
	}

	// connect to systemd
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	conn, err := dbus.NewSystemdConnectionContext(ctx)
	if err != nil {
		return err
	}

	defer conn.Close()

	_, _, err = conn.EnableUnitFilesContext(ctx, []string{nameOrPath}, false, true)
	if err != nil {
		return err
	}

	name := filepath.Base(nameOrPath)

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

func DisableService(name string, wait ...time.Duration) error {
	timeout := 30 * time.Second
	if len(wait) > 0 {
		timeout = wait[0]
	}

	// connect to systemd
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
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
		_ = StopService(name) // don't care about the result
	}

	_, err = conn.DisableUnitFilesContext(ctx, []string{name}, false)
	if err != nil {
		return err
	}

	return nil
}

func StartService(name string, wait ...time.Duration) error {
	timeout := 30 * time.Second
	if len(wait) > 0 {
		timeout = wait[0]
	}

	// connect to systemd
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
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

func StopService(name string, wait ...time.Duration) error {
	timeout := 30 * time.Second
	if len(wait) > 0 {
		timeout = wait[0]
	}

	// connect to systemd
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
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

func RestartService(name string, force bool, wait ...time.Duration) error {
	timeout := 30 * time.Second
	if len(wait) > 0 {
		timeout = wait[0]
	}

	// connect to systemd
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	conn, err := dbus.NewSystemdConnectionContext(ctx)
	if err != nil {
		return err
	}

	defer conn.Close()

	ch := make(chan string)

	if force {
		_, err = conn.RestartUnitContext(ctx, name, "replace", ch)
		if err != nil {
			return err
		}
	} else {
		_, err = conn.ReloadOrRestartUnitContext(ctx, name, "replace", ch)
		if err != nil {
			return err
		}
	}

	result := <-ch
	if result != ResultDone {
		err, ok := ErrorMap[result]
		if !ok {
			return ErrorUnknown
		}

		return err
	}

	_ = ReloadDaemon()

	return nil
}

func ReloadDaemon(wait ...time.Duration) error {
	timeout := 30 * time.Second
	if len(wait) > 0 {
		timeout = wait[0]
	}

	// connect to systemd
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	conn, err := dbus.NewSystemdConnectionContext(ctx)
	if err != nil {
		return err
	}

	defer conn.Close()

	return conn.ReloadContext(ctx)
}

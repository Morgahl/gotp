package grts

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"time"

	"github.com/Morgahl/gotp"
	"github.com/Morgahl/gotp/application"
	"github.com/Morgahl/gotp/assert"
	"github.com/Morgahl/gotp/dbg"
	"github.com/Morgahl/gotp/internal/ctx"
	"github.com/Morgahl/gotp/process"
	"github.com/Morgahl/gotp/supervisor"
	"github.com/Morgahl/gotp/term"
	"github.com/Morgahl/gotp/tuple"
)

var (
	init_ref process.Ref
)

type Status uint8

const (
	STARTING Status = iota
	STARTED
	STOPPING
)

type Flags struct{}

func Boot(flags Flags, apps ...application.Application) (err error) {
	assert.Zero(init_ref, "grts.Boot: already booted")
	ctx := ctx.Root()
	defer ctx.Cancel(fmt.Errorf("grts.Boot: exiting"))
	if err = startMetricsServer(); err != nil {
		slog.Warn("grts.Boot: failed to start metrics server", slog.Any("error", err))
	}

	exitCh := make(chan error)
	if init_ref, err = process.Spawn(__init(ctx, flags, apps, exitCh), process.Named("init")); err != nil {
		return err
	}
	slog.Debug("grts.Boot: init process started", slog.Any("pid", init_ref.PID()))

	err = <-exitCh
	slog.Debug("grts.Run: exiting")
	return err
}

func Stop(reason gotp.Atom) {
	assert.NotNil(init_ref, "grts.Stop: not booted")
	if initRef := process.WhereIs("init"); initRef.IsValid() {
		process.Send(initRef, stopMsg{reason})
	}
}

type appStarted tuple.T2[gotp.Atom, process.Ref]

type message[T term.Term] tuple.T2[process.PID, T]

type stopMsg tuple.T1[error]

func __init(rootCtx ctx.Cancellable, flags Flags, apps []application.Application, exitCh chan<- error) process.RunFn {
	var bootRef process.Ref
	status := STARTING
	kernel := make(map[gotp.Atom]process.Ref)
	pidToName := make(map[process.PID]gotp.Atom)
	return func(pctx process.Context) (reason error) {
		defer func() {
			reason = dbg.Recover(recover(), "grts.__init", reason)
			exitCh <- reason
			close(exitCh)
			rootCtx.Cancel(reason)
		}()
		pctx.TrapExit(true)
		bootRef, reason = doBoot(pctx.Ref(), apps)
		if reason != nil {
			return reason
		}

		var msg term.Term
		var ok bool

		// boot loop; starting applications
	BOOT_LOOP:
		for {
			select {
			case <-rootCtx.Done():
				reason = context.Cause(rootCtx)
				goto STOP_LOOP
			default:
				if status == STOPPING {
					goto STOP_LOOP
				}
			}

			if msg, ok, reason = process.ReceiveContext[term.Term](pctx, rootCtx); reason != nil {
				if cause := context.Cause(rootCtx); errors.Is(reason, cause) {
					slog.DebugContext(pctx.Context(), "init.receive: received shutdown signal", slog.Any("reason", reason))
					reason = cause
					goto STOP_LOOP
				}
				return reason
			} else if !ok {
				dbg.Throw("init.receive: channel closed")
			}
			switch msg := msg.(type) {

			case appStarted:
				if _, exists := kernel[msg.E0]; exists {
					dbg.Throw("application started twice: %s", msg.E0)
				}
				kernel[msg.E0] = msg.E1
				pidToName[msg.E1.PID()] = msg.E0

			case process.ExitMsg:
				switch msg.PID {

				case bootRef.PID():
					if errors.Is(msg.Reason, process.NORMAL) {
						// boot process exited normally
						status = STARTED
						bootRef = process.Ref{}
						break BOOT_LOOP
					}
					dbg.Throw("runtime terminated during boot: %v", msg.Reason)

				default:
					if appName, exists := pidToName[msg.PID]; exists {
						delete(pidToName, msg.PID)
						delete(kernel, appName)
						return msg.Reason
					}
					// ignore exits from unknown pids
				}

			case stopMsg:
				reason = msg.E0
				goto STOP_LOOP

			case message[gotp.Atom]:
				switch msg.E1 {
				case "get_applications":
					kApps := make([]application.Application, len(apps))
					copy(kApps, apps)
					process.Send(msg.E0, kApps)
				case "get_status":
					process.Send(msg.E0, status)
				}

			default:
				dbg.Throw("unexpected message during boot: %T", msg)
			}
		}

		// main loop; running applications
	MAIN_LOOP:
		for {
			if msg, ok, reason = process.ReceiveContext[term.Term](pctx, rootCtx); reason != nil {
				if cause := context.Cause(rootCtx); errors.Is(reason, cause) {
					slog.DebugContext(pctx.Context(), "init.receive: received shutdown signal", slog.Any("reason", reason))
					reason = cause
					goto STOP_LOOP
				}
				return reason
			} else if !ok {
				dbg.Throw("init.receive: channel closed")
			}
			switch msg := msg.(type) {

			case appStarted:
				if _, exists := kernel[msg.E0]; exists {
					dbg.Throw("application started twice: %s", msg.E0)
				}
				kernel[msg.E0] = msg.E1
				pidToName[msg.E1.PID()] = msg.E0

			case process.ExitMsg:
				if appName, exists := pidToName[msg.PID]; exists {
					delete(pidToName, msg.PID)
					delete(kernel, appName)
					return msg.Reason
				}
				// ignore exits from unknown pids

			case stopMsg:
				reason = msg.E0
				status = STOPPING
				break MAIN_LOOP

			case message[gotp.Atom]:
				switch msg.E1 {
				case "get_applications":
					kApps := make([]application.Application, len(apps))
					copy(kApps, apps)
					process.Send(msg.E0, kApps)
				case "get_status":
					process.Send(msg.E0, status)
				}

			default:
				dbg.Throw("unexpected message during boot: %T", msg)
			}
		}

		// stop loop; shutting down applications
	STOP_LOOP:
		status = STOPPING
		idx := len(apps) - 1
		var waitingPID process.PID
		for {
			slog.DebugContext(pctx.Context(), "grts.__init: shutting down applications", slog.Any("remaining", idx+1))
			if bootRef.IsValid() {
				slog.DebugContext(pctx.Context(), "grts.__init: stopping boot process", slog.Any("pid", bootRef.PID()))
				process.Send(bootRef, process.ExitMsg{PID: bootRef.PID(), Reason: reason})
			} else if idx >= 0 {
				app := apps[idx]
				slog.DebugContext(pctx.Context(), "grts.__init: stopping application", slog.Any("name", app.Name()))
				if ref, exists := kernel[app.Name()]; exists {
					if ref.PID() != waitingPID {
						waitingPID = ref.PID()
						process.Send(ref, process.ExitMsg{PID: ref.PID(), Reason: reason})
					}
				}
			} else {
				goto EXIT
			}
		INNER:
			for {
				msg, ok, err := process.ReceiveWithTimeout[term.Term](pctx, 0)
				if err != nil {
					slog.ErrorContext(pctx.Context(), "grts.__init: error receiving ExitMsg during shutdown", slog.Any("error", err))
					break
				} else if !ok {
					slog.DebugContext(pctx.Context(), "grts.__init: process exited during shutdown")
					break
				}
				switch msg := msg.(type) {

				case appStarted:
					if _, exists := kernel[msg.E0]; exists {
						dbg.Throw("application started twice: %s", msg.E0)
					}
					kernel[msg.E0] = msg.E1
					pidToName[msg.E1.PID()] = msg.E0

				case process.ExitMsg:
					switch msg.PID {
					case bootRef.PID():
						slog.DebugContext(pctx.Context(), "grts.__init: boot process exited", slog.Any("reason", msg.Reason), slog.Any("idx", idx))
						bootRef = process.Ref{}
						break INNER

					case waitingPID:
						if appName, exists := pidToName[msg.PID]; exists {
							slog.DebugContext(pctx.Context(), "grts.__init: application exited", slog.Any("app_name", appName), slog.Any("reason", msg.Reason))
							delete(pidToName, msg.PID)
							delete(kernel, appName)
						}
						waitingPID = process.PID{}
						idx--
						if idx < 0 {
							goto EXIT
						}
						break INNER
					default:
						// ignore exits from unknown pids
						slog.DebugContext(pctx.Context(), "grts.__init: ignoring exit from unknown pid", slog.Any("pid", msg.PID), slog.Any("reason", msg.Reason))
					}

				case message[gotp.Atom]:
					switch msg.E1 {
					case "get_applications":
						kApps := make([]application.Application, len(apps))
						copy(kApps, apps)
						process.Send(msg.E0, kApps)
					case "get_status":
						process.Send(msg.E0, status)
					}

				default:
					slog.DebugContext(pctx.Context(), "grts.__init: ignoring unexpected message during shutdown", slog.Any("msg", msg))
				}
			}
		}
	EXIT:
		return reason
	}
}

func doBoot(initRef process.Ref, apps []application.Application) (process.Ref, error) {
	return process.SpawnLink(boot(initRef, apps), initRef, process.Named("boot"))
}

func boot(initRef process.Ref, apps []application.Application) process.RunFn {
	return func(pctx process.Context) (reason error) {
		pctx.TrapExit(true)
		startApps := time.Now()

		// TODO: review https://github.com/erlang/otp/blob/master/erts/preloaded/src/init.erl#L1193-L1221
		var supervisable supervisor.Supervisable
		var ref process.Ref
		for _, app := range apps {
			startUp := time.Now()
			slog.DebugContext(pctx.Context(), "grts.boot: starting application", slog.Any("name", app.Name()), slog.Any("version", app.Version()))
			if supervisable, reason = app.Start(application.Normal()); reason != nil {
				slog.ErrorContext(pctx.Context(), "grts.boot: failed to start application", slog.Any("error", reason))
				return reason
			}

			slog.InfoContext(pctx.Context(), "grts.boot: starting application supervision tree", slog.Duration("start_hook_took", time.Since(startUp)))
			if ref, reason = supervisable.ChildSpec().Start(process.Link(initRef)); reason != nil {
				slog.ErrorContext(pctx.Context(), "grts.boot: failed to start application supervision tree", slog.Any("error", reason))
				return reason
			}

			slog.InfoContext(pctx.Context(), "grts.boot: application supervision tree started", slog.Duration("took", time.Since(startUp)))
			initRef.Send(appStarted{app.Name(), ref})
		}

		slog.InfoContext(pctx.Context(), "grts.boot: all applications started", slog.Duration("took", time.Since(startApps)))

		return process.NORMAL
	}
}

func startMetricsServer() error {
	ln6, err6 := net.Listen("tcp", "[::1]:6060")
	ln4, err4 := net.Listen("tcp", "127.0.0.1:6060")

	if err6 != nil && err4 != nil {
		slog.Warn("grts.Boot: failed to start http server for metrics", slog.Any("ipv6_error", err6), slog.Any("ipv4_error", err4))
		return errors.Join(err6, err4)
	}

	if ln6 != nil {
		go func() {
			if err := http.Serve(ln6, nil); err != nil {
				slog.Error("http.Serve ipv6 failed", slog.Any("error", err))
			}
		}()
	}

	if ln4 != nil {
		go func() {
			if err := http.Serve(ln4, nil); err != nil {
				slog.Error("http.Serve ipv4 failed", slog.Any("error", err))
			}
		}()
	}
	return nil
}

package main

import (
	"context"
	"flag"
	"log"
	"log/slog"
	"net/netip"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/Morgahl/gotp/internal/ctx"
	"github.com/Morgahl/gotp/internal/gpmd"
)

const (
	HEARTHBEAT_INTERVAL = time.Minute
	RECONNECT_INTERVAL  = 5 * time.Second
)

var (
	Daemon = flag.Bool(
		"daemon", false,
		"run as a daemon process (default: false)")
	Debug = flag.Uint(
		"debug", 0,
		"enable debug logging; 0: off, 1: on, 2: verbose")
	Address = flag.String(
		"address", "",
		"list of addresses to bind to, comma separated; loopback is always implicitly included")
	Port = flag.Uint(
		"port", 0,
		"the port to bind this instance to if launched as a daemon or to connect to if launched as an inquiry; default: 4469")
	Names = flag.Bool(
		"names", false,
		"lists the names registered in the currently running gpmd")
	Kill = flag.Bool(
		"kill", false,
		"kills the currently running gpmd; only occurs if the database is empty")

	Regisiter = flag.String(
		"register", "",
		"registers a node with the given name@address format",
	)
)

var (
	addresses []netip.AddrPort
)

func init() {
	flag.Parse()

	if *Port == 0 {
		if port := os.Getenv("GERL_GPMD_PORT"); port != "" {
			if port, err := strconv.Atoi(port); err != nil {
				panic(err)
			} else {
				*Port = uint(port)
			}
		} else {
			*Port = 4469
		}
	}

	if *Address == "" {
		*Address = os.Getenv("GERL_GPMD_ADDRESS")
	}

	foundIP4Loopback, foundIP6Loopback := false, false
	for _, add := range strings.Split(*Address, ",") {
		if add == "" {
			continue
		}

		var addr netip.AddrPort
		addr, err := netip.ParseAddrPort(add)
		if err != nil {
			slog.Error("invalid address", "address", add, "error", err)
			panic("invalid address: " + err.Error())
		}

		if addr.Addr().IsLoopback() {
			if addr.Addr().Is4() {
				foundIP4Loopback = true
			} else if addr.Addr().Is6() {
				foundIP6Loopback = true
			}
		}

		addresses = append(addresses, addr)
	}

	if !foundIP4Loopback {
		addresses = append(addresses, netip.AddrPortFrom(netip.AddrFrom4([4]byte{127, 0, 0, 1}), uint16(*Port)))
	}

	if !foundIP6Loopback {
		addresses = append(addresses, netip.AddrPortFrom(netip.IPv6Loopback(), uint16(*Port)))
	}
}

func main() {
	ctx := ctx.Root(context.Background())
	switch {
	default:
		flag.Usage()
		return

	case *Daemon:
		asDaemon(ctx)

	case *Names:
		names(ctx)

	case *Kill:
		kill(ctx)

	case *Regisiter != "":
		for {
			select {
			case <-ctx.Done():
				return
			default:
				if err := register(ctx); err != nil {
					slog.Error("failed to register", "error", err)
					time.Sleep(RECONNECT_INTERVAL)
				}
			}
		}
	}
}

func asDaemon(ctx ctx.Cancellable) {
	_, err := gpmd.New(ctx, addresses)
	if err != nil {
		log.Fatalf("failed to start gpmd: %v", err)
	}

	<-ctx.Done()
	slog.Info("shutting down", "reason", context.Cause(ctx))
}

func names(ctx ctx.Cancellable) {
	var client *gpmd.Client
	var err error
	defer ctx.Cancel(err)
	if client, err = gpmd.NewClient(ctx, addresses[0]); err != nil {
		return
	}
	defer client.Close()

	var result gpmd.ActionResult[gpmd.NodeList]
	if result, err = gpmd.Names(client); err != nil {
		return
	}
	if result.Result.Error != nil {
		slog.Error("failed to list names", "reason", result.Result.Error)
		return
	}
	for _, node := range result.Result.OK {
		slog.Info("node", "name", node)
	}
}

func kill(ctx ctx.Cancellable) {
	var client *gpmd.Client
	var err error
	defer ctx.Cancel(err)
	if client, err = gpmd.NewClient(ctx, addresses[0]); err != nil {
		return
	}
	defer client.Close()

	var result gpmd.ActionResult[gpmd.None]
	if result, err = gpmd.Kill(client); err != nil {
		return
	}
	if result.Result.Error != nil {
		slog.Error("failed to kill", "reason", result.Result.Error)
		return
	}
	slog.Info("killed")
}

func register(ctx ctx.Cancellable) (err error) {
	var client *gpmd.Client
	if client, err = gpmd.NewClient(ctx, addresses[0]); err != nil {
		return
	}
	defer client.Close()

	node, err := gpmd.NodeFromString(*Regisiter)
	if err != nil {
		slog.Error("invalid node", "error", err)
		return
	}

	var result gpmd.ActionResult[gpmd.None]
	if result, err = gpmd.Register(client, node); err != nil {
		return
	}
	if result.Result.Error != nil {
		slog.Error("failed to register", "reason", result.Result.Error)
		return
	}
	slog.Info("REGISTERED")

	defer func() {
		var result gpmd.ActionResult[gpmd.None]
		if result, err = gpmd.Unregister(client, node); err != nil {
			return
		}
		if result.Result.Error != nil {
			slog.Error("failed to unregister", "reason", result.Result.Error)
			return
		}
		slog.Info("UNREGISTERED")
	}()

	heartbeat := time.NewTicker(HEARTHBEAT_INTERVAL)
	defer heartbeat.Stop()
	for {
		select {
		case <-ctx.Done():
			slog.Info("shutting down", "reason", context.Cause(ctx))
			return

		case <-heartbeat.C:
			if result, err = gpmd.Heartbeat(client, node); err != nil {
				return
			} else if result.Result.Error != nil {
				slog.Error("failed to heartbeat", "reason", result.Result.Error)
				return
			}
		}
	}
}

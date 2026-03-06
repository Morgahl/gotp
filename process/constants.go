package process

import (
	"strings"

	"github.com/Morgahl/gotp"
	"github.com/Morgahl/gotp/dbg"
)

const (
	KILL                    gotp.Atom = "kill"
	KILLED                  gotp.Atom = "killed"
	NORMAL                  gotp.Atom = "normal"
	MESSAGE                 gotp.Atom = "message"
	LINK                    gotp.Atom = "link"
	UNLINK                  gotp.Atom = "unlink"
	EXIT                    gotp.Atom = "exit"
	MONITOR                 gotp.Atom = "monitor"
	DEMONITOR               gotp.Atom = "demonitor"
	DOWN                    gotp.Atom = "down"
	GROUP_LEADER            gotp.Atom = "group_leader"
	ALIVE_REQUEST           gotp.Atom = "alive_request"
	ALIVE_REPLY             gotp.Atom = "alive_reply"
	SPAWN_REQUEST           gotp.Atom = "spawn_request"
	SPAWN_REPLY             gotp.Atom = "spawn_reply"
	PROCESS_INFO_REQUEST    gotp.Atom = "process_info_request"
	PROCESS_INFO_REPLY      gotp.Atom = "process_info_reply"
	REGISTER_NAME_REQUEST   gotp.Atom = "register_name_request"
	REGISTER_NAME_REPLY     gotp.Atom = "register_name_reply"
	UNREGISTER_NAME_REQUEST gotp.Atom = "unregister_name_request"
	UNREGISTER_NAME_REPLY   gotp.Atom = "unregister_name_reply"
	WHERE_IS_REQUEST        gotp.Atom = "where_is_request"
	WHERE_IS_REPLY          gotp.Atom = "where_is_reply"
)

const (
	MAILBOX_SIZE = 4
	CHANNEL_SIZE = 20
)

type Flags uint8

const (
	// NO_FLAGS is used to indicate that no flags are set and is used as a default value
	NO_FLAGS Flags = 0

	// SENSITIVE_FLAG is used to indicate that the process is sensitive for logging and inspection purposes
	SENSITIVE_FLAG Flags = 1 << iota

	// TRAP_EXIT_FLAG is used to indicate that the process should trap exit signals
	TRAP_EXIT_FLAG
)

func (f Flags) IsSensitive() bool {
	return f&SENSITIVE_FLAG != 0
}

func (f Flags) IsTrapExit() bool {
	return f&TRAP_EXIT_FLAG != 0
}

func (f Flags) String() string {
	var flags []string
	if f.IsSensitive() {
		flags = append(flags, "sensitive")
	}
	if f.IsTrapExit() {
		flags = append(flags, "trap_exit")
	}
	if len(flags) == 0 {
		return "none"
	}
	return strings.Join(flags, "|")
}

type signalFlags uint8

const (
	// no_FLAGS is used to indicate that no flags are set and is used as a default value
	no_FLAGS signalFlags = 0

	// link_FLAG is used to indicate that the signal is due to a link or unlink operation
	link_FLAG signalFlags = 1 << iota

	// monitor_FLAG is used to indicate that the signal is due to a monitor or de-monitor operation
	monitor_FLAG

	// cast_FLAG is used to indicate that the signal is a cast message, i.e. it does not expect a reply
	cast_FLAG

	// request_FLAG is used to indicate that the signal is a request message, i.e. it expects a reply
	request_FLAG

	// reply_FLAG is used to indicate that the signal is a response to a request message
	reply_FLAG
)

func (f signalFlags) IsLink() bool {
	return f&link_FLAG != 0
}

func (f signalFlags) IsMonitor() bool {
	return f&monitor_FLAG != 0
}

func (f signalFlags) IsCast() bool {
	return f&cast_FLAG != 0
}

func (f signalFlags) IsRequest() bool {
	return f&request_FLAG != 0
}

func (f signalFlags) IsReply() bool {
	return f&reply_FLAG != 0
}

func (f signalFlags) String() string {
	var flags []string
	if f.IsLink() {
		flags = append(flags, "link")
	}
	if f.IsMonitor() {
		flags = append(flags, "monitor")
	}
	if f.IsCast() {
		flags = append(flags, "cast")
	}
	if f.IsRequest() {
		flags = append(flags, "request")
	}
	if f.IsReply() {
		flags = append(flags, "reply")
	}
	if len(flags) == 0 {
		return "none"
	}
	return strings.Join(flags, "|")
}

type signalType uint8

const (
	// sent when using the `Send` function
	MESSAGE_SIGNAL signalType = iota

	// Sent when calling the `Link` function
	LINK_SIGNAL

	// Sent when calling the `Unlink` function
	UNLINK_SIGNAL

	// Sent when calling the `Exit` function, oo when a linked process terminates
	EXIT_SIGNAL

	// Sent when calling the `Monitor` function
	MONITOR_SIGNAL

	// Sent when calling the `Demonitor` function
	DE_MONITOR_SIGNAL

	// Sent by a monitored process that terminates, or when a process monitoring another process terminates
	DOWN_SIGNAL

	// Sent when calling the `GroupLeader` function
	GROUP_LEADER_SIGNAL

	// Sent due to a call to one of the `Spawn`, `SpawnLink`, or `SpawnMonitor`, `SpawnOpt`, or `SpawnRequest` functions
	SPAWN_REQUEST_SIGNAL

	// Sent in response to a `Spawn`, `SpawnLink`, `SpawnMonitor`, `SpawnOpt`, or `SpawnRequest` function
	SPAWN_REPLY_SIGNAL

	// Sent when calling the `IsProcessAlive` function
	ALIVE_REQUEST_SIGNAL

	// Sent in response to the `IsProcessAlive` function
	ALIVE_REPLY_SIGNAL

	// Sent when calling the `ProcessInfo` function, no signalling is performed if directed at the calling process
	PROCESS_INFO_REQUEST_SIGNAL

	// Sent in response to the `ProcessInfo` function, no signalling is performed if directed at the calling process
	PROCESS_INFO_REPLY_SIGNAL

	// Sent when calling the `RegisterName` function
	REGISTER_NAME_REQUEST_SIGNAL

	// Sent in response to the `RegisterName` function
	REGISTER_NAME_REPLY_SIGNAL

	// Sent when calling the `UnregisterName` function
	UNREGISTER_NAME_REQUEST_SIGNAL

	// Sent in response to the `UnregisterName` function
	UNREGISTER_NAME_REPLY_SIGNAL

	// Sent when calling the `WhereIs` function
	WHERE_IS_REQUEST_SIGNAL

	// Sent in response to the `WhereIs` function
	WHERE_IS_REPLY_SIGNAL
)

func (s signalType) String() string {
	return s.Atom().String()
}

func (s signalType) Atom() gotp.Atom {
	switch s {
	case MESSAGE_SIGNAL:
		return MESSAGE
	case LINK_SIGNAL:
		return LINK
	case UNLINK_SIGNAL:
		return UNLINK
	case EXIT_SIGNAL:
		return EXIT
	case MONITOR_SIGNAL:
		return MONITOR
	case DE_MONITOR_SIGNAL:
		return DEMONITOR
	case DOWN_SIGNAL:
		return DOWN
	case GROUP_LEADER_SIGNAL:
		return GROUP_LEADER
	case SPAWN_REQUEST_SIGNAL:
		return SPAWN_REQUEST
	case SPAWN_REPLY_SIGNAL:
		return SPAWN_REPLY
	case ALIVE_REQUEST_SIGNAL:
		return ALIVE_REQUEST
	case ALIVE_REPLY_SIGNAL:
		return ALIVE_REPLY
	case PROCESS_INFO_REQUEST_SIGNAL:
		return PROCESS_INFO_REQUEST
	case PROCESS_INFO_REPLY_SIGNAL:
		return PROCESS_INFO_REPLY
	case REGISTER_NAME_REQUEST_SIGNAL:
		return REGISTER_NAME_REQUEST
	case REGISTER_NAME_REPLY_SIGNAL:
		return REGISTER_NAME_REPLY
	case UNREGISTER_NAME_REQUEST_SIGNAL:
		return UNREGISTER_NAME_REQUEST
	case UNREGISTER_NAME_REPLY_SIGNAL:
		return UNREGISTER_NAME_REPLY
	case WHERE_IS_REQUEST_SIGNAL:
		return WHERE_IS_REQUEST
	case WHERE_IS_REPLY_SIGNAL:
		return WHERE_IS_REPLY
	default:
		dbg.Throw("process.signalType.String: unknown signal type %d", s)
		panic("unreachable")
	}
}

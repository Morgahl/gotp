package process

import (
	"github.com/Morgahl/gotp/debug"
)

type flags uint8

const (
	NO_FLAGS  flags = 0
	LINK_FLAG flags = 1 << iota
	MONITOR_FLAG
	CAST_FLAG
	REQUEST_FLAG
	RESPONSE_FLAG
)

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
	switch s {
	case MESSAGE_SIGNAL:
		return "message"
	case LINK_SIGNAL:
		return "link"
	case UNLINK_SIGNAL:
		return "unlink"
	case EXIT_SIGNAL:
		return "exit"
	case MONITOR_SIGNAL:
		return "monitor"
	case DE_MONITOR_SIGNAL:
		return "demonitor"
	case DOWN_SIGNAL:
		return "down"
	case GROUP_LEADER_SIGNAL:
		return "group_leader"
	case SPAWN_REQUEST_SIGNAL:
		return "spawn_request"
	case SPAWN_REPLY_SIGNAL:
		return "spawn_reply"
	case ALIVE_REQUEST_SIGNAL:
		return "alive_request"
	case ALIVE_REPLY_SIGNAL:
		return "alive_reply"
	case PROCESS_INFO_REQUEST_SIGNAL:
		return "process_info_request"
	case PROCESS_INFO_REPLY_SIGNAL:
		return "process_info_reply"
	case REGISTER_NAME_REQUEST_SIGNAL:
		return "register_name_request"
	case REGISTER_NAME_REPLY_SIGNAL:
		return "register_name_reply"
	case UNREGISTER_NAME_REQUEST_SIGNAL:
		return "unregister_name_request"
	case UNREGISTER_NAME_REPLY_SIGNAL:
		return "unregister_name_reply"
	case WHERE_IS_REQUEST_SIGNAL:
		return "where_is_request"
	case WHERE_IS_REPLY_SIGNAL:
		return "where_is_reply"
	default:
		debug.Throw("process.signalType.String: unknown signal type %d", s)
		panic("unreachable")
	}
}

package process

import "fmt"

type signal[M Message] struct {
	_type   signalType
	flags   signalFlags
	message M
}

func messageSignal[M Message](flags signalFlags, message M) signal[M] {
	return signal[M]{
		_type:   MESSAGE_SIGNAL,
		flags:   flags,
		message: message,
	}
}

func linkSignal(link *Ref) signal[Message] {
	return signal[Message]{
		_type:   LINK_SIGNAL,
		flags:   LINK_FLAG | CAST_FLAG,
		message: link,
	}
}

func unlinkSignal(unlink *Ref) signal[Message] {
	return signal[Message]{
		_type:   UNLINK_SIGNAL,
		flags:   LINK_FLAG | CAST_FLAG,
		message: unlink,
	}
}

type exit struct {
	Sender   *Ref
	Receiver *Ref
	Reason   fmt.Stringer
}

func exitSignal(flags signalFlags, sender, receiver *Ref, reason fmt.Stringer) signal[Message] {
	return signal[Message]{
		_type: EXIT_SIGNAL,
		flags: flags,
		message: exit{
			Sender:   sender,
			Receiver: receiver,
			Reason:   reason,
		},
	}
}

func monitorSignal(monitor *Ref) signal[Message] {
	return signal[Message]{
		_type:   MONITOR_SIGNAL,
		flags:   MONITOR_FLAG | CAST_FLAG,
		message: monitor,
	}
}

func deMonitorSignal[M Message](deMonitor *Ref) signal[Message] {
	return signal[Message]{
		_type:   DE_MONITOR_SIGNAL,
		flags:   MONITOR_FLAG | CAST_FLAG,
		message: deMonitor,
	}
}

func downSignal(from PID, re *Ref, reason error) signal[Message] {
	return signal[Message]{
		_type: DOWN_SIGNAL,
		message: Down{
			From:   from,
			Ref:    re,
			Reason: reason,
		},
	}
}

func groupLeaderSignal(re *Ref) signal[Message] {
	return signal[Message]{
		_type:   GROUP_LEADER_SIGNAL,
		message: re,
	}
}

func aliveRequestSignal(from PID, re *Ref) signal[Message] {
	return signal[Message]{
		_type: ALIVE_REQUEST_SIGNAL,
		message: Request[any]{
			From: from,
			Ref:  re,
		},
	}
}

func aliveReplySignal(from PID, ref *Ref, err error) signal[Message] {
	return signal[Message]{
		_type: ALIVE_REPLY_SIGNAL,
		message: Reply[error]{
			From:    from,
			Ref:     ref,
			Message: err,
		},
	}
}

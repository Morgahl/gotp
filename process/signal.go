package process

type signal[M Message] struct {
	_type   signalType
	flags   flags
	message M
}

func messageSignal[M Message](flags flags, message M) signal[M] {
	return signal[M]{
		_type:   MESSAGE_SIGNAL,
		flags:   flags,
		message: message,
	}
}

func linkSignal(flags flags, link *Ref) signal[Message] {
	return signal[Message]{
		_type:   LINK_SIGNAL,
		flags:   flags | LINK_FLAG,
		message: link,
	}
}

func unlinkSignal(flags flags, unlink *Ref) signal[Message] {
	return signal[Message]{
		_type:   UNLINK_SIGNAL,
		flags:   flags | LINK_FLAG,
		message: unlink,
	}
}

type Exit struct {
	Sender   PID
	Receiver PID
	Reason   error
}

func exitSignal[M Message](flags flags, sender, receiver PID, reason error) signal[Message] {
	return signal[Message]{
		_type: EXIT_SIGNAL,
		flags: flags,
		message: Exit{
			Sender:   sender,
			Receiver: receiver,
			Reason:   reason,
		},
	}
}

func monitorSignal(flags flags, monitor *Ref) signal[Message] {
	return signal[Message]{
		_type:   MONITOR_SIGNAL,
		flags:   flags | MONITOR_FLAG,
		message: monitor,
	}
}

func deMonitorSignal[M Message](flags flags, deMonitor *Ref) signal[Message] {
	return signal[Message]{
		_type:   DE_MONITOR_SIGNAL,
		flags:   flags | MONITOR_FLAG,
		message: deMonitor,
	}
}

type Down struct {
	From   PID
	Ref    *Ref
	Reason error
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

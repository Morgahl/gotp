package process

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
		flags:   link_FLAG | cast_FLAG,
		message: link,
	}
}

func unlinkSignal(unlink *Ref) signal[Message] {
	return signal[Message]{
		_type:   UNLINK_SIGNAL,
		flags:   link_FLAG | cast_FLAG,
		message: unlink,
	}
}

type exitSig struct {
	Sender   PID
	Receiver *Ref
	Reason   error
}

func (e exitSig) ToExit() ExitMsg {
	return ExitMsg{
		Sender: e.Sender,
		Reason: e.Reason,
	}
}

func exitSignal(flags signalFlags, sender PID, receiver *Ref, reason error) signal[Message] {
	return signal[Message]{
		_type: EXIT_SIGNAL,
		flags: flags,
		message: exitSig{
			Sender:   sender,
			Receiver: receiver,
			Reason:   reason,
		},
	}
}

func monitorSignal(monitor *Ref) signal[Message] {
	return signal[Message]{
		_type:   MONITOR_SIGNAL,
		flags:   monitor_FLAG | cast_FLAG,
		message: monitor,
	}
}

func deMonitorSignal[M Message](deMonitor *Ref) signal[Message] {
	return signal[Message]{
		_type:   DE_MONITOR_SIGNAL,
		flags:   monitor_FLAG | cast_FLAG,
		message: deMonitor,
	}
}

func downSignal(from PID, re *Ref, reason error) signal[Message] {
	return signal[Message]{
		_type: DOWN_SIGNAL,
		message: DownMsg{
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

func aliveRequestSignal[F From](from F) signal[Message] {
	return signal[Message]{
		_type:   ALIVE_REQUEST_SIGNAL,
		message: RequestFrom[F, Message](from, nil),
	}
}

func aliveReplySignal[F From](request RequestMsg[Message], from F, err error) signal[Message] {
	return signal[Message]{
		_type:   ALIVE_REPLY_SIGNAL,
		message: ReplyTo(request, from, err),
	}
}

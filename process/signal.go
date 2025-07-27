package process

import (
	"log/slog"

	"github.com/Morgahl/gotp"
)

type signal[M gotp.Term] struct {
	_type   signalType
	flags   signalFlags
	message M
}

func (s signal[M]) LogValue() slog.Value {
	return slog.GroupValue(
		slog.Any("type", s._type.Atom()),
		slog.String("flags", s.flags.String()),
		slog.Any("message", s.message),
	)
}

func messageSignal[M gotp.Term](flags signalFlags, message M) signal[M] {
	return signal[M]{
		_type:   MESSAGE_SIGNAL,
		flags:   flags,
		message: message,
	}
}

func linkRequestSignal(link RequestMsg[Ref]) signal[gotp.Term] {
	return signal[gotp.Term]{
		_type:   LINK_SIGNAL,
		flags:   link_FLAG | request_FLAG,
		message: link,
	}
}

func linkReplySignal(link ReplyMsg[Ref]) signal[gotp.Term] {
	return signal[gotp.Term]{
		_type:   LINK_SIGNAL,
		flags:   link_FLAG | reply_FLAG,
		message: link,
	}
}

func unlinkSignal(unlink RequestMsg[Ref]) signal[gotp.Term] {
	return signal[gotp.Term]{
		_type:   UNLINK_SIGNAL,
		flags:   link_FLAG | cast_FLAG,
		message: unlink,
	}
}

type exitSig struct {
	PID    PID
	Ref    Ref
	Reason error
}

func (e exitSig) ToExit() ExitMsg {
	return ExitMsg{
		PID:    e.PID,
		Reason: e.Reason,
	}
}

func exitSignal(flags signalFlags, sender PID, ref Ref, reason error) signal[gotp.Term] {
	return signal[gotp.Term]{
		_type: EXIT_SIGNAL,
		flags: flags,
		message: exitSig{
			PID:    sender,
			Ref:    ref,
			Reason: reason,
		},
	}
}

func monitorSignal(monitor RequestMsg[Ref]) signal[gotp.Term] {
	return signal[gotp.Term]{
		_type:   MONITOR_SIGNAL,
		flags:   monitor_FLAG | cast_FLAG,
		message: monitor,
	}
}

func deMonitorSignal[M gotp.Term](deMonitor RequestMsg[Ref]) signal[gotp.Term] {
	return signal[gotp.Term]{
		_type:   DE_MONITOR_SIGNAL,
		flags:   monitor_FLAG | cast_FLAG,
		message: deMonitor,
	}
}

func downSignal(from PID, re Ref, reason error) signal[gotp.Term] {
	return signal[gotp.Term]{
		_type: DOWN_SIGNAL,
		message: DownMsg{
			From:   from,
			Ref:    re,
			Reason: reason,
		},
	}
}

func groupLeaderSignal(re Ref) signal[gotp.Term] {
	return signal[gotp.Term]{
		_type:   GROUP_LEADER_SIGNAL,
		message: re,
	}
}

func aliveRequestSignal[F From](from F) signal[gotp.Term] {
	return signal[gotp.Term]{
		_type:   ALIVE_REQUEST_SIGNAL,
		message: RequestFrom[F, gotp.Term](from, nil),
	}
}

func aliveReplySignal[F From](request RequestMsg[gotp.Term], from F, err error) signal[gotp.Term] {
	return signal[gotp.Term]{
		_type:   ALIVE_REPLY_SIGNAL,
		message: ReplyTo(request, from, err),
	}
}

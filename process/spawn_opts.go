package process

type SpawnOpt func(*Process)

func Link(link *Ref) SpawnOpt {
	return func(p *Process) {
		p.links.push(link)
		link.send(linkSignal(p.Ref()))
	}
}

func Monitor(monitor *Ref) SpawnOpt {
	return func(p *Process) {
		monitor.send(monitorSignal(p.Ref()))
	}
}

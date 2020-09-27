package common

import "sync"

type KillList struct {
	killers *sync.Map
}

func NewKillList() KillList {
	return KillList{
		&sync.Map{},
	}
}

type Killer func()

// AddKiller adds a killer function for the specified object
// only one killer is allowed per victim
func (t KillList) AddKiller(victim interface{}, killer Killer) {
	if _, ok := t.killers.Load(victim); ok {
		panic("victim already has killer")
	}
	t.killers.Store(victim, killer)
}

// Kill calls the killer function for the specified object
// nothing happens if the object is already killed or hasn't been added
func (t KillList) Kill(victim interface{}) {
	if killer, ok := t.killers.LoadAndDelete(victim); ok {
		killer.(Killer)()
	}
}

// KillAll tries to kill every connection open by this KillList
// errors during killing are not returned
// the struct cannot be used after being closed
func (t KillList) KillAll() {
	t.killers.Range(func(victim, killer interface{}) bool {
		t.Kill(victim)
		return true
	})
}

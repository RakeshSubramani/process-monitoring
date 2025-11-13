package alerts

import (
	"time"

	"github.com/RakeshSubramani/process-monitoring/pkg/agent"
	models "github.com/RakeshSubramani/process-monitoring/pkg/model"
)

type Rule struct {
	Name     string
	Interval time.Duration
	CheckFn  func(models.Metrics) bool
	ActionFn func(name string, m models.Metrics)
	lastFire time.Time
}

type Manager struct {
	rules []Rule
}

func NewManager() *Manager { return &Manager{} }

func (m *Manager) AddRule(r Rule) { m.rules = append(m.rules, r) }

func (m *Manager) Start(tick time.Duration) {
	t := time.NewTicker(tick)
	defer t.Stop()
	for range t.C {
		snap := agent.GetLatest()
		if !snap.Ready {
			continue
		}
		for i := range m.rules {
			r := &m.rules[i]
			if r.CheckFn(snap.System) {
				// cooldown  according to rule.Interval
				if time.Since(r.lastFire) > r.Interval {
					r.lastFire = time.Now()
					go r.ActionFn(r.Name, snap.System)
				}
			}
		}
	}
}

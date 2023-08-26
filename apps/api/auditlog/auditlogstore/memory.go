package auditlogstore

import (
	"go.dot.industries/brease/auditlog"
	"sync"
)

type Memory struct {
	entries []auditlog.Entry

	mu sync.Mutex
}

func (d *Memory) Store(entry auditlog.Entry) error {
	d.mu.Lock()
	defer d.mu.Unlock()

	d.entries = append(d.entries, entry)

	return nil
}

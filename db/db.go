package db

import (
	"fmt"
	"github.com/kazmanavt/ovsdb/monitor"
	"github.com/kazmanavt/ovsdb/schema"
	"github.com/kazmanavt/ovsdb/types"
	"strings"
	"sync"
	"time"
)

type DB interface {
	fmt.Stringer

	// RLock locks the database for reading.
	RLock()
	// RUnlock unlocks the database for reading.
	RUnlock()

	// Schema returns the schema of the database.
	Schema() *schema.DbSchema

	// TableSchema returns the schema of the table.
	// it panics if the table does not exist.
	TableSchema(tName string) *schema.TableSchema

	// TableLen returns the number of rows in the table.
	// it panics if the table does not exist.
	TableLen(tName string) int

	// TableRow returns the row with the given UUID in the table.
	// it returns nil if the row does not exist.
	TableRow(tName string, uuid types.UUID) schema.Row
	TableRowS(tName string, uuid string) schema.Row

	// Get returns the value of the column in the row.
	Get(tName string, uuid types.UUID, cName string) any
	GetS(tName string, uuid string, cName string) any

	// FindRecord returns a list of UUIDs of rows in the table that match the conditions.
	// it returns an empty list if no rows match the conditions.
	FindRecord(tName string, wheres ...[]types.Condition) []string

	// Update2 applies the updates2 received as result of monitor_cond or monitor_cond to current database.
	Update2(upd2 monitor.Updates2) error

	// WaitRevision waits until the database is updated to the given revision.
	// it returns true if the database is updated to the given revision. Otherwise, it returns false.
	WaitRevision(rev int, timeout time.Duration) bool
}

type dbImpl struct {
	name    string
	sch     *schema.DbSchema
	tNames  []string
	mu      sync.RWMutex
	tables  map[string]*tableImpl
	updated map[string]chan<- struct{}
}

func NewDB(sch *schema.DbSchema) DB {
	tNames := make([]string, 0, len(sch.Tables))
	tables := make(map[string]*tableImpl, len(sch.Tables))
	for tName, tSch := range sch.Tables {
		tNames = append(tNames, tName)
		cNames := make([]string, 0, len(tSch.Columns))
		for cName := range tSch.Columns {
			cNames = append(cNames, cName)
		}
		tables[tName] = &tableImpl{
			name:   tName,
			sch:    tSch,
			cNames: cNames,
			rows:   make(map[string]schema.Row),
		}
	}
	return &dbImpl{
		name:    sch.Name,
		sch:     sch,
		tNames:  tNames,
		tables:  tables,
		updated: make(map[string]chan<- struct{}),
	}
}

func (d *dbImpl) RLock() {
	d.mu.RLock()
}

func (d *dbImpl) RUnlock() {
	d.mu.RUnlock()
}

func (d *dbImpl) String() string {
	d.mu.RLock()
	defer d.mu.RUnlock()
	var sb strings.Builder
	_, _ = fmt.Fprintf(&sb, "DB %q:\n", d.name)
	for _, tN := range d.tNames {
		t := d.tables[tN]
		if len(t.rows) == 0 {
			continue
		}
		_, _ = fmt.Fprintf(&sb, "  Table %q:\n    ", t.name)
		for _, k := range t.cNames {
			_, _ = fmt.Fprintf(&sb, "| %q", k)
		}
		for uuid, row := range t.rows {
			_, _ = fmt.Fprintf(&sb, "\n    %s:\n    ", uuid)
			for _, k := range t.cNames {
				v := row.Get(k)
				_, _ = fmt.Fprintf(&sb, "| %v", v)
			}
		}
		_, _ = fmt.Fprintln(&sb)
	}
	return sb.String()
}

func (d *dbImpl) Schema() *schema.DbSchema {
	d.mu.RLock()
	defer d.mu.RUnlock()
	return d.sch
}

func (d *dbImpl) TableSchema(tName string) *schema.TableSchema {
	d.mu.RLock()
	defer d.mu.RUnlock()
	tSch, ok := d.sch.Tables[tName]
	if !ok {
		// FIXME: panic? really?
		panic("table not found")
	}
	return tSch
}

func (d *dbImpl) TableLen(tName string) int {
	d.mu.RLock()
	defer d.mu.RUnlock()
	t, ok := d.tables[tName]
	if !ok {
		panic(fmt.Sprintf("table %q does not exist", tName))
	}
	return len(t.rows)
}

func (d *dbImpl) TableRow(tName string, uuid types.UUID) schema.Row {
	d.mu.RLock()
	defer d.mu.RUnlock()
	t, ok := d.tables[tName]
	if !ok {
		return nil
	}
	row, ok := t.rows[string(uuid)]
	return row
}

func (d *dbImpl) TableRowS(tName, uuid string) schema.Row {
	return d.TableRow(tName, types.UUID(uuid))
}
func (d *dbImpl) Get(tName string, uuid types.UUID, cName string) any {
	return d.TableRow(tName, uuid).Get(cName)
}

func (d *dbImpl) GetS(tName, uuid, cName string) any {
	return d.TableRow(tName, types.UUID(uuid)).Get(cName)
}

// FindRecord returns a list of UUIDs of rows in the table that match the conditions.
// it returns an empty list if no rows match the conditions.
func (d *dbImpl) FindRecord(tName string, wheres ...[]types.Condition) []string {
	d.mu.RLock()
	defer d.mu.RUnlock()
	if t, ok := d.tables[tName]; ok {
		var res []string
		for _, where := range wheres {
			res = append(res, t.findRecord(where)...)
		}
		return res
	}
	return []string{}
}

func (d *dbImpl) Update2(upd2 monitor.Updates2) error {
	d.mu.Lock()
	defer d.mu.Unlock()
	defer func() {
		for _, ch := range d.updated {
			select {
			case ch <- struct{}{}:
			default:
			}
		}
	}()
	for tName, tUpd2 := range upd2 {
		if t, ok := d.tables[tName]; ok {
			if err := t.update2(tUpd2); err != nil {
				return err
			}
		}
	}
	return nil
}

func (d *dbImpl) SubscribeUpdates(uId string) <-chan struct{} {
	d.mu.Lock()
	defer d.mu.Unlock()
	updChan := make(chan struct{}, 100)
	d.updated[uId] = updChan
	return updChan
}

func (d *dbImpl) UnsubscribeUpdates(uId string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	close(d.updated[uId])
	delete(d.updated, uId)
}

func (d *dbImpl) WaitRevision(rev int, timeout time.Duration) bool {
	uuids := d.FindRecord("Open_vSwitch", nil)
	if len(uuids) != 1 {
		return false
	}

	upd := d.SubscribeUpdates("waitRevision")
	defer d.UnsubscribeUpdates("waitRevision")
	timeoutTimer := time.NewTimer(timeout)
	for {
		select {
		case <-upd:
			if d.GetS("Open_vSwitch", uuids[0], "cur_cfg").(int) >= rev {
				timeoutTimer.Stop()
				return true
			}
			continue
		case <-timeoutTimer.C:
			return false
		}
	}
}

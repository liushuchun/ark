package ta

import (
	"bytes"
	"github.com/qiniu/errors"
	"github.com/qiniu/log.v1"
	"github.com/qiniu/reliable"
	"github.com/qiniu/reliable/osl"
	"sync"
	"syscall"
)

// --------------------------------------------------------------------
// type IUnderlayer

type IUnderlayer interface {
	Underlayer() []osl.File
}

type IComponent interface {
	DoAct(act []byte) (err error)
	IUnderlayer
}

// --------------------------------------------------------------------

type fileModRecord struct {
	f   osl.File
	buf []byte
	off int64
}

type fileModRecords struct {
	recs []*fileModRecord
}

func (p *fileModRecords) add(f osl.File, buf []byte, off int64) {

	b := make([]byte, len(buf))
	copy(b, buf)

	rec := &fileModRecord{f: f, buf: b, off: off}
	p.recs = append(p.recs, rec)
}

func (p *fileModRecords) commit() (err error) {

	for _, rec := range p.recs {
		if len(rec.buf) != 0 {
			_, err = rec.f.WriteAt(rec.buf, rec.off)
			if err != nil {
				log.Warn("ta.fileModRecords.commit: WriteAt failed -", err)
				return
			}
		} else {
			err = rec.f.Truncate(rec.off)
			if err != nil {
				log.Warn("ta.fileModRecords.commit: Truncate failed -", err)
				return
			}
		}
	}
	p.recs = nil
	return nil
}

// --------------------------------------------------------------------
// type transFile

type transFile struct {
	osl.File
	*fileModRecords
}

func (p *transFile) WriteAt(buf []byte, off int64) (n int, err error) {

	if len(buf) > 0 {
		p.add(p.File, buf, off)
	}
	return len(buf), nil
}

func (p *transFile) Truncate(fsize int64) (err error) {

	p.add(p.File, nil, fsize)
	return nil
}

// --------------------------------------------------------------------
// type Transaction

const maxComponent = 64

type Transaction struct {
	coms  []IComponent
	mods  *fileModRecords
	mutex sync.RWMutex
	rl    RevertLog
	rlf   *reliable.Config
}

func OpenTransaction(f *reliable.Config) *Transaction {

	coms := make([]IComponent, maxComponent)
	mods := new(fileModRecords)
	p := &Transaction{coms: coms, mods: mods, rlf: f}
	p.rl.Buffer = bytes.NewBuffer(nil)
	return p
}

func (p *Transaction) Close() error {

	return p.rlf.Close()
}

func (p *Transaction) init(id int, comp IComponent) (err error) {

	if id >= maxComponent {
		return syscall.EINVAL
	}
	if p.coms[id] != nil {
		return syscall.EEXIST
	}
	files := comp.Underlayer()
	for i, f := range files {
		if f != nil {
			files[i] = &transFile{f, p.mods}
			if err != nil {
				return err
			}
		}
	}
	p.coms[id] = comp
	return
}

func (p *Transaction) beginRlog(id int) (rl RevertLog, hint int, err error) {

	if id >= maxComponent {
		err = syscall.EINVAL
		return
	}
	if p.coms[id] == nil {
		err = syscall.ENOENT
		return
	}
	return p.rl.begin(id)
}

func (p *Transaction) Begin() (err error) {

	p.mutex.Lock()
	return nil
}

func (p *Transaction) Rollback() {

	p.mods.recs = nil
	err := rollback(p.rl.Bytes(), p.coms)
	if err != nil {
		panic("Transaction.Rollback failed: " + err.Error())
	}
	p.rl.Reset()
}

func (p *Transaction) End() (err error) {

	return p.EndWithFail(false)
}

func (p *Transaction) EndWithFail(fail bool) (err error) {

	defer p.mutex.Unlock()

	if len(p.mods.recs) == 0 {
		p.rl.Reset()
		return
	}

	rlog := p.rl.Bytes()
	err = p.rlf.WriteFile(rlog)
	if err != nil {
		p.Rollback()
		err = errors.Info(err, "Transaction.End: rlf.WriteFile failed").Detail(err)
		return
	}

	if fail {
		return
	}

	err = p.mods.commit()
	if err != nil {
		log.Warn("Transaction.End: commit failed, to rollback")
		p.Rollback()
		err = errors.Info(err, "Transaction.End: commit failed, rollbacked").Detail(err)
	} else {
		p.rl.Reset()
	}
	p.rlf.WriteFile(nil)
	return
}

func (p *Transaction) Setup() error {

	rlog, err := p.rlf.ReadFile()
	if err != nil {
		err = p.rlf.Validate()
		if err == syscall.ENOENT {
			return nil
		}
		return err
	}

	err = rollback(rlog, p.coms)
	if err != nil {
		err = errors.Info(err, "Transaction.Setup: rollback failed").Detail(err)
		return err
	}

	err = p.mods.commit()
	if err != nil {
		err = errors.Info(err, "Transaction.End: commit failed").Detail(err)
	} else {
		p.rlf.WriteFile(nil)
	}
	return err
}

func (p *Transaction) RBegin() {

	p.mutex.RLock()
}

func (p *Transaction) REnd() {

	p.mutex.RUnlock()
}

// --------------------------------------------------------------------

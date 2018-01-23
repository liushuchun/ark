package ns

import (
	"github.com/qiniu/errors"
	"github.com/qiniu/reliable/log"
	"github.com/qiniu/reliable/osl"
	"io"
	"sync"
	"time"
)

// --------------------------------------------------------------------

type NamedServer struct {
	nmap  map[string]uint32
	vmap  map[uint32]string
	mutex sync.RWMutex
	logf  *log.Logger
	base  uint32
}

func OpenNsfile(fnames []string, linemax, allowfails int) (ns *NamedServer, err error) {

	files, err := osl.Open(fnames, allowfails)
	if err != nil {
		err = errors.Info(err, "reliable.ns.OpenNsfile failed", fnames).Detail(err)
		return
	}

	return Open(files, linemax, allowfails)
}

func Open(files []osl.File, linemax, allowfails int) (ns *NamedServer, err error) {

	f, err := log.OpenEx(files, linemax, allowfails)
	if err != nil {
		err = errors.Info(err, "reliable.ns.Open: OpenLogger failed").Detail(err)
		return
	}

	nmap := make(map[string]uint32)
	vmap := make(map[uint32]string)
	ns = &NamedServer{nmap: nmap, vmap: vmap, logf: f}
	err = ns.load()
	if err != nil {
		ns.Close()
		err = errors.Info(err, "reliable.ns.Open: load failed").Detail(err)
		return
	}

	return ns, nil
}

func (r *NamedServer) load() (err error) {

	var rr = r.logf.Reader(0)
	var id, idmax uint32
	var name string
	for {
		err = rr.Scanln(&id, &name)
		if err != nil {
			break
		}
		r.nmap[name] = id
		r.vmap[id] = name
		if idmax < id {
			idmax = id
		}
	}
	if err != io.EOF {
		err = errors.Info(err, "reliable.NamedServer.load failed").Detail(err)
		return
	}

	id = uint32(time.Now().Unix())
	if idmax < id {
		idmax = id
	}
	r.base = idmax
	return nil
}

func (r *NamedServer) Close() error {

	r.mutex.Lock()
	defer r.mutex.Unlock()

	if r.logf != nil {
		r.logf.Close()
		r.logf = nil
	}
	return nil
}

func (r *NamedServer) Register(name string) (val uint32, err error) {

	r.mutex.Lock()
	defer r.mutex.Unlock()

	val, ok := r.nmap[name]
	if ok {
		return
	}

	for {
		r.base++
		if _, ok := r.vmap[r.base]; !ok {
			break
		}
	}

	val = r.base
	err = r.logf.Println(val, name)
	if err != nil {
		err = errors.Info(err, "reliable.NamedServer.Register failed:", name).Detail(err)
		return
	}

	r.nmap[name] = val
	r.vmap[val] = name
	return
}

func (r *NamedServer) Find(name string) (val uint32, ok bool) {

	r.mutex.RLock()
	defer r.mutex.RUnlock()

	val, ok = r.nmap[name]
	return
}

func (r *NamedServer) FindRev(val uint32) (name string, ok bool) {

	r.mutex.RLock()
	defer r.mutex.RUnlock()

	name, ok = r.vmap[val]
	return
}

// --------------------------------------------------------------------

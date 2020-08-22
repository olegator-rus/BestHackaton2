package reassembly

import (
	"bytes"
	"fmt"
	"log"
	"sync"
	"time"
)

/*
 * StreamPool
 */

// StreamPool stores all streams created by Assemblers, allowing multiple
// assemblers to work together on stream processing while enforcing the fact
// that a single stream receives its data serially.  It is safe
// for concurrency, usable by multiple Assemblers at once.
//
// StreamPool handles the creation and storage of Stream objects used by one or
// more Assembler objects.  When a new TCP stream is found by an Assembler, it
// creates an associated Stream by calling its streamFactory's New method.
// Thereafter (until the stream is closed), that Stream object will receive
// assembled TCP data via Assembler's calls to the stream's Reassembled
// function.
//
// Like the Assembler, StreamPool attempts to minimize allocation.  Unlike the
// Assembler, though, it does have to do some locking to make sure that the
// connection objects it stores are accessible to multiple Assemblers.
type StreamPool struct {
	conns              map[key]*connection
	users              int
	mu                 sync.Mutex
	factory            streamFactory
	free               []*connection
	all                [][]connection
	nextAlloc          int
	newConnectionCount int64
}

func (p *StreamPool) grow() {
	conns := make([]connection, p.nextAlloc)
	p.all = append(p.all, conns)
	for i := range conns {
		p.free = append(p.free, &conns[i])
	}
	if Debug {
		log.Println("StreamPool: created", p.nextAlloc, "new connections")
	}
	p.nextAlloc *= 2
}

// dump logs all connections.
func (p *StreamPool) dump() {
	p.mu.Lock()
	defer p.mu.Unlock()
	log.Printf("Remaining %d connections: ", len(p.conns))
	for _, conn := range p.conns {
		log.Printf("%v %s", conn.key, conn)
	}
}

// DumpString logs all connections and returns a string.
func (p *StreamPool) DumpString() string {
	p.mu.Lock()
	defer p.mu.Unlock()
	var b bytes.Buffer

	b.WriteString(fmt.Sprintf("Remaining %d connections: \n", len(p.conns)))
	for _, conn := range p.conns {
		b.WriteString(fmt.Sprintf("%v %s\n", conn.key, conn))
	}

	return b.String()
}

func (p *StreamPool) remove(conn *connection) {
	p.mu.Lock()
	if _, ok := p.conns[*conn.key]; ok {
		delete(p.conns, *conn.key)
		p.free = append(p.free, conn)
	}
	p.mu.Unlock()
}

// NewStreamPool creates a new connection pool.  Streams will
// be created as necessary using the passed-in streamFactory.
func NewStreamPool(factory streamFactory) *StreamPool {
	return &StreamPool{
		conns:     make(map[key]*connection, initialAllocSize),
		free:      make([]*connection, 0, initialAllocSize),
		factory:   factory,
		nextAlloc: initialAllocSize,
	}
}

func (p *StreamPool) connections() []*connection {
	p.mu.Lock()
	conns := make([]*connection, 0, len(p.conns))

	for _, conn := range p.conns {
		conns = append(conns, conn)
	}

	p.mu.Unlock()

	return conns
}

func (p *StreamPool) newConnection(k *key, s Stream, ts time.Time) (c *connection, h *halfconnection, r *halfconnection) {
	if Debug {
		p.newConnectionCount++
		if p.newConnectionCount&0x7FFF == 0 {
			log.Println("StreamPool:", p.newConnectionCount, "requests,", len(p.conns), "used,", len(p.free), "free")
		}
	}

	if len(p.free) == 0 {
		p.grow()
	}

	index := len(p.free) - 1
	c, p.free = p.free[index], p.free[:index]
	c.reset(k, s, ts)

	return c, &c.c2s, &c.s2c
}

func (p *StreamPool) getHalf(k *key) (*connection, *halfconnection, *halfconnection) {
	conn := p.conns[*k]
	if conn != nil {
		return conn, &conn.c2s, &conn.s2c
	}

	rk := k.reverse()
	conn = p.conns[rk]

	if conn != nil {
		return conn, &conn.s2c, &conn.c2s
	}

	return nil, nil, nil
}

// getConnection returns a connection.  If end is true and a connection
// does not already exist, returns nil.  This allows us to check for a
// connection without actually creating one if it doesn't already exist.
func (p *StreamPool) getConnection(k *key, end bool, ts time.Time, ac AssemblerContext) (*connection, *halfconnection, *halfconnection) {
	p.mu.Lock()
	defer p.mu.Unlock()

	conn, half, rev := p.getHalf(k)

	if end || conn != nil {
		return conn, half, rev
	}

	s := p.factory.New(k[0], k[1], ac)

	conn, half, rev = p.newConnection(k, s, ts)

	conn2, half2, rev2 := p.getHalf(k)
	if conn2 != nil {
		if conn2.key != k {
			fmt.Println(conn.key, conn.c2s)
			panic("FIXME: other dir added in the meantime...")
		}

		// FIXME: delete s ?
		return conn2, half2, rev2
	}

	p.conns[*k] = conn

	return conn, half, rev
}

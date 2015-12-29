package pilosa_test

import (
	"reflect"
	"strings"
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/umbel/pilosa"
	"github.com/umbel/pilosa/pql"
)

// Ensure a get query can be executed.
func TestExecutor_Execute_Get(t *testing.T) {
	idx := MustOpenIndex()
	defer idx.Close()
	idx.MustCreateFragmentIfNotExists("d", "f", 0).MustSetBit(10, 3)
	idx.MustCreateFragmentIfNotExists("d", "f", 1).MustSetBit(10, SliceWidth+1)

	e := NewExecutor(idx.Index, NewCluster(1))
	if res, err := e.Execute("d", MustParse(`get(id=10, frame=f)`), nil); err != nil {
		t.Fatal(err)
	} else if chunks := res.(*pilosa.Bitmap).Chunks(); len(chunks) != 2 {
		t.Fatalf("unexpected chunk length: %s", spew.Sdump(chunks))
	} else if chunks[0].Value[0] != 8 {
		t.Fatalf("unexpected chunk(0): %s", spew.Sdump(chunks[0]))
	} else if chunks[1].Value[0] != 2 {
		t.Fatalf("unexpected chunk(1): %s", spew.Sdump(chunks[1]))
	}
}

// Ensure a difference query can be executed.
func TestExecutor_Execute_Difference(t *testing.T) {
	idx := MustOpenIndex()
	defer idx.Close()
	idx.MustCreateFragmentIfNotExists("d", "general", 0).MustSetBit(10, 1)
	idx.MustCreateFragmentIfNotExists("d", "general", 0).MustSetBit(10, 2)
	idx.MustCreateFragmentIfNotExists("d", "general", 0).MustSetBit(10, 3)
	idx.MustCreateFragmentIfNotExists("d", "general", 0).MustSetBit(11, 2)

	e := NewExecutor(idx.Index, NewCluster(1))
	if res, err := e.Execute("d", MustParse(`difference(get(id=10), get(id=11))`), nil); err != nil {
		t.Fatal(err)
	} else if chunks := res.(*pilosa.Bitmap).Chunks(); len(chunks) != 1 {
		t.Fatalf("unexpected chunk length: %s", spew.Sdump(chunks))
	} else if chunks[0].Value[0] != 10 { // b1010
		t.Fatalf("unexpected chunk(0): %s", spew.Sdump(chunks[0]))
	}
}

// Ensure an intersect query can be executed.
func TestExecutor_Execute_Intersect(t *testing.T) {
	idx := MustOpenIndex()
	defer idx.Close()
	idx.MustCreateFragmentIfNotExists("d", "general", 0).MustSetBit(10, 1)
	idx.MustCreateFragmentIfNotExists("d", "general", 1).MustSetBit(10, SliceWidth+1)
	idx.MustCreateFragmentIfNotExists("d", "general", 1).MustSetBit(10, SliceWidth+2)

	idx.MustCreateFragmentIfNotExists("d", "general", 0).MustSetBit(11, 1)
	idx.MustCreateFragmentIfNotExists("d", "general", 0).MustSetBit(11, 2)
	idx.MustCreateFragmentIfNotExists("d", "general", 1).MustSetBit(11, SliceWidth+2)

	e := NewExecutor(idx.Index, NewCluster(1))
	if res, err := e.Execute("d", MustParse(`intersect(get(id=10), get(id=11))`), nil); err != nil {
		t.Fatal(err)
	} else if chunks := res.(*pilosa.Bitmap).Chunks(); len(chunks) != 2 {
		t.Fatalf("unexpected chunk length: %s", spew.Sdump(chunks))
	} else if chunks[0].Value[0] != 2 {
		t.Fatalf("unexpected chunk(0): %s", spew.Sdump(chunks[0]))
	} else if chunks[1].Value[0] != 4 {
		t.Fatalf("unexpected chunk(1): %s", spew.Sdump(chunks[1]))
	}
}

// Ensure a union query can be executed.
func TestExecutor_Execute_Union(t *testing.T) {
	idx := MustOpenIndex()
	defer idx.Close()
	idx.MustCreateFragmentIfNotExists("d", "general", 0).MustSetBit(10, 0)
	idx.MustCreateFragmentIfNotExists("d", "general", 1).MustSetBit(10, SliceWidth+1)
	idx.MustCreateFragmentIfNotExists("d", "general", 1).MustSetBit(10, SliceWidth+2)

	idx.MustCreateFragmentIfNotExists("d", "general", 0).MustSetBit(11, 2)
	idx.MustCreateFragmentIfNotExists("d", "general", 1).MustSetBit(11, SliceWidth+2)

	e := NewExecutor(idx.Index, NewCluster(1))
	if res, err := e.Execute("d", MustParse(`union(get(id=10), get(id=11))`), nil); err != nil {
		t.Fatal(err)
	} else if chunks := res.(*pilosa.Bitmap).Chunks(); len(chunks) != 2 {
		t.Fatalf("unexpected chunk length: %s", spew.Sdump(chunks))
	} else if chunks[0].Value[0] != 5 {
		t.Fatalf("unexpected chunk(0): %s", spew.Sdump(chunks[0]))
	} else if chunks[1].Value[0] != 6 {
		t.Fatalf("unexpected chunk(1): %s", spew.Sdump(chunks[1]))
	}
}

// Ensure a count query can be executed.
func TestExecutor_Execute_Count(t *testing.T) {
	idx := MustOpenIndex()
	defer idx.Close()
	idx.MustCreateFragmentIfNotExists("d", "f", 0).MustSetBit(10, 3)
	idx.MustCreateFragmentIfNotExists("d", "f", 1).MustSetBit(10, SliceWidth+1)
	idx.MustCreateFragmentIfNotExists("d", "f", 1).MustSetBit(10, SliceWidth+2)

	e := NewExecutor(idx.Index, NewCluster(1))
	if n, err := e.Execute("d", MustParse(`count(get(id=10, frame=f))`), nil); err != nil {
		t.Fatal(err)
	} else if n != uint64(3) {
		t.Fatalf("unexpected n: %d", n)
	}
}

// Ensure a set query can be executed.
func TestExecutor_Execute_Set(t *testing.T) {
	idx := MustOpenIndex()
	defer idx.Close()

	e := NewExecutor(idx.Index, NewCluster(1))
	if _, err := e.Execute("d", MustParse(`set(id=10, frame=f, profile_id=1)`), nil); err != nil {
		t.Fatal(err)
	}

	f := idx.MustCreateFragmentIfNotExists("d", "f", 0)
	if n := f.Bitmap(10).Count(); n != 1 {
		t.Fatalf("unexpected bitmap count: %d", n)
	}
}

// Ensure a remote query can return a bitmap.
func TestExecutor_Execute_Remote_Bitmap(t *testing.T) {
	c := NewCluster(2)

	// Create secondary server and update second cluster node.
	s := NewServer()
	defer s.Close()
	c.Nodes[1].Host = s.Host()

	// Mock secondary server's executor to verify arguments and return a bitmap.
	s.Handler.Executor.ExecuteFn = func(db string, query *pql.Query, slices []uint64) (interface{}, error) {
		if db != `d` {
			t.Fatalf("unexpected db: %s", db)
		} else if query.String() != `get(id=10, frame=f)` {
			t.Fatalf("unexpected query: %s", query.String())
		} else if !reflect.DeepEqual(slices, []uint64{0, 2, 4}) {
			t.Fatalf("unexpected slices: %+v", slices)
		}

		// Set bits in slice 0 & 2.
		bm := pilosa.NewBitmap(
			(0*SliceWidth)+1,
			(0*SliceWidth)+2,
			(2*SliceWidth)+4,
		)
		return bm, nil
	}

	// Create local executor data.
	// The local node owns slice 1.
	idx := MustOpenIndex()
	defer idx.Close()
	idx.MustCreateFragmentIfNotExists("d", "f", 1).MustSetBit(10, (1*SliceWidth)+1)

	e := NewExecutor(idx.Index, c)
	if res, err := e.Execute("d", MustParse(`get(id=10, frame=f)`), nil); err != nil {
		t.Fatal(err)
	} else if chunks := res.(*pilosa.Bitmap).Chunks(); len(chunks) != 3 {
		t.Fatalf("unexpected chunk length: %s", spew.Sdump(chunks))
	} else if chunks[0].Value[0] != 6 {
		t.Fatalf("unexpected chunk(0): %s", spew.Sdump(chunks[0]))
	} else if chunks[1].Value[0] != 2 {
		t.Fatalf("unexpected chunk(1): %s", spew.Sdump(chunks[1]))
	}
}

// Ensure a remote query can return a count.
func TestExecutor_Execute_Remote_Count(t *testing.T) {
	c := NewCluster(2)

	// Create secondary server and update second cluster node.
	s := NewServer()
	defer s.Close()
	c.Nodes[1].Host = s.Host()

	// Mock secondary server's executor to return a count.
	s.Handler.Executor.ExecuteFn = func(db string, query *pql.Query, slices []uint64) (interface{}, error) {
		return uint64(10), nil
	}

	// Create local executor data. The local node owns slice 1.
	idx := MustOpenIndex()
	defer idx.Close()
	idx.MustCreateFragmentIfNotExists("d", "f", 1).MustSetBit(10, (1*SliceWidth)+1)
	idx.MustCreateFragmentIfNotExists("d", "f", 1).MustSetBit(10, (1*SliceWidth)+2)

	e := NewExecutor(idx.Index, c)
	if n, err := e.Execute("d", MustParse(`count(get(id=10, frame=f))`), nil); err != nil {
		t.Fatal(err)
	} else if n != uint64(12) {
		t.Fatalf("unexpected n: %d", n)
	}
}

// Ensure a remote query can set bits on multiple nodes.
func TestExecutor_Execute_Remote_Set(t *testing.T) {
	c := NewCluster(2)
	c.ReplicaN = 2

	// Create secondary server and update second cluster node.
	s := NewServer()
	defer s.Close()
	c.Nodes[1].Host = s.Host()

	// Mock secondary server's executor to verify arguments.
	var remoteCalled bool
	s.Handler.Executor.ExecuteFn = func(db string, query *pql.Query, slices []uint64) (interface{}, error) {
		if db != `d` {
			t.Fatalf("unexpected db: %s", db)
		} else if query.String() != `set(id=10, frame=f, profile_id=2)` {
			t.Fatalf("unexpected query: %s", query.String())
		}
		remoteCalled = true
		return nil, nil
	}

	// Create local executor data.
	idx := MustOpenIndex()
	defer idx.Close()

	e := NewExecutor(idx.Index, c)
	if _, err := e.Execute("d", MustParse(`set(id=10, frame=f, profile_id=2)`), nil); err != nil {
		t.Fatal(err)
	}

	// Verify that one bit is set on both node's index.
	if n := idx.MustCreateFragmentIfNotExists("d", "f", 0).Bitmap(10).Count(); n != 1 {
		t.Fatalf("unexpected local count: %d", n)
	}
	if !remoteCalled {
		t.Fatalf("expected remote execution")
	}
}

// Executor represents a test wrapper for pilosa.Executor.
type Executor struct {
	*pilosa.Executor
}

// NewExecutor returns a new instance of Executor.
// The executor always matches the hostname of the first cluster node.
func NewExecutor(index *pilosa.Index, cluster *pilosa.Cluster) *Executor {
	e := &Executor{Executor: pilosa.NewExecutor(index)}
	e.Cluster = cluster
	e.Host = cluster.Nodes[0].Host
	return e
}

// MustParse parses s into a PQL query. Panic on error.
func MustParse(s string) *pql.Query {
	q, err := pql.NewParser(strings.NewReader(s)).Parse()
	if err != nil {
		panic(err)
	}
	return q
}

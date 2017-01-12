package pilosa_test

import (
	"context"
	"flag"
	"testing"

	"github.com/pilosa/pilosa"
	"github.com/pilosa/pilosa/pql"
)

var plugin = flag.String("plugin", "", "path to 'test' plugin")

// Ensure plugins can be loaded and run.
func TestPlugin(t *testing.T) {
	if !pilosa.PluginsSupported {
		t.Skip("plugins not supported")
	} else if *plugin == "" {
		t.Skip("-plugin not set")
	}

	r := pilosa.NewPluginRegistry()
	if err := r.Load(*plugin); err != nil {
		t.Fatal(err)
	}

	// Instantiate plugin.
	p, err := r.NewPlugin("Debug")
	if err != nil {
		t.Fatal(err)
	}

	// Execute map function.
	// This simply prints out against STDERR to verify that args can be passed.
	p.Map(context.Background(), "d", []pql.Arg{{Key: 1, Value: 2}, {Key: "foo", Value: "bar"}}, 100)
}

// MockPlugin represents a plugin that is implemented as mockable functions.
type MockPlugin struct {
	MapFn    func(ctx context.Context, db string, args []pql.Arg, slice uint64) (interface{}, error)
	ReduceFn func(ctx context.Context, prev, v interface{}) interface{}
}

func (p *MockPlugin) Map(ctx context.Context, db string, args []pql.Arg, slice uint64) (interface{}, error) {
	return p.MapFn(ctx, db, args, slice)
}

func (p *MockPlugin) Reduce(ctx context.Context, prev, v interface{}) interface{} {
	return p.ReduceFn(ctx, prev, v)
}

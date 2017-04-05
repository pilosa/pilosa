package pilosa

import "time"

const (
	// DefaultHost is the default hostname and port to use.
	DefaultHost = "localhost"
	DefaultPort = "10101"
)

// Config represents the configuration for the command.
type Config struct {
	DataDir string `toml:"data-dir"`
	Host    string `toml:"host"`

	Cluster struct {
		ReplicaN        int      `toml:"replicas"`
		Nodes           []string `toml:"hosts"`
		PollingInterval Duration `toml:"polling-interval"`
	} `toml:"cluster"`

	Plugins struct {
		Path string `toml:"path"`
	} `toml:"plugins"`

	AntiEntropy struct {
		Interval Duration `toml:"interval"`
	} `toml:"anti-entropy"`

	LogPath string `toml:"log-path"`
}

// NewConfig returns an instance of Config with default options.
func NewConfig() *Config {
	c := &Config{
		Host: DefaultHost + ":" + DefaultPort,
	}
	c.Cluster.ReplicaN = DefaultReplicaN
	c.Cluster.PollingInterval = Duration(DefaultPollingInterval)
	c.Cluster.Nodes = []string{}
	c.AntiEntropy.Interval = Duration(DefaultAntiEntropyInterval)
	return c
}

func NewConfigForHosts(hosts []string) *Config {
	conf := NewConfig()
	for _, hostport := range hosts {
		conf.Cluster.Nodes = append(conf.Cluster.Nodes, hostport)
	}
	return conf
}

// PilosaCluster returns a new instance of Cluster based on the config.
func (c *Config) PilosaCluster() *Cluster {
	cluster := NewCluster()
	cluster.ReplicaN = c.Cluster.ReplicaN

	for _, hostport := range c.Cluster.Nodes {
		cluster.Nodes = append(cluster.Nodes, &Node{Host: hostport})
	}

	return cluster
}

// Duration is a TOML wrapper type for time.Duration.
type Duration time.Duration

// String returns the string representation of the duration.
func (d Duration) String() string { return time.Duration(d).String() }

// UnmarshalText parses a TOML value into a duration value.
func (d *Duration) UnmarshalText(text []byte) error {
	v, err := time.ParseDuration(string(text))
	if err != nil {
		return err
	}

	*d = Duration(v)
	return nil
}

func (d Duration) MarshalText() (text []byte, err error) {
	return []byte(d.String()), nil
}

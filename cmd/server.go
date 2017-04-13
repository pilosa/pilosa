package cmd

import (
	"fmt"
	"io"
	"os"
	"os/signal"
	"runtime/pprof"
	"time"

	"github.com/spf13/cobra"

	"github.com/pilosa/pilosa/server"
)

// Serve is global so that tests can control and verify it.
var Server *server.Command

func NewServeCmd(stdin io.Reader, stdout, stderr io.Writer) *cobra.Command {
	Server = server.NewCommand(stdin, stdout, stderr)
	serveCmd := &cobra.Command{
		Use:   "server",
		Short: "Run Pilosa.",
		Long: `pilosa server runs Pilosa.

It will load existing data from the configured
directory, and start listening client connections
on the configured port.`,
		RunE: func(cmd *cobra.Command, args []string) error {
			Server.Server.Handler.Version = Version
			fmt.Fprintf(Server.Stderr, "Pilosa %s, build time %s\n", Version, BuildTime)

			// Start CPU profiling.
			if Server.CPUProfile != "" {
				f, err := os.Create(Server.CPUProfile)
				if err != nil {
					return fmt.Errorf("create cpu profile: %v", err)
				}
				defer f.Close()

				fmt.Fprintln(Server.Stderr, "Starting cpu profile")
				pprof.StartCPUProfile(f)
				time.AfterFunc(Server.CPUTime, func() {
					fmt.Fprintln(Server.Stderr, "Stopping cpu profile")
					pprof.StopCPUProfile()
					f.Close()
				})
			}

			// Execute the program.
			if err := Server.Run(); err != nil {
				return fmt.Errorf("error running server: %v", err)
			}

			// First SIGKILL causes server to shut down gracefully.
			c := make(chan os.Signal, 2)
			signal.Notify(c, os.Interrupt)
			select {
			case sig := <-c:
				fmt.Fprintf(Server.Stderr, "Received %s; gracefully shutting down...\n", sig.String())

				// Second signal causes a hard shutdown.
				go func() { <-c; os.Exit(1) }()

				if err := Server.Close(); err != nil {
					return err
				}
			case <-Server.Done:
				fmt.Fprintf(Server.Stderr, "Server closed externally")
			}
			return nil
		},
	}
	flags := serveCmd.Flags()

	flags.StringVarP(&Server.Config.DataDir, "data-dir", "d", "~/.pilosa", "Directory to store pilosa data files.")
	flags.StringVarP(&Server.Config.Host, "bind", "b", ":10101", "Default URI on which pilosa should listen.")
	flags.IntVarP(&Server.Config.Cluster.ReplicaN, "cluster.replicas", "", 1, "Number of hosts each piece of data should be stored on.")
	flags.StringSliceVarP(&Server.Config.Cluster.Nodes, "cluster.hosts", "", []string{}, "Comma separated list of hosts in cluster.")
	flags.DurationVarP((*time.Duration)(&Server.Config.Cluster.PollingInterval), "cluster.poll-interval", "", time.Minute, "Polling interval for cluster.") // TODO what actually is this?
	flags.StringVarP(&Server.Config.Plugins.Path, "plugins.path", "", "", "Path to plugin directory.")
	flags.StringVar(&Server.Config.LogPath, "log-path", "", "Log path")
	flags.DurationVarP((*time.Duration)(&Server.Config.AntiEntropy.Interval), "anti-entropy.interval", "", time.Minute*10, "Interval at which to run anti-entropy routine.")
	flags.StringVarP(&Server.CPUProfile, "profile.cpu", "", "", "Where to store CPU profile.")
	flags.DurationVarP(&Server.CPUTime, "profile.cpu-time", "", 30*time.Second, "CPU profile duration.")
	flags.StringVarP(&Server.Config.Cluster.MessengerType, "cluster.messenger-type", "", "static", "Type of Messenger to use for inter-host messaging. Choose from [static, broadcast, gossip]")
	flags.StringVarP(&Server.Config.Cluster.Gossip.Seed, "cluster.gossip.seed", "", "", "Host with which to seed the gossip membership.")
	flags.IntVarP(&Server.Config.Cluster.Gossip.Port, "cluster.gossip.port", "", 0, "Port to which pilosa should bind for gossip.")
	flags.StringVarP(&Server.Config.Metric.Service, "metric.service", "", "noop", "Default URI on which pilosa should listen.")
	flags.StringVarP(&Server.Config.Metric.Host, "metric.host", "", "", "Default URI to send metrics.")
	flags.DurationVarP((*time.Duration)(&Server.Config.Metric.PollingInterval), "metric.poll-interval", "", time.Minute*0, "Polling interval metrics.")

	return serveCmd
}

func init() {
	subcommandFns["server"] = NewServeCmd
}

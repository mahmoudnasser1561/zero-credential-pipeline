// Package app is the wiring layer: it owns CLI flags, the extension-point
// registries (which stack, which discovery mechanism, which execution
// model), and the orchestration that ties them to the generic check engine.
// main.go stays a slim trigger; everything that decides what actually runs
// lives here.
package app

import (
	"context"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ssm"

	"github.com/mahmoudnasser1561/zero-credential-pipeline/stackcheck/internal/check"
	"github.com/mahmoudnasser1561/zero-credential-pipeline/stackcheck/internal/checks/observability"
	"github.com/mahmoudnasser1561/zero-credential-pipeline/stackcheck/internal/discovery"
	"github.com/mahmoudnasser1561/zero-credential-pipeline/stackcheck/internal/executor"
	"github.com/mahmoudnasser1561/zero-credential-pipeline/stackcheck/internal/report"
)

type Config struct {
	Region    string
	TagKey    string
	TagValue  string
	Stack     string
	Discovery string
	Executor  string
	Format    string
	Timeout   time.Duration
}

// ExecutorFactory binds a networking-model-specific Executor implementation
// to one discovered target.
type ExecutorFactory func(target string) executor.Executor

// StackFactory returns every check for one stack, bound to one target's
// Executor. The sole extension point for verifying a different piece of
// software later.
type StackFactory func(exec executor.Executor, target string) []check.Checker

// discoverers is the extension point for adding a new kind of target
// environment (a Kubernetes cluster, a different cloud, ...). Deliberately
// independent of executors: how you find a target and how you run commands
// against it are different concerns, selected by different flags.
var discoverers = map[string]func(ctx context.Context, cfg Config) (discovery.Discoverer, error){
	"ec2-tag": newEC2TagDiscoverer,
}

// executors is the extension point for adding a new networking/reachability
// model (SSH, kubectl exec, ...). Deliberately independent of discoverers.
var executors = map[string]func(ctx context.Context, cfg Config) (ExecutorFactory, error){
	"ssm": newSSMExecutorFactory,
}

// stacks is the extension point for verifying a different piece of software.
var stacks = map[string]StackFactory{
	"observability": observability.AllChecks,
}

func newEC2TagDiscoverer(ctx context.Context, cfg Config) (discovery.Discoverer, error) {
	awsCfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(cfg.Region))
	if err != nil {
		return nil, fmt.Errorf("load aws config: %w", err)
	}
	return discovery.NewEC2Tag(ec2.NewFromConfig(awsCfg), cfg.TagKey, cfg.TagValue), nil
}

func newSSMExecutorFactory(ctx context.Context, cfg Config) (ExecutorFactory, error) {
	awsCfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(cfg.Region))
	if err != nil {
		return nil, fmt.Errorf("load aws config: %w", err)
	}
	ssmClient := ssm.NewFromConfig(awsCfg)
	return func(target string) executor.Executor {
		return executor.NewSSM(ssmClient, target)
	}, nil
}

// Run parses args, wires the selected stack/discovery/executor together,
// executes every check, prints the report, and returns the process exit
// code.
func Run(args []string) int {
	cfg, err := parseFlags(args)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 2
	}

	ctx, cancel := context.WithTimeout(context.Background(), cfg.Timeout)
	defer cancel()

	newDiscoverer, ok := discoverers[cfg.Discovery]
	if !ok {
		fmt.Fprintf(os.Stderr, "unknown --discovery %q\n", cfg.Discovery)
		return 2
	}
	newExecutorFactory, ok := executors[cfg.Executor]
	if !ok {
		fmt.Fprintf(os.Stderr, "unknown --executor %q\n", cfg.Executor)
		return 2
	}
	stackChecks, ok := stacks[cfg.Stack]
	if !ok {
		fmt.Fprintf(os.Stderr, "unknown --stack %q\n", cfg.Stack)
		return 2
	}

	disc, err := newDiscoverer(ctx, cfg)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 2
	}
	execFactory, err := newExecutorFactory(ctx, cfg)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 2
	}

	targets, err := disc.Discover(ctx)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 2
	}

	var checkers []check.Checker
	for _, t := range targets {
		checkers = append(checkers, stackChecks(execFactory(t), t)...)
	}

	results := check.NewRunner().Run(ctx, checkers)

	var reporter report.Reporter
	switch cfg.Format {
	case "text":
		reporter = report.NewTable()
	default:
		reporter = report.NewJSON()
	}

	fmt.Println(reporter.Render(results))

	return report.ExitCode(results)
}

func parseFlags(args []string) (Config, error) {
	fs := flag.NewFlagSet("stackcheck", flag.ContinueOnError)

	region := fs.String("region", "us-east-1", "AWS region")
	tag := fs.String("tag", "", "discovery selector as key=value, e.g. Project=zero-credential-pipeline")
	stack := fs.String("stack", "observability", "which stack to verify (see stacks registry)")
	discoveryType := fs.String("discovery", "ec2-tag", "how to find targets (see discoverers registry)")
	executorType := fs.String("executor", "ssm", "how to reach a target and run commands (see executors registry)")
	format := fs.String("format", "json", "output format: json (default) or text")
	timeout := fs.Duration("timeout", 90*time.Second, "overall deadline for discovery and checks")

	if err := fs.Parse(args); err != nil {
		return Config{}, err
	}

	key, value, found := strings.Cut(*tag, "=")
	if !found {
		return Config{}, fmt.Errorf("--tag must be in key=value form, got %q", *tag)
	}

	return Config{
		Region:    *region,
		TagKey:    key,
		TagValue:  value,
		Stack:     *stack,
		Discovery: *discoveryType,
		Executor:  *executorType,
		Format:    *format,
		Timeout:   *timeout,
	}, nil
}

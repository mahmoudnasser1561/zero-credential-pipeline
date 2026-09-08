package main

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
	Region     string
	TagKey     string
	TagValue   string
	TargetType string
	Format     string
	Timeout    time.Duration
}

// ExecutorFactory binds a target-agnostic Executor implementation to one
// discovered target. Kept separate from Discoverer so a new target type
// only needs one new entry in targetTypes, not a change to how targets are
// discovered vs. how they're executed against.
type ExecutorFactory func(target string) executor.Executor

// targetTypes is the sole extension point for adding a new kind of target
// environment later (a Kubernetes cluster, a different cloud, etc.) — one
// new map entry, nothing else in this file or in internal/check changes.
var targetTypes = map[string]func(ctx context.Context, cfg Config) (discovery.Discoverer, ExecutorFactory, error){
	"ec2": newEC2Target,
}

func newEC2Target(ctx context.Context, cfg Config) (discovery.Discoverer, ExecutorFactory, error) {
	awsCfg, err := config.LoadDefaultConfig(ctx, config.WithRegion(cfg.Region))
	if err != nil {
		return nil, nil, fmt.Errorf("load aws config: %w", err)
	}

	ec2Client := ec2.NewFromConfig(awsCfg)
	ssmClient := ssm.NewFromConfig(awsCfg)

	disc := discovery.NewEC2Tag(ec2Client, cfg.TagKey, cfg.TagValue)
	execFactory := func(target string) executor.Executor {
		return executor.NewSSM(ssmClient, target)
	}

	return disc, execFactory, nil
}

func main() {
	os.Exit(run())
}

func run() int {
	cfg, err := parseFlags(os.Args[1:])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		return 2
	}

	ctx, cancel := context.WithTimeout(context.Background(), cfg.Timeout)
	defer cancel()

	newTarget, ok := targetTypes[cfg.TargetType]
	if !ok {
		fmt.Fprintf(os.Stderr, "unknown --target-type %q\n", cfg.TargetType)
		return 2
	}

	disc, execFactory, err := newTarget(ctx, cfg)
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
		checkers = append(checkers, observability.AllChecks(execFactory(t), t)...)
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
	tag := fs.String("tag", "", "target selector as key=value, e.g. Project=zero-credential-pipeline")
	targetType := fs.String("target-type", "ec2", "target environment type (see targetTypes registry)")
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
		Region:     *region,
		TagKey:     key,
		TagValue:   value,
		TargetType: *targetType,
		Format:     *format,
		Timeout:    *timeout,
	}, nil
}

package executor

import (
	"context"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/aws/aws-sdk-go-v2/service/ssm/types"
)

type SSMExecutor struct {
	client   *ssm.Client
	target   string
	poll     time.Duration
	deadline time.Duration
}

func NewSSM(client *ssm.Client, target string) *SSMExecutor {
	return &SSMExecutor{
		client:   client,
		target:   target,
		poll:     2 * time.Second,
		deadline: 60 * time.Second,
	}
}

func (e *SSMExecutor) Run(ctx context.Context, command string) (Output, error) {
	sendOut, err := e.client.SendCommand(ctx, &ssm.SendCommandInput{
		DocumentName: aws.String("AWS-RunShellScript"),
		InstanceIds:  []string{e.target},
		Parameters: map[string][]string{
			"commands": {command},
		},
	})
	if err != nil {
		return Output{}, fmt.Errorf("ssm send-command: %w", err)
	}

	commandID := aws.ToString(sendOut.Command.CommandId)

	ctx, cancel := context.WithTimeout(ctx, e.deadline)
	defer cancel()

	for {
		invOut, err := e.client.GetCommandInvocation(ctx, &ssm.GetCommandInvocationInput{
			CommandId:  aws.String(commandID),
			InstanceId: aws.String(e.target),
		})
		if err != nil {
			// the invocation record can take a moment to appear after
			// SendCommand returns; keep polling until the deadline
			select {
			case <-ctx.Done():
				return Output{}, fmt.Errorf("ssm get-command-invocation: %w", ctx.Err())
			case <-time.After(e.poll):
				continue
			}
		}

		switch invOut.Status {
		case types.CommandInvocationStatusSuccess,
			types.CommandInvocationStatusFailed,
			types.CommandInvocationStatusCancelled,
			types.CommandInvocationStatusTimedOut:
			return Output{
				Stdout:   aws.ToString(invOut.StandardOutputContent),
				Stderr:   aws.ToString(invOut.StandardErrorContent),
				ExitCode: invOut.ResponseCode,
			}, nil
		}

		select {
		case <-ctx.Done():
			return Output{}, fmt.Errorf("ssm command %s did not complete before deadline", commandID)
		case <-time.After(e.poll):
		}
	}
}

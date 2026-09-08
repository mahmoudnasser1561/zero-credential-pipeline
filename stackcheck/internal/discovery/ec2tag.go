package discovery

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

type EC2TagDiscoverer struct {
	client   *ec2.Client
	tagKey   string
	tagValue string
}

func NewEC2Tag(client *ec2.Client, tagKey, tagValue string) *EC2TagDiscoverer {
	return &EC2TagDiscoverer{client: client, tagKey: tagKey, tagValue: tagValue}
}

func (d *EC2TagDiscoverer) Discover(ctx context.Context) ([]string, error) {
	out, err := d.client.DescribeInstances(ctx, &ec2.DescribeInstancesInput{
		Filters: []types.Filter{
			{Name: aws.String(fmt.Sprintf("tag:%s", d.tagKey)), Values: []string{d.tagValue}},
			{Name: aws.String("instance-state-name"), Values: []string{"running"}},
		},
	})
	if err != nil {
		return nil, fmt.Errorf("ec2 describe-instances: %w", err)
	}

	var targets []string
	for _, reservation := range out.Reservations {
		for _, instance := range reservation.Instances {
			targets = append(targets, aws.ToString(instance.InstanceId))
		}
	}

	if len(targets) == 0 {
		return nil, fmt.Errorf("no running instances found with tag %s=%s", d.tagKey, d.tagValue)
	}

	return targets, nil
}

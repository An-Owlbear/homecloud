package dns

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/route53"
	"github.com/aws/aws-sdk-go-v2/service/route53/types"
)

var HostedZoneID = os.Getenv("HOSTED_ZONE_ID")

var MissingRecordsError = errors.New("expected DNS records not found")

func New(ctx context.Context) (*route53.Client, error) {
	cfg, err := config.LoadDefaultConfig(ctx)
	if err != nil {
		return nil, err
	}

	return route53.NewFromConfig(cfg), nil
}

// SetRecord sets the specified DNS record for the given subdomain and IP address
func SetRecord(ctx context.Context, client *route53.Client, subdomainBase string, address string) error {
	hostedZone, err := client.GetHostedZone(ctx, &route53.GetHostedZoneInput{
		Id: aws.String(HostedZoneID),
	})
	if err != nil {
		return err
	}

	for _, subdomain := range []string{subdomainBase, "*." + subdomainBase} {
		_, err = client.ChangeResourceRecordSets(ctx, &route53.ChangeResourceRecordSetsInput{
			HostedZoneId: aws.String(HostedZoneID),
			ChangeBatch: &types.ChangeBatch{
				Changes: []types.Change{
					{
						Action: types.ChangeActionUpsert,
						ResourceRecordSet: &types.ResourceRecordSet{
							Name: aws.String(fmt.Sprintf("%s.%s", subdomain, *hostedZone.HostedZone.Name)),
							Type: types.RRTypeA,
							TTL:  aws.Int64(300),
							ResourceRecords: []types.ResourceRecord{
								{
									Value: aws.String(address),
								},
							},
						},
					},
				},
			},
		})
	}

	if err != nil {
		return err
	}

	return nil
}

// RemoveRecord removes the DNS record for the given subdomain and IP address
func RemoveRecord(ctx context.Context, client *route53.Client, subdomainBase string, address string) error {
	hostedZone, err := client.GetHostedZone(ctx, &route53.GetHostedZoneInput{
		Id: aws.String(HostedZoneID),
	})
	if err != nil {
		return err
	}

	for _, subdomain := range []string{subdomainBase, "*." + subdomainBase} {
		fqdn := fmt.Sprintf("%s.%s", subdomain, *hostedZone.HostedZone.Name)

		resourceRecords, err := client.ListResourceRecordSets(ctx, &route53.ListResourceRecordSetsInput{
			HostedZoneId:    aws.String(HostedZoneID),
			StartRecordName: aws.String(fqdn),
			StartRecordType: types.RRTypeA,
			MaxItems:        aws.Int32(1),
		})
		if err != nil {
			return err
		}
		if len(resourceRecords.ResourceRecordSets) == 0 {
			return MissingRecordsError
		}

		_, err = client.ChangeResourceRecordSets(ctx, &route53.ChangeResourceRecordSetsInput{
			HostedZoneId: aws.String(HostedZoneID),
			ChangeBatch: &types.ChangeBatch{
				Changes: []types.Change{
					{
						Action:            types.ChangeActionDelete,
						ResourceRecordSet: &resourceRecords.ResourceRecordSets[0],
					},
				},
			},
		})
		if err != nil {
			return err
		}
	}

	return nil
}

# Homecloud DNS Functions

This folder contains the code the AWS lambda functions.

## Deployment

For deployment, you will need:
- The aws cli installed
- A Lambda function of the same name to deploy to
- Permission update Lambda functions
- A domain name registered through Route53 with a hosted zone configured
- A DynamoDB database setup
- [just](https://just.systems)

The Lambda functions themselves, environment variables and permissions will
need to be set manually. For any of the functions using hashing execution time
will also need to be increased.

To deploy, run the command:
```shell
just deploy [function] [profile] [region]
```
Where [function] is the name of the function to deploy, [profile] is the name
of the local profile to use, and [region] is the AWS region to deploy to.

The names of the functions to deploy are the names of the folders in `cmd`
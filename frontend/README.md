# Homecloud Frontend

This folder cotnains all the code the for the main frontend for Homecloud.

## Building

The build process is handled by the Homecloud Dockerfile, see instructions at
[backend/README.md](../backend/README.md#building) for more details.

## Running in development

For a smoother development process the code should be sent to the test device
and run there.

For this you will need
- `pnpm` and `node` installed on the test device
- The Homecloud backend deployed on the test device
- SSH access to the test device

For this the test device will need `pnpm` and `node` installed.

The script `pnpm run deploy_dev` can be used to sync the data with the test
device. This script assumes it uses the domain `homecloud.local` and has the
user `homecloud`. For different values the script must be modified.
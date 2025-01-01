# homecloud backend

The homecloud backend manages the applications and access to them.

## Development Setup

A docker instance is required for developing and running the application. Put
the location of the socket file or the address in the .env file to use it.

The backend is only designed to run inside docker, as the networking for the
reverse proxy would be much more complex otherwise. Only targeting this single
also reduces the discrepencies between the development and deployment
environments.

[just](https://just.systems) is used for running the build configurations.

### MacOS Recommended Docker Setup

For good performance on MacOS a more specific setup is required.

Docker Desktop is recommended for running Docker, colima and rancher desktop can
also be used but require more precise configuration to work properly with good
performance for this application.

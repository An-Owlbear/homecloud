# Homecloud backend

The homecloud backend manages the applications and access to them.

## Directory Structure

The key structure of the code in this directory is as follows:

- assets: contains static web assets (css/js)
- cmd: contains code used to produce executables
  - homecloud: produces the main backend executable
  - launcher: produces the set/launcher executable
- internal: contains the majority of the application code
- migrations: contains SQL migration files
- ory_config: contains ory configuration files
- queries: contains SQL query files used by sqlc to autogenerate Go code

## Building

[just](https://just.systems) is used for build scripts for this project.

This project contains code the only runs on Linux, and can't be compiled for
other operating systems. Thw SQLite driver makes cross compilation difficult,
so the setup will likely be operating system and architecture dependent.

The backend is designed to only run in a Docker container, so by building
during the container build the operating system issue is bypassed. A container
image can be built with the following command:

```shell
just build_docker [version]
```
Where version is the version to assign to the build.

The launcher can be built using the command

```shell
just build_launcher [version] [os] [architecture]
```
Where [version] is the version number to assign, [os] is the OS to compile for
and [architecture] is the architecture to compile for.

The launcher doesn't use SQLite so the cross-compilation issue does not apply
here.

## Running

The justfile provides scripts for deploying the program to a target test
environment. This test environment must meet the following conditions:
- Be running a Debian based OS
- Be available over SSH
- Must have Docker installed
- Must have superuser access

### Development

The docker container can be sent to the target device by running
```shell
just deploy_docker_dev [address] [version] [port]
```
Where [address] is the ssh address of the device, [version] is the version of
the container and [port] is the port to use.

The deploy the launcher run
```shell
just deploy_launcher_dev [address] [port]
```

### Production Testing

The whole system can be deployed to a target device with productions settings
using the command
```shell
just build_deploy_launcher_test_prod [version] [address] [port]
```


### MacOS Recommended Docker Setup

For good performance on MacOS a more specific setup is required.

Docker Desktop is recommended for running Docker, colima and rancher desktop can
also be used but require more precise configuration to work properly with good
performance for this application.

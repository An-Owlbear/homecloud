# Homecloud

Homecloud is a system made to make setting up a Home server as easy as possible, and not require technical configuration.

This repository contains the code for all aspects of the project, including the server code and the code for the few cloud functions used for DNS management.

This repository and this page are not intended for end users, and contain technical details on the program and how to build it.

## Repository Structure

Each directory in the repository root is for a different part of the program:
- apps: contains app packages. By being accessible through GitHub this also acts as a public repository
- backend: the backend for the Homecloud system. Manages apps and users, etc. Also contains code for the setup/launcher app since it shares much of its code with the backend
- dns: AWS lambda functions for managing DNS records
- frontend: the main application frontend, written in SvelteKit
- performance_tests: tests used to verify the resource usage of the application
- setup-frontend: the frontend for the setup application

## Building

Build instructions for the system can be found in [backend/README.md](backend/README.md)
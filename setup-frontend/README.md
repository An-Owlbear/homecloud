# Homecloud Setup Frontend

This folder contains all code related to the Homecloud setup frontend.

## Building

Similar to the main frontend, the build process for this is handled by scripts
 as detailed in [backend//README.md](../backend/README.md)

This component be tested locally more easily than the main frontend, but has a
few requirements:
- The launcher API is available at http://homecloud.local:1234
- For testing the update API it must be running on a Debian system
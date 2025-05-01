#!/bin/bash

# Deletes all rocker resources. Only should be used to reset a test device

echo 'Stopping containers';
docker stop $(docker ps -a -q);
echo 'Removing containers';
docker rm $(docker ps -a -q);
echo 'Removing volumes';
docker volume rm $(docker volume ls -q);
echo 'Removing networks';
docker network rm $(docker network ls -q);

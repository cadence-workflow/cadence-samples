#!/bin/bash

# Script to list workflows in the test-domain

docker run --network=docker-samples_cadence-network --rm ubercadence/cli:master \
    --address cadence:7833 \
    --domain test-domain workflow list




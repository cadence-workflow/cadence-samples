#!/bin/bash

# Cadence Samples Quick Start Script
# This script helps you quickly start and test the Cadence samples

set -e

SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
cd "$SCRIPT_DIR"

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${BLUE}╔════════════════════════════════════════╗${NC}"
echo -e "${BLUE}║   Cadence Samples Quick Start         ║${NC}"
echo -e "${BLUE}╚════════════════════════════════════════╝${NC}"
echo ""

# Step 1: Start services
echo -e "${GREEN}Step 1:${NC} Starting Cadence server and workers..."
docker-compose up -d

echo -e "${YELLOW}Waiting for Cadence server to be ready...${NC}"
sleep 10

# Check if Cadence is up
until docker-compose exec -T cadence sh -c "nc -z localhost 7833" 2>/dev/null; do
    echo -e "${YELLOW}Still waiting for Cadence...${NC}"
    sleep 5
done

echo -e "${GREEN}✓ Cadence server is ready!${NC}"
echo ""

# Step 2: Register domain
echo -e "${GREEN}Step 2:${NC} Registering test domain..."
docker run --network=docker-samples_cadence-network --rm ubercadence/cli:master \
    --address cadence:7833 \
    --domain test-domain domain register 2>/dev/null || echo -e "${YELLOW}Domain already exists (OK)${NC}"

echo -e "${GREEN}✓ Domain registered!${NC}"
echo ""

# Step 3: Wait for worker
echo -e "${GREEN}Step 3:${NC} Waiting for worker to be ready..."
sleep 5

# Check worker logs
echo -e "${BLUE}Worker logs:${NC}"
docker-compose logs --tail=10 hello-world-worker

echo ""

# Step 4: Trigger sample workflow
echo -e "${GREEN}Step 4:${NC} Triggering hello world workflow..."
docker run --network=docker-samples_cadence-network --rm ubercadence/cli:master \
    --address cadence:7833 \
    --domain test-domain workflow start \
    --tasklist test-worker \
    --workflow_type main.helloWorldWorkflow \
    --execution_timeout 60 \
    --input '"Docker Quick Start"'

echo -e "${GREEN}✓ Workflow triggered!${NC}"
echo ""

# Step 5: Show results
sleep 2
echo -e "${GREEN}Step 5:${NC} Checking worker logs for results..."
echo -e "${BLUE}Recent worker activity:${NC}"
docker-compose logs --tail=20 hello-world-worker | grep -E "(Hello|workflow|activity)" || true

echo ""
echo -e "${GREEN}╔════════════════════════════════════════╗${NC}"
echo -e "${GREEN}║   🎉 Quick Start Complete!            ║${NC}"
echo -e "${GREEN}╚════════════════════════════════════════╝${NC}"
echo ""
echo -e "${BLUE}Next steps:${NC}"
echo -e "  • View Web UI: ${GREEN}http://localhost:8088${NC} (domain: test-domain)"
echo -e "  • View worker logs: ${GREEN}docker-compose logs -f hello-world-worker${NC}"
echo -e "  • List workflows: ${GREEN}./list-workflows.sh${NC}"
echo -e "  • Stop services: ${GREEN}docker-compose down${NC}"
echo ""
echo -e "${BLUE}Trigger more workflows:${NC}"
echo -e "  ${GREEN}docker run --network=docker-samples_cadence-network --rm ubercadence/cli:master \\${NC}"
echo -e "    ${GREEN}--address cadence:7833 --domain test-domain workflow start \\${NC}"
echo -e "    ${GREEN}--tasklist test-worker --workflow_type main.helloWorldWorkflow \\${NC}"
echo -e "    ${GREEN}--execution_timeout 60 --input '\"Your Name\"'${NC}"
echo ""




#!/bin/bash
# This test is heavily based on /test/integration/suites/join-token in the spire package

echo "Start spire server"
docker-compose up -d spire-server

echo "Bootstrap agent"
docker-compose exec -T spire-server \
    /opt/spire/bin/spire-server bundle show > conf/agent/bootstrap.crt

echo "generating join token..."
TOKEN=$(docker-compose exec -T spire-server \
    /opt/spire/bin/spire-server token generate -spiffeID spiffe://domain.test/node | awk '{print $2}' | tr -d '\r')

echo "using join token ${TOKEN}..."
cat conf/agent/agent.conf.template | sed "s#TOKEN#${TOKEN}#g" > conf/agent/agent.conf

echo "start the spire agent"
docker-compose up -d spire-agent

echo "creating registration entry..."
docker-compose exec -T spire-server \
    /opt/spire/bin/spire-server entry create \
    -parentID "spiffe://domain.test/node" \
    -spiffeID "spiffe://domain.test/workload" \
    -selector "unix:uid:0" \
    -ttl 0

# Check at most 30 times (with one second in between) that the agent has
# successfully synced down the workload entry.
MAXCHECKS=30
CHECKINTERVAL=1
COMPLETE=0
for ((i=1;i<=MAXCHECKS;i++)); do
    echo "checking for synced workload entry ($i of $MAXCHECKS max)..."
    docker-compose logs spire-agent
    if docker-compose logs spire-agent | grep "spiffe://domain.test/workload"; then
        COMPLETE=1
        break
    fi
    sleep "${CHECKINTERVAL}"
done

if [[ $COMPLETE -eq 0 ]]; then
echo "timed out waiting for agent to sync down entry"
fi

echo "checking X509-SVID..."
docker-compose exec -T spire-agent \
    /opt/spire/bin/spire-agent api fetch x509 -socketPath /socket/socket/api.sock || (echo "Agent registration didn't work"; exit)


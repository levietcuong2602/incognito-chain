#!/usr/bin/env bash
#GrafanaURL=http://128.199.96.206:8086/write?db=mydb
###### MULTI_MEMBERS
# Shard 0
if [ "$1" == "shard0-0" ]; then
INCOGNITO_NETWORK_KEY=testnet ./incognito --discoverpeersaddress "0.0.0.0:9330" --miningkeys "126jUGQ7T6WBNTB2hM1BQZF9n96Xumfmwxtw62G2wB9ZUpAdaMA" --nodemode "auto" --datadir "data/shard0-0" --listen "0.0.0.0:9434" --externaladdress "0.0.0.0:9434" --norpcauth --rpclisten "0.0.0.0:9334" --loglevel debug
fi
if [ "$1" == "shard0-1" ]; then
INCOGNITO_NETWORK_KEY=testnet ./incognito --discoverpeersaddress "0.0.0.0:9330" --miningkeys "1GR564zSAhySsQaR4SFDgGeBpEjfvBF5DiU7h5KmjzW7YP6fxq" --nodemode "auto" --datadir "data/shard0-1" --listen "0.0.0.0:9435" --externaladdress "0.0.0.0:9435" --norpcauth --rpclisten "0.0.0.0:9335" --loglevel debug
fi
if [ "$1" == "shard0-2" ]; then
INCOGNITO_NETWORK_KEY=testnet ./incognito --discoverpeersaddress "0.0.0.0:9330" --miningkeys "1wGHt3i6JhvSngyBZgdETrmkmQ5KmxPeC9yR2gq1dLadKWQVwd" --nodemode "auto" --datadir "data/shard0-2" --listen "0.0.0.0:9436" --externaladdress "0.0.0.0:9436" --norpcauth --rpclisten "0.0.0.0:9336" --loglevel debug
fi
if [ "$1" == "shard0-3" ]; then
INCOGNITO_NETWORK_KEY=testnet ./incognito --discoverpeersaddress "0.0.0.0:9330" --miningkeys "12bvvmdQ2HpD7yL8d4TBjxcivkVzArT4C62CjT6yNmNvcF8MRcH" --nodemode "auto" --datadir "data/shard0-3" --listen "0.0.0.0:9437" --externaladdress "0.0.0.0:9437" --norpcauth --rpclisten "0.0.0.0:9337" --loglevel debug
fi
# Shard 1
if [ "$1" == "shard1-0" ]; then
INCOGNITO_NETWORK_KEY=testnet ./incognito --discoverpeersaddress "0.0.0.0:9330" --miningkeys "129w9e1EWJZCCcc5xUm4UqFZYQ6JYuQmL5FCArave7y9Bzn2cYY" --nodemode "auto" --datadir "data/shard1-0" --listen "0.0.0.0:9438" --externaladdress "0.0.0.0:9438" --norpcauth --rpclisten "0.0.0.0:9338" --loglevel debug
fi
if [ "$1" == "shard1-1" ]; then
INCOGNITO_NETWORK_KEY=testnet ./incognito --discoverpeersaddress "0.0.0.0:9330" --miningkeys "1LCJLmBekfhnubERtwAF1rquwnQm5VywfUxfnLQkshjFQZ7ANf" --nodemode "auto" --datadir "data/shard1-1" --listen "0.0.0.0:9439" --externaladdress "0.0.0.0:9439" --norpcauth --rpclisten "0.0.0.0:9339" --loglevel debug
fi
if [ "$1" == "shard1-2" ]; then
INCOGNITO_NETWORK_KEY=testnet ./incognito --discoverpeersaddress "0.0.0.0:9330" --miningkeys "12sfSuYBWVYQ1cCeWKAqRc9w7qv9BJm5aM4YGwHDYPFiLfQQj5f" --nodemode "auto" --datadir "data/shard1-2" --listen "0.0.0.0:9440" --externaladdress "0.0.0.0:9440" --norpcauth --rpclisten "0.0.0.0:9340" --loglevel debug
fi
if [ "$1" == "shard1-3" ]; then
INCOGNITO_NETWORK_KEY=testnet ./incognito --discoverpeersaddress "0.0.0.0:9330" --miningkeys "12EKZk8uLXRRZzsBYSnbVL4d6m4mqNwxgR9DW4gJvTGX2MAC2Fa" --nodemode "auto" --datadir "data/shard1-3" --listen "0.0.0.0:9441" --externaladdress "0.0.0.0:9441" --norpcauth --rpclisten "0.0.0.0:9341" --loglevel debug
fi
# Beacon
if [ "$1" == "beacon-0" ]; then
INCOGNITO_NETWORK_KEY=testnet ./incognito --discoverpeersaddress "0.0.0.0:9330" --miningkeys "12wGVmW2P7AF5fhMLW9xjv1zyGB3weT3th1AQ7fDoijENggbd4V" --nodemode "auto" --datadir "data/beacon-0" --listen "0.0.0.0:9450" --externaladdress "0.0.0.0:9450" --norpcauth --rpclisten "0.0.0.0:9350" --loglevel debug
fi
if [ "$1" == "beacon-1" ]; then
INCOGNITO_NETWORK_KEY=testnet ./incognito --discoverpeersaddress "0.0.0.0:9330" --miningkeys "1SkeQ9AgMGbNPQQ7jQSL5bpHYcGeWiNfwft1s7tE6tUz6qSUcV" --nodemode "auto" --datadir "data/beacon-1" --listen "0.0.0.0:9451" --externaladdress "0.0.0.0:9451" --norpcauth --rpclisten "0.0.0.0:9351" --loglevel debug
fi
if [ "$1" == "beacon-2" ]; then
INCOGNITO_NETWORK_KEY=testnet ./incognito --discoverpeersaddress "0.0.0.0:9330" --miningkeys "1hJACmjc9JrVzuhSQSaDvTVkJkif6TZ4QqBxtZPtsKSkovEUUW" --nodemode "auto" --datadir "data/beacon-2" --listen "0.0.0.0:9452" --externaladdress "0.0.0.0:9452" --norpcauth --rpclisten "0.0.0.0:9352" --loglevel debug
fi
if [ "$1" == "beacon-3" ]; then
INCOGNITO_NETWORK_KEY=testnet ./incognito --discoverpeersaddress "0.0.0.0:9330" --miningkeys "1287VLGKX4u9tLkLbZJpQBzBrpoG6EfwT2WnuX2SKGTZLdyBQ35" --nodemode "auto" --datadir "data/beacon-3" --listen "0.0.0.0:9453" --externaladdress "0.0.0.0:9453" --norpcauth --rpclisten "0.0.0.0:9353" --loglevel debug
fi
# FullNode
if [ "$1" == "full_node" ]; then
INCOGNITO_NETWORK_KEY=testnet ./incognito --discoverpeersaddress "0.0.0.0:9330" --nodemode "relay" --datadir "data/full_node" --listen "0.0.0.0:9454" --externaladdress "0.0.0.0:9454" --norpcauth --rpclisten "0.0.0.0:9354" --relayshards "all"  --txpoolmaxtx 100000 --loglevel debug
fi


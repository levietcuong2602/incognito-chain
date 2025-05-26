#!/usr/bin/env bash
#GrafanaURL=http://128.199.96.206:8086/write?db=mydb
###### MULTI_MEMBERS
# Shard 0
if [ "$1" == "shard0-0" ]; then
INCOGNITO_NETWORK_KEY=local ./incognito --discoverpeersaddress "0.0.0.0:9330" --privatekey "112t8rnfXYskvWnHAXKs8dXLtactxRqpPTYJ6PzwkVHnF1begkenMviATTJVM6gVAgSdXsN5DEpTkLFPHtFVnS5RePi6aqTStdpb3St3uRni" --nodemode "auto" --datadir "data/shard0-0" --listen "0.0.0.0:9434" --externaladdress "0.0.0.0:9434" --norpcauth --rpclisten "0.0.0.0:9334" --rpcwslisten "0.0.0.0:19334" 
INCOGNITO_NETWORK_KEY=local ./incognito --discoverpeersaddress "0.0.0.0:9330" --privatekey "112t8rnfXYskvWnHAXKs8dXLtactxRqpPTYJ6PzwkVHnF1begkenMviATTJVM6gVAgSdXsN5DEpTkLFPHtFVnS5RePi6aqTStdpb3St3uRni" --nodemode "auto" --datadir "data/shard0-0" --listen "0.0.0.0:9434" --externaladdress "0.0.0.0:9434" --norpcauth --rpclisten "0.0.0.0:9334" --rpcwslisten "0.0.0.0:19334" 
fi
if [ "$1" == "shard0-1" ]; then
INCOGNITO_NETWORK_KEY=local ./incognito --discoverpeersaddress "0.0.0.0:9330" --privatekey "112t8rngZ1rZ3eWHZucwf9vrpD1DNUAmrTTARSsptNDFrEoHv3QsDY3dZe8LXy3GeKXmeso8nUPsNwUM2qmZibQVXxGzstF4v4vbfQvgk5Ci" --nodemode "auto" --datadir "data/shard0-1" --listen "0.0.0.0:9435" --externaladdress "0.0.0.0:9435" --norpcauth --rpclisten "0.0.0.0:9335" --rpcwslisten "0.0.0.0:19335" 
INCOGNITO_NETWORK_KEY=local ./incognito --discoverpeersaddress "0.0.0.0:9330" --privatekey "112t8rngZ1rZ3eWHZucwf9vrpD1DNUAmrTTARSsptNDFrEoHv3QsDY3dZe8LXy3GeKXmeso8nUPsNwUM2qmZibQVXxGzstF4v4vbfQvgk5Ci" --nodemode "auto" --datadir "data/shard0-1" --listen "0.0.0.0:9435" --externaladdress "0.0.0.0:9435" --norpcauth --rpclisten "0.0.0.0:9335" --rpcwslisten "0.0.0.0:19335" 
fi
if [ "$1" == "shard0-2" ]; then
INCOGNITO_NETWORK_KEY=local ./incognito --discoverpeersaddress "0.0.0.0:9330" --privatekey "112t8rnpXg6CLjvBg2ZiyMDgpgQoZuAjYGzbm6b2eXVSHUKjZUyb2LVJmJDPw4yNaP5M14DomzC514joTH3EVknRwnnGViWuH1QucebGtVxd" --nodemode "auto" --datadir "data/shard0-2" --listen "0.0.0.0:9436" --externaladdress "0.0.0.0:9436" --norpcauth --rpclisten "0.0.0.0:9336" --rpcwslisten "0.0.0.0:19336" 
INCOGNITO_NETWORK_KEY=local ./incognito --discoverpeersaddress "0.0.0.0:9330" --privatekey "112t8rnpXg6CLjvBg2ZiyMDgpgQoZuAjYGzbm6b2eXVSHUKjZUyb2LVJmJDPw4yNaP5M14DomzC514joTH3EVknRwnnGViWuH1QucebGtVxd" --nodemode "auto" --datadir "data/shard0-2" --listen "0.0.0.0:9436" --externaladdress "0.0.0.0:9436" --norpcauth --rpclisten "0.0.0.0:9336" --rpcwslisten "0.0.0.0:19336" 
fi
if [ "$1" == "shard0-3" ]; then
INCOGNITO_NETWORK_KEY=local ./incognito --discoverpeersaddress "0.0.0.0:9330" --privatekey "112t8rnqijhT2AqiS8NBVgifb86sqjfwQwf4MHLMAxK3gr1mwxMaeUWQtR1MfxHscrKQ2MsyQMvJ3LEu49LEcZzTzoJCkCiewApeZP48v3no" --nodemode "auto" --datadir "data/shard0-3" --listen "0.0.0.0:9437" --externaladdress "0.0.0.0:9437" --norpcauth --rpclisten "0.0.0.0:9337" --rpcwslisten "0.0.0.0:19337" 
INCOGNITO_NETWORK_KEY=local ./incognito --discoverpeersaddress "0.0.0.0:9330" --privatekey "112t8rnqijhT2AqiS8NBVgifb86sqjfwQwf4MHLMAxK3gr1mwxMaeUWQtR1MfxHscrKQ2MsyQMvJ3LEu49LEcZzTzoJCkCiewApeZP48v3no" --nodemode "auto" --datadir "data/shard0-3" --listen "0.0.0.0:9437" --externaladdress "0.0.0.0:9437" --norpcauth --rpclisten "0.0.0.0:9337" --rpcwslisten "0.0.0.0:19337" 
fi
# Shard 1
if [ "$1" == "shard1-0" ]; then
INCOGNITO_NETWORK_KEY=local ./incognito --discoverpeersaddress "0.0.0.0:9330" --privatekey "112t8rnYRAAQ9BqLA9CF7ESWQzAAUBL1EZQwVPx4z5gPstyNpLk9abFp7iXQFu1rQ5xKukKtvorrxyetpP6Crs7Hj7GeVaVPDaHCkXeHGHSM" --nodemode "auto" --datadir "data/shard1-0" --listen "0.0.0.0:9438" --externaladdress "0.0.0.0:9438" --norpcauth --rpclisten "0.0.0.0:9338" --rpcwslisten "0.0.0.0:19338" 
INCOGNITO_NETWORK_KEY=local ./incognito --discoverpeersaddress "0.0.0.0:9330" --privatekey "112t8rnYRAAQ9BqLA9CF7ESWQzAAUBL1EZQwVPx4z5gPstyNpLk9abFp7iXQFu1rQ5xKukKtvorrxyetpP6Crs7Hj7GeVaVPDaHCkXeHGHSM" --nodemode "auto" --datadir "data/shard1-0" --listen "0.0.0.0:9438" --externaladdress "0.0.0.0:9438" --norpcauth --rpclisten "0.0.0.0:9338" --rpcwslisten "0.0.0.0:19338" 
fi
if [ "$1" == "shard1-1" ]; then
INCOGNITO_NETWORK_KEY=local ./incognito --discoverpeersaddress "0.0.0.0:9330" --privatekey "112t8rnZkBMAJ2DYpYpnmLVJB7YkCWU7NvxxWaETLnKvdMZbhKxVU5iP97GRUBCVZbsknVsGvrdfiajD3d4Av44MXSZQd6DfGiDKdrw9SmtV" --nodemode "auto" --datadir "data/shard1-1" --listen "0.0.0.0:9439" --externaladdress "0.0.0.0:9439" --norpcauth --rpclisten "0.0.0.0:9339" --rpcwslisten "0.0.0.0:19339" 
INCOGNITO_NETWORK_KEY=local ./incognito --discoverpeersaddress "0.0.0.0:9330" --privatekey "112t8rnZkBMAJ2DYpYpnmLVJB7YkCWU7NvxxWaETLnKvdMZbhKxVU5iP97GRUBCVZbsknVsGvrdfiajD3d4Av44MXSZQd6DfGiDKdrw9SmtV" --nodemode "auto" --datadir "data/shard1-1" --listen "0.0.0.0:9439" --externaladdress "0.0.0.0:9439" --norpcauth --rpclisten "0.0.0.0:9339" --rpcwslisten "0.0.0.0:19339" 
fi
if [ "$1" == "shard1-2" ]; then
INCOGNITO_NETWORK_KEY=local ./incognito --discoverpeersaddress "0.0.0.0:9330" --privatekey "112t8rnasPw9nNQqLJ4oposEYxzos63dzDUv33yJTXEaFsNfESFHenv3j32gp9DujciWXouvzPbnP3CFnpysqSPGwrYqfswb4nM1pDLofRAF" --nodemode "auto" --datadir "data/shard1-2" --listen "0.0.0.0:9440" --externaladdress "0.0.0.0:9440" --norpcauth --rpclisten "0.0.0.0:9340" --rpcwslisten "0.0.0.0:19340" 
INCOGNITO_NETWORK_KEY=local ./incognito --discoverpeersaddress "0.0.0.0:9330" --privatekey "112t8rnasPw9nNQqLJ4oposEYxzos63dzDUv33yJTXEaFsNfESFHenv3j32gp9DujciWXouvzPbnP3CFnpysqSPGwrYqfswb4nM1pDLofRAF" --nodemode "auto" --datadir "data/shard1-2" --listen "0.0.0.0:9440" --externaladdress "0.0.0.0:9440" --norpcauth --rpclisten "0.0.0.0:9340" --rpcwslisten "0.0.0.0:19340" 
fi
if [ "$1" == "shard1-3" ]; then
INCOGNITO_NETWORK_KEY=local ./incognito --discoverpeersaddress "0.0.0.0:9330" --privatekey "112t8rncy1vEiCMxvev5EkUQyfH9HLeManjS4kbcsSiMgp4FEiddsiMunhYL2pa8wciCAWxYtt9USgCv21fe2PkSxfnRkiq4AQxTz4KgvLvB" --nodemode "auto" --datadir "data/shard1-3" --listen "0.0.0.0:9441" --externaladdress "0.0.0.0:9441" --norpcauth --rpclisten "0.0.0.0:9341" --rpcwslisten "0.0.0.0:19341" 
INCOGNITO_NETWORK_KEY=local ./incognito --discoverpeersaddress "0.0.0.0:9330" --privatekey "112t8rncy1vEiCMxvev5EkUQyfH9HLeManjS4kbcsSiMgp4FEiddsiMunhYL2pa8wciCAWxYtt9USgCv21fe2PkSxfnRkiq4AQxTz4KgvLvB" --nodemode "auto" --datadir "data/shard1-3" --listen "0.0.0.0:9441" --externaladdress "0.0.0.0:9441" --norpcauth --rpclisten "0.0.0.0:9341" --rpcwslisten "0.0.0.0:19341" 
fi
# Beacon
if [ "$1" == "beacon-0" ]; then
INCOGNITO_NETWORK_KEY=local ./incognito --discoverpeersaddress "0.0.0.0:9330" --privatekey "112t8rncBDbGaFrAE7MZz14d2NPVWprXQuHHXCD2TgSV8USaDFZY3MihVWSqKjwy47sTQ6XvBgNYgdKH2iDVZruKQpRSB5Jqx3A2tef8qVj1" --nodemode "auto" --datadir "data/beacon-0" --listen "0.0.0.0:9450" --externaladdress "0.0.0.0:9450" --norpcauth --rpclisten "0.0.0.0:9350" 
INCOGNITO_NETWORK_KEY=local ./incognito --discoverpeersaddress "0.0.0.0:9330" --privatekey "112t8rncBDbGaFrAE7MZz14d2NPVWprXQuHHXCD2TgSV8USaDFZY3MihVWSqKjwy47sTQ6XvBgNYgdKH2iDVZruKQpRSB5Jqx3A2tef8qVj1" --nodemode "auto" --datadir "data/beacon-0" --listen "0.0.0.0:9450" --externaladdress "0.0.0.0:9450" --norpcauth --rpclisten "0.0.0.0:9350" 
fi
if [ "$1" == "beacon-1" ]; then
INCOGNITO_NETWORK_KEY=local ./incognito --discoverpeersaddress "0.0.0.0:9330" --privatekey "112t8rnY3WLfkE9MsKyW9s3Z5qGnPgCkeutTXJzcT5KJgAMS3vgTL9YbaJ7wyc52CzMnrj8QtwHuCpDzo47PV1qCnrui2dfJzVPU1Wn8q2Jm" --nodemode "auto" --datadir "data/beacon-1" --listen "0.0.0.0:9451" --externaladdress "0.0.0.0:9451" --norpcauth --rpclisten "0.0.0.0:9351" 
INCOGNITO_NETWORK_KEY=local ./incognito --discoverpeersaddress "0.0.0.0:9330" --privatekey "112t8rnY3WLfkE9MsKyW9s3Z5qGnPgCkeutTXJzcT5KJgAMS3vgTL9YbaJ7wyc52CzMnrj8QtwHuCpDzo47PV1qCnrui2dfJzVPU1Wn8q2Jm" --nodemode "auto" --datadir "data/beacon-1" --listen "0.0.0.0:9451" --externaladdress "0.0.0.0:9451" --norpcauth --rpclisten "0.0.0.0:9351" 
fi
if [ "$1" == "beacon-2" ]; then
INCOGNITO_NETWORK_KEY=local ./incognito --discoverpeersaddress "0.0.0.0:9330" --privatekey "112t8rnX5AVkpTZtBo97KdyuDavCtufutWu8tBdDvt6D4WvULd4yyQtiVACadFdDZ28XTGgdfHkmf7wKY9iHo5gsKGwSTnsXZEHW6G7WaPss" --nodemode "auto" --datadir "data/beacon-2" --listen "0.0.0.0:9452" --externaladdress "0.0.0.0:9452" --norpcauth --rpclisten "0.0.0.0:9352" 
INCOGNITO_NETWORK_KEY=local ./incognito --discoverpeersaddress "0.0.0.0:9330" --privatekey "112t8rnX5AVkpTZtBo97KdyuDavCtufutWu8tBdDvt6D4WvULd4yyQtiVACadFdDZ28XTGgdfHkmf7wKY9iHo5gsKGwSTnsXZEHW6G7WaPss" --nodemode "auto" --datadir "data/beacon-2" --listen "0.0.0.0:9452" --externaladdress "0.0.0.0:9452" --norpcauth --rpclisten "0.0.0.0:9352" 
fi
if [ "$1" == "beacon-3" ]; then
INCOGNITO_NETWORK_KEY=local ./incognito --discoverpeersaddress "0.0.0.0:9330" --privatekey "112t8rnaXH1znBqZX1Ry6xvE5hFbQUCuWJb9oiEVfCJDUWbVA9mD4NpL3dLW3TMEUtFajEsu3oKgPLMQyDPEWBuB6JfP4fXEnnAU9hdW1yCb" --nodemode "auto" --datadir "data/beacon-3" --listen "0.0.0.0:9453" --externaladdress "0.0.0.0:9453" --norpcauth --rpclisten "0.0.0.0:9353" 
INCOGNITO_NETWORK_KEY=local ./incognito --discoverpeersaddress "0.0.0.0:9330" --privatekey "112t8rnaXH1znBqZX1Ry6xvE5hFbQUCuWJb9oiEVfCJDUWbVA9mD4NpL3dLW3TMEUtFajEsu3oKgPLMQyDPEWBuB6JfP4fXEnnAU9hdW1yCb" --nodemode "auto" --datadir "data/beacon-3" --listen "0.0.0.0:9453" --externaladdress "0.0.0.0:9453" --norpcauth --rpclisten "0.0.0.0:9353" 
fi
# FullNode
if [ "$1" == "full_node" ]; then
INCOGNITO_NETWORK_KEY=local ./incognito --discoverpeersaddress "0.0.0.0:9330" --nodemode "auto" --datadir "data/full_node" --listen "0.0.0.0:9454" --externaladdress "0.0.0.0:9454" --norpcauth --rpclisten "0.0.0.0:9354" --relayshards "all"  --txpoolmaxtx 100000
fi
######

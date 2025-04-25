#!/usr/bin/env bash
echo "Start Install Dependencies Package"
GO111MODULE=on go get -v

cd ./blockchain/committeestate/ && mockery --name=BeaconCommitteeState --outpkg=externalmocks --output=./externalmocks && cd -
cd ./blockchain/committeestate/ && mockery --name=SplitRewardRuleProcessor --outpkg=externalmocks --output=./externalmocks && cd -
cd ./metadata/ && mockery --name=ChainRetriever && mockery --name=BeaconViewRetriever && mockery --name=ShardViewRetriever && mockery --name=Transaction && cd -
cd ./consensus_v2/blsbft/ && mockery --name=NodeInterface  && mockery --name=CommitteeChainHandler && mockery  --name=Chain && cd -
cd ./blockchain/types/ && mockery --name=BlockInterface && cd -
cd ./multiview/ && mockery --name=View && cd -
echo "Start Unit-Test"
echo "package committeestate"
GO111MODULE=on go test -cover ./blockchain/committeestate/*.go
echo "package finishsync"
GO111MODULE=on go test -cover ./syncker/finishsync/*.go
echo "package statedb"
GO111MODULE=on go test -cover ./dataaccessobject/statedb/*.go
echo "package instruction"
GO111MODULE=on go test -cover ./instruction/*.go
echo "package blockchain"
GO111MODULE=on go test -cover ./blockchain/*.go
echo "package metadata"
GO111MODULE=on go test -cover ./metadata/*.go
echo "package signaturecounter"
GO111MODULE=on go test -cover ./blockchain/signaturecounter/*.go
echo "package blsbft"
GO111MODULE=on go test -cover ./consensus_v2/blsbft/*.go


echo "Start build Incognito"
APP_NAME="incognito"
echo "go build -o $APP_NAME"
GO111MODULE=on go build -o $APP_NAME

echo "Build Incognito success!"
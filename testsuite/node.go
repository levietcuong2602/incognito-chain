package devframework

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"
	"time"

	"github.com/levietcuong2602/incognito-chain/config"
	"github.com/levietcuong2602/incognito-chain/consensus_v2/blsbft"
	"github.com/levietcuong2602/incognito-chain/consensus_v2/consensustypes"
	"github.com/levietcuong2602/incognito-chain/portal"
	zkp "github.com/levietcuong2602/incognito-chain/privacy/privacy_v1/zeroknowledge"
	"github.com/levietcuong2602/incognito-chain/wallet"
	"github.com/levietcuong2602/incognito-chain/wire"

	"github.com/levietcuong2602/incognito-chain/blockchain/types"
	"github.com/levietcuong2602/incognito-chain/txpool"

	"github.com/levietcuong2602/incognito-chain/consensus_v2"
	"github.com/levietcuong2602/incognito-chain/consensus_v2/signatureschemes"
	"github.com/levietcuong2602/incognito-chain/incognitokey"

	"github.com/levietcuong2602/incognito-chain/testsuite/account"
	"github.com/levietcuong2602/incognito-chain/testsuite/mock"
	"github.com/levietcuong2602/incognito-chain/testsuite/rpcclient"
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/filter"
	"github.com/syndtr/goleveldb/leveldb/opt"

	"github.com/levietcuong2602/incognito-chain/pubsub"

	"github.com/levietcuong2602/incognito-chain/syncker"

	"github.com/levietcuong2602/incognito-chain/blockchain"
	"github.com/levietcuong2602/incognito-chain/common"

	"github.com/levietcuong2602/incognito-chain/incdb"
	_ "github.com/levietcuong2602/incognito-chain/incdb/lvdb"
	"github.com/levietcuong2602/incognito-chain/memcache"
	"github.com/levietcuong2602/incognito-chain/mempool"
	"github.com/levietcuong2602/incognito-chain/metadata"
	"github.com/levietcuong2602/incognito-chain/rpcserver"

	lvdbErrors "github.com/syndtr/goleveldb/leveldb/errors"

	"github.com/pkg/errors"
)

type Config struct {
	Network int
	DataDir string
	ResetDB bool
	AppNode bool
}

type NodeEngine struct {
	config      Config
	param       blockchain.Config
	appNodeMode string
	simName     string
	timer       *TimeEngine

	//for account manager
	accountSeed       string
	accountGenHistory map[int]int
	committeeAccount  map[int][]account.Account
	accounts          []*account.Account

	GenesisAccount account.Account

	//blockchain dependency object
	bc          *blockchain.BlockChain
	ps          *pubsub.PubSubManager
	consensus   mock.ConsensusInterface
	txpool      *mempool.TxPool
	temppool    *mempool.TxPool
	btcrd       *mock.BTCRandom
	syncker     *syncker.SynckerManager
	server      *mock.Server
	cPendingTxs chan metadata.Transaction
	cRemovedTxs chan metadata.Transaction
	rpcServer   *rpcserver.RpcServer
	cQuit       chan struct{}

	RPC               *rpcclient.RPCClient
	listennerRegister map[int][]func(msg interface{})

	userDB        *leveldb.DB
	lightNodeData struct {
		Shards                map[byte]*currentShardState
		ProcessedBeaconHeight uint64
	}
}

type currentShardState struct {
	// BestHeight  uint64
	// BestHash    *common.Hash
	LocalHeight uint64
	LocalHash   *common.Hash
}

func (sim *NodeEngine) NewAccountFromShard(sid int) account.Account {
	lastID := sim.accountGenHistory[sid]
	lastID++
	sim.accountGenHistory[sid] = lastID
	acc, _ := account.GenerateAccountByShard(sid, lastID, sim.accountSeed)
	acc.SetName(fmt.Sprintf("ACC_%v", len(sim.accounts)+1))
	sim.accounts = append(sim.accounts, &acc)
	return acc
}

func (sim *NodeEngine) GetAllAccounts() []*account.Account {
	return sim.accounts
}

func (sim *NodeEngine) NewAccountFromPrivateKey(prv string) account.Account {
	acc, _ := account.NewAccountFromPrivatekey(prv)
	sim.accounts = append(sim.accounts, &acc)
	return acc
}

func (sim *NodeEngine) NewAccount() account.Account {
	lastID := sim.accountGenHistory[0]
	lastID++
	sim.accountGenHistory[0] = lastID
	acc, _ := account.GenerateAccountByShard(0, lastID, sim.accountSeed)
	return acc
}

func (sim *NodeEngine) EnableDebug() {
	dbLogger.SetLevel(common.LevelTrace)
	blockchainLogger.SetLevel(common.LevelInfo)
	bridgeLogger.SetLevel(common.LevelTrace)
	rpcLogger.SetLevel(common.LevelTrace)
	rpcServiceLogger.SetLevel(common.LevelTrace)
	rpcServiceBridgeLogger.SetLevel(common.LevelTrace)
	transactionLogger.SetLevel(common.LevelTrace)
	privacyLogger.SetLevel(common.LevelTrace)
	mempoolLogger.SetLevel(common.LevelTrace)
}

func (sim *NodeEngine) Init() {
	os.Setenv("TXPOOL_VERSION", "0")
	simName := sim.simName
	common.MaxShardNumber = config.Param().ActiveShards
	common.TIMESLOT = config.Param().ConsensusParam.Timeslot
	InitLogRotator(filepath.Join(sim.config.DataDir, simName+".log"))
	activeNetParams := sim.param
	wallet.InitPublicKeyBurningAddressByte()
	config.Param().GenesisParam.PreSelectBeaconNodeSerializedPubkey = []string{}
	config.Param().GenesisParam.PreSelectBeaconNodeSerializedPaymentAddress = []string{}
	config.Param().GenesisParam.PreSelectShardNodeSerializedPubkey = []string{}
	config.Param().GenesisParam.PreSelectShardNodeSerializedPaymentAddress = []string{}

	config.Param().GenesisParam.SelectBeaconNodeSerializedPubkeyV2 = map[uint64][]string{}
	config.Param().GenesisParam.SelectBeaconNodeSerializedPaymentAddressV2 = map[uint64][]string{}
	config.Param().GenesisParam.SelectShardNodeSerializedPubkeyV2 = map[uint64][]string{}
	config.Param().GenesisParam.SelectShardNodeSerializedPaymentAddressV2 = map[uint64][]string{}

	sim.GenesisAccount = sim.NewAccount()
	for i := 0; i < config.Param().CommitteeSize.MinBeaconCommitteeSize; i++ {
		acc := sim.NewAccountFromShard(-1)
		sim.committeeAccount[-1] = append(sim.committeeAccount[-1], acc)
		config.Param().GenesisParam.PreSelectBeaconNodeSerializedPubkey = append(config.Param().GenesisParam.PreSelectBeaconNodeSerializedPubkey, acc.SelfCommitteePubkey)
		config.Param().GenesisParam.PreSelectBeaconNodeSerializedPaymentAddress = append(config.Param().GenesisParam.PreSelectBeaconNodeSerializedPaymentAddress, acc.PaymentAddress)
	}
	for i := 0; i < config.Param().ActiveShards; i++ {
		for a := 0; a < config.Param().CommitteeSize.MinShardCommitteeSize; a++ {
			acc := sim.NewAccountFromShard(i)
			sim.committeeAccount[i] = append(sim.committeeAccount[i], acc)
			config.Param().GenesisParam.PreSelectShardNodeSerializedPubkey = append(config.Param().GenesisParam.PreSelectShardNodeSerializedPubkey, acc.SelfCommitteePubkey)
			config.Param().GenesisParam.PreSelectShardNodeSerializedPaymentAddress = append(config.Param().GenesisParam.PreSelectShardNodeSerializedPaymentAddress, acc.PaymentAddress)
		}
	}
	initTxs := createGenesisTx([]account.Account{sim.GenesisAccount})
	config.Param().GenesisParam.InitialIncognito = initTxs

	zkp.InitCheckpoint(config.Param().BCHeightBreakPointNewZKP)

	blockchain.CreateGenesisBlocks()

	//init time
	layout := "2006-01-02T15:04:05.000Z"
	str := config.Param().GenesisParam.BlockTimestamp
	genesisTime, err := time.Parse(layout, str)
	sim.timer.init(int64(genesisTime.Unix() + 10))

	//init blockchain
	bc := blockchain.BlockChain{}

	cs := mock.Consensus{}
	txpoolV1 := mempool.TxPool{}
	temppool := mempool.TxPool{}
	btcrd := mock.BTCRandom{} // use mock for now
	sync := syncker.NewSynckerManager()
	server := mock.Server{
		BlockChain: &bc,
		TxPool:     &txpoolV1,
	}
	ps := pubsub.NewPubSubManager()
	fees := make(map[byte]*mempool.FeeEstimator)
	relayShards := []byte{}
	for i := byte(0); i < byte(config.Param().ActiveShards); i++ {
		relayShards = append(relayShards, i)
		fees[i] = mempool.NewFeeEstimator(
			mempool.DefaultEstimateFeeMaxRollback,
			mempool.DefaultEstimateFeeMinRegisteredBlocks,
			config.Config().LimitFee, config.Config().MinFeePerTx, 0, 0, 0)
	}
	cPendingTxs := make(chan metadata.Transaction, 500)
	cRemovedTxs := make(chan metadata.Transaction, 500)
	cQuit := make(chan struct{})
	blockgen, err := blockchain.NewBlockGenerator(&txpoolV1, &bc, sync, cPendingTxs, cRemovedTxs)
	if err != nil {
		panic(err)
	}
	dbpath := filepath.Join(sim.config.DataDir)
	db, err := incdb.OpenMultipleDB("leveldb", dbpath)
	// Create db and use it.
	if err != nil {
		panic(err)
	}

	//listenFunc := net.Listen
	//listener, err := listenFunc("tcp", "0.0.0.0:8000")
	//if err != nil {
	//	panic(err)
	//}

	rpcConfig := rpcserver.RpcServerConfig{
		HttpListenters:  []net.Listener{nil},
		RPCMaxClients:   1,
		DisableAuth:     true,
		BlockChain:      &bc,
		Blockgen:        blockgen,
		TxMemPool:       &txpoolV1,
		Server:          &server,
		Database:        db,
		ConsensusEngine: &cs,
	}
	rpcServer := &rpcserver.RpcServer{}
	rpclocal := &LocalRPCClient{rpcServer}

	btcChain, err := getBTCRelayingChain(portal.GetPortalParams().RelayingParam.BTCRelayingHeaderChainID, "btcchain", sim.config.DataDir)
	if err != nil {
		panic(err)
	}
	bnbChainState, err := getBNBRelayingChainState(portal.GetPortalParams().RelayingParam.BNBRelayingHeaderChainID, sim.config.DataDir)
	if err != nil {
		panic(err)
	}

	txpoolV1.Init(&mempool.Config{
		ConsensusEngine: &cs,
		BlockChain:      &bc,
		DataBase:        db,
		FeeEstimator:    fees,
		TxLifeTime:      100,
		MaxTx:           1000,
		// DataBaseMempool:   dbmp,
		IsLoadFromMempool: false,
		PersistMempool:    false,
		RelayShards:       nil,
		PubSubManager:     ps,
	})

	// serverObj.blockChain.AddTxPool(serverObj.memPool)
	txpoolV1.InitChannelMempool(cPendingTxs, cRemovedTxs)

	temppool.Init(&mempool.Config{
		BlockChain:    &bc,
		DataBase:      db,
		FeeEstimator:  fees,
		MaxTx:         1000,
		PubSubManager: ps,
	})
	txpoolV1.IsBlockGenStarted = true
	go temppool.Start(cQuit)
	go txpoolV1.Start(cQuit)
	poolManager, _ := txpool.NewPoolManager(
		common.MaxShardNumber,
		ps,
		time.Duration(15*60)*time.Second,
	)
	otadb, err := incdb.Open("leveldb", "/tmp/database/ota")
	if err != nil {
		panic(err)
	}
	err = bc.Init(&blockchain.Config{
		BTCChain:          btcChain,
		BNBChainState:     bnbChainState,
		DataBase:          db,
		OutCoinByOTAKeyDb: &otadb,
		MemCache:          memcache.New(),
		BlockGen:          blockgen,
		TxPool:            &txpoolV1,
		TempTxPool:        &temppool,
		Server:            &server,
		Syncker:           sync,
		PubSubManager:     ps,
		FeeEstimator:      make(map[byte]blockchain.FeeEstimator),
		ConsensusEngine:   &cs,
		PoolManager:       poolManager,
	})

	if err != nil {
		panic(err)
	}
	bc.InitChannelBlockchain(cRemovedTxs)
	go poolManager.Start(relayShards)
	for shardID, feeEstimator := range fees {
		bc.SetFeeEstimator(feeEstimator, shardID)
	}

	sim.param = activeNetParams
	sim.bc = &bc
	//sim.consensus = &cs
	sim.txpool = &txpoolV1
	sim.temppool = &temppool
	sim.btcrd = &btcrd
	sim.syncker = sync
	sim.server = &server
	sim.cPendingTxs = cPendingTxs
	sim.cRemovedTxs = cRemovedTxs
	sim.rpcServer = rpcServer
	sim.RPC = rpcclient.NewRPCClient(rpclocal)
	sim.cQuit = cQuit
	sim.listennerRegister = make(map[int][]func(msg interface{}))
	sim.ps = ps
	rpcServer.Init(&rpcConfig)
	go func() {
		for {
			select {
			case <-cQuit:
				return
			case <-cRemovedTxs:
			}
		}
	}()
	go blockgen.Start(cQuit)

	sim.startPubSub()

	//init syncker
	sim.syncker.Init(&syncker.SynckerManagerConfig{Blockchain: sim.bc, Consensus: sim.consensus})

	//init user database
	handles := 256
	cache := 8
	userDBPath := filepath.Join(dbpath, "userdb")
	lvdb, err := leveldb.OpenFile(userDBPath, &opt.Options{
		OpenFilesCacheCapacity: handles,
		BlockCacheCapacity:     cache / 2 * opt.MiB,
		WriteBuffer:            cache / 4 * opt.MiB, // Two of these are used internally
		Filter:                 filter.NewBloomFilter(10),
	})
	if _, corrupted := err.(*lvdbErrors.ErrCorrupted); corrupted {
		lvdb, err = leveldb.RecoverFile(userDBPath, nil)
	}
	sim.userDB = lvdb
	if err != nil {
		panic(errors.Wrapf(err, "levelvdb.OpenFile %s", userDBPath))
	}
}

func (sim *NodeEngine) startPubSub() {
	go sim.ps.Start()
	go func() {
		_, subChan, err := sim.ps.RegisterNewSubscriber(pubsub.BeaconBeststateTopic)
		if err != nil {
			panic("something wrong with subscriber")
		}
		for {
			event := <-subChan
			for _, f := range sim.listennerRegister[BLK_BEACON] {
				f(event.Value)
			}
		}
	}()

	go func() {
		_, subChan, err := sim.ps.RegisterNewSubscriber(pubsub.ShardBeststateTopic)
		if err != nil {
			panic("something wrong with subscriber")
		}
		for {
			event := <-subChan
			for _, f := range sim.listennerRegister[BLK_SHARD] {
				f(event.Value)
			}
		}
	}()
}

func (sim *NodeEngine) EmptyPool() {
	sim.temppool.EmptyPool()
	sim.txpool.EmptyPool()
}
func (sim *NodeEngine) StopSync() {
	sim.syncker.Stop()
}

func (sim *NodeEngine) Pause() {
	fmt.Print("Simulation pause! Press Enter to continue ...")
	var input string
	fmt.Scanln(&input)
	fmt.Print("\n")
}

func (sim *NodeEngine) PrintBlockChainInfo() {
	fmt.Println("Beacon Chain:")

	fmt.Println("Shard Chain:")
}

// life cycle of a block generation process:
// PreCreate -> PreValidation -> PreInsert ->
func (sim *NodeEngine) GenerateBlock(args ...interface{}) *NodeEngine {
	time.Sleep(time.Nanosecond)
	var chainArray = []int{-1}
	var validatorIndex ValidatorIndex = nil
	for i := 0; i < config.Param().ActiveShards; i++ {
		chainArray = append(chainArray, i)
	}
	//beacon
	chain := sim.bc

	var err error

	for _, arg := range args {
		switch arg.(type) {
		case *Execute:
			exec := arg.(*Execute)
			chainArray = exec.appliedChain
		case ValidatorIndex:
			validatorIndex = arg.(ValidatorIndex)
		}
	}
	//Create blocks for apply chain
	for _, chainID := range chainArray {
		var block types.BlockInterface = nil
		curView := sim.bc.GetChain(chainID).(Chain).GetBestView()
		for _, arg := range args {
			switch arg.(type) {
			case PreView:
				curView = arg.(PreView).View
			}
		}

		var proposerPK incognitokey.CommitteePublicKey
		committeeFromBlock := common.Hash{}
		committees := curView.GetCommittee()
		version := 9
		versionNew := sim.GetBlockVersion(chainID)
		fmt.Printf("Create new block of chain %v, version %v, version by feature %v\n", chainID, version, versionNew)
		if version < versionNew {
			version = versionNew
		}
		proposerPK, _ = curView.GetProposerByTimeSlot(int64((uint64(sim.timer.Now()) / common.TIMESLOT)), 2)
		committeeFromBlock = *chain.BeaconChain.FinalView().GetHash()
		if chainID > -1 {
			committees, _ = sim.bc.GetShardCommitteeFromBeaconHash(committeeFromBlock, byte(chainID))
			fmt.Println("version 3 from beacon", chain.BeaconChain.FinalView().GetHeight(), committeeFromBlock)
		}
		proposerPkStr, _ := proposerPK.ToBase58()
		// fmt.Println("Create new block", chain.BeaconChain.GetBestView().GetHeight()+1)
		if chainID == -1 {
			block, err = chain.BeaconChain.CreateNewBlock(version, proposerPkStr, 1, sim.timer.Now(), committees, common.Hash{})
			if err != nil {
				fmt.Printf("Can not create block for chain %v %+v\n", chainID, err)
				Logger.log.Error(err)
				return sim
			}

		} else {
			block, err = chain.ShardChain[byte(chainID)].CreateNewBlock(version, proposerPkStr, 1, sim.timer.Now(), committees, committeeFromBlock)
			if err != nil {
				fmt.Printf("Can not create block for chain %v %+v\n", chainID, err)
				Logger.log.Error(err)
				return sim
			}
		}
		// fmt.Printf("Proposer %+v %v \n", proposerPK, proposerPkStr)
		//SignBlock
		proposeAcc := sim.GetAccountByCommitteePubkey(&proposerPK)
		// fmt.Printf("Proposer acc %+v %+v \n", proposeAcc, *proposeAcc)
		userKey, _ := consensus_v2.GetMiningKeyFromPrivateSeed(proposeAcc.MiningKey)
		sim.SignBlock(userKey, block)

		//simulate network transfer
		b, err := json.Marshal(block)
		if err != nil {
			panic(err)
		}
		err = json.Unmarshal(b, block)
		if err != nil {
			panic(err)
		}

		//Validation
		if chainID == -1 {
			err = chain.BeaconChain.ValidatePreSignBlock(block.(*types.BeaconBlock), nil, committees)
			if err != nil {
				Logger.log.Error(err)
				return sim
			}
		} else {
			err = chain.ShardChain[byte(chainID)].ValidatePreSignBlock(block.(*types.ShardBlock), nil, committees)
			if err != nil {
				panic(err)
			}
		}

		//Combine votes
		accs, err := sim.GetListAccountByCommitteePubkey(committees)
		if err != nil {
			panic(err)
		}
		if validatorIndex == nil {
			err = sim.SignBlockWithCommittee(block, accs, GenerateCommitteeIndex(len(committees)))
			if err != nil {
				panic(err)
			}
		} else {
			err = sim.SignBlockWithCommittee(block, accs, validatorIndex)
			if err != nil {
				panic(err)
			}

		}

		//Insert
		if chainID == -1 {
			err = chain.BeaconChain.InsertBlock(block.(*types.BeaconBlock), true)
			if err != nil {
				panic(err)
			}
			log.Printf("BEACON | Produced block %v hash %v", block.GetHeight(), block.Hash().String())
		} else {
			err = chain.ShardChain[byte(chainID)].InsertBlock(block.(*types.ShardBlock), true)
			if err != nil {
				panic(err)
			} else {
				crossX := types.CreateAllCrossShardBlock(block.(*types.ShardBlock), config.Param().ActiveShards)
				log.Printf("SHARD %v | Produced block %v hash %v", chainID, block.GetHeight(), block.Hash().String())
				for _, blk := range crossX {
					sim.syncker.InsertCrossShardBlock(blk)
				}
			}
		}
	}

	return sim
}

// number of second we want simulation to forward
// default = round interval
func (sim *NodeEngine) NextRound() {
	sim.timer.Forward(int64(common.TIMESLOT))
}

//func (sim *NodeEngine) InjectTx(txBase58 string) error {
//	rawTxBytes, _, err := base58.Base58Check{}.Decode(txBase58)
//	if err != nil {
//		return err
//	}
//	var tx transaction.Tx
//	err = json.Unmarshal(rawTxBytes, &tx)
//	if err != nil {
//		return err
//	}
//	sim.cPendingTxs <- &tx
//
//	return nil
//}

func (sim *NodeEngine) GetBlockchain() *blockchain.BlockChain {
	return sim.bc
}

func (sim *NodeEngine) GetSyncker() *syncker.SynckerManager {
	return sim.syncker
}

func (s *NodeEngine) GetUserDatabase() *leveldb.DB {
	return s.userDB
}

func (s *NodeEngine) SignBlockWithCommittee(block types.BlockInterface, committees []account.Account, committeeIndex []int) error {
	committeePubKey := []incognitokey.CommitteePublicKey{}
	miningKeys := []*signatureschemes.MiningKey{}
	//if len(committees) != len(committeeIndex) {
	//	fmt.Println(len(committees), len(committeeIndex), committeeIndex)
	//}
	// fmt.Printf("%+v %+v %+v aaaaaaaaaaaaaaaaaaaaaaa\n", committeeIndex, block.GetHeight(), block.GetShardID())
	if block.GetVersion() >= 2 {
		votes := make(map[string]*blsbft.BFTVote)
		for _, committee := range committees {
			miningKey, _ := consensus_v2.GetMiningKeyFromPrivateSeed(committee.MiningKey)
			committeePubKey = append(committeePubKey, *miningKey.GetPublicKey())
			miningKeys = append(miningKeys, miningKey)
			//if len(committees) != len(committeeIndex) {
			//	fmt.Println(committee.Name)
			//}
		}
		var c blsbft.Chain
		if block.GetShardID() == -1 {
			c = s.bc.BeaconChain
		} else {
			c = s.bc.ShardChain[block.GetShardID()]
		}
		for _, committeeID := range committeeIndex {
			vote, _ := blsbft.CreateVote(c, miningKeys[committeeID], block, committeePubKey, s.bc.GetChain(-1).(*blockchain.BeaconChain).GetPortalParamsV4(0))
			vote.IsValid = 1
			votes[vote.Validator] = vote
		}
		committeeBLSString, _ := incognitokey.ExtractPublickeysFromCommitteeKeyList(committeePubKey, common.BlsConsensus)
		aggSig, brigSigs, validatorIdx, portalSigs, err := blsbft.CombineVotes(votes, committeeBLSString)

		valData, err := consensustypes.DecodeValidationData(block.GetValidationField())
		if err != nil {
			return errors.New("decode validation data")
		}
		valData.AggSig = aggSig
		valData.BridgeSig = brigSigs
		valData.ValidatiorsIdx = validatorIdx
		valData.PortalSig = portalSigs
		validationDataString, _ := consensustypes.EncodeValidationData(*valData)
		block.(blsbft.BlockValidation).AddValidationField(validationDataString)
	}
	return nil
}

func (s *NodeEngine) SignBlock(userMiningKey *signatureschemes.MiningKey, block types.BlockInterface) {
	var validationData consensustypes.ValidationData
	validationData.ProducerBLSSig, _ = userMiningKey.BriSignData(block.ProposeHash().GetBytes())
	validationDataString, _ := consensustypes.EncodeValidationData(validationData)
	block.(blsbft.BlockValidation).AddValidationField(validationDataString)
}

func (s *NodeEngine) GetAccountByCommitteePubkey(cpk *incognitokey.CommitteePublicKey) *account.Account {
	miningPK := cpk.GetMiningKeyBase58(common.BlsConsensus)
	for _, acc := range s.accounts {
		if acc.MiningPubkey == miningPK {
			return acc
		}
	}
	return nil
}

func (s *NodeEngine) GetListAccountByCommitteePubkey(cpks []incognitokey.CommitteePublicKey) ([]account.Account, error) {
	accounts := []account.Account{}
	for _, cpk := range cpks {
		if acc := s.GetAccountByCommitteePubkey(&cpk); acc != nil {
			accounts = append(accounts, *acc)
		}
	}
	if len(accounts) != len(cpks) {
		return nil, errors.New("Mismatch number of committee pubkey in beststate")
	}
	return accounts, nil
}

func (sim *NodeEngine) GetMultiview(chainID int) ([]account.Account, error) {
	committees := sim.bc.GetChain(chainID).(Chain).GetBestView().GetCommittee()
	return sim.GetListAccountByCommitteePubkey(committees)
}

type SimulationParam struct {
}

func InitChainParam(cfg Config, customParam func(), postInit func(*NodeEngine)) *NodeEngine {
	node := NewStandaloneSimulation("sim", cfg)
	customParam()
	node.Init()
	customParam()
	postInit(node)
	for i := 0; i < 3; i++ {
		node.GenerateBlock().NextRound()
	}
	go func() {
		for {
			node.SendFeatureStat(GetAllFixedNodesAccount(node), []string{})
			time.Sleep(2 * time.Second)
		}
	}()
	for i := 0; i < 3; i++ {
		node.GenerateBlock().NextRound()
	}
	node.RPC.API_SubmitKey(node.GenesisAccount.PrivateKey)
	err := node.RPC.API_CreateConvertCoinVer1ToVer2Transaction(node.GenesisAccount.PrivateKey)
	if err != nil {
		panic(err)
	}

	for i := 0; i < 5; i++ {
		node.GenerateBlock().NextRound()
	}
	return node
}

func (s *NodeEngine) SendFinishSync(accs []account.Account, sid byte) {
	finishedSyncValidators := []string{}
	finishedSyncSignatures := [][]byte{}
	for _, v := range accs {
		signature, err := v.MiningKeySet.BriSignData([]byte(wire.CmdMsgFinishSync))
		if err != nil {
			continue
		}
		cpk := v.SelfCommitteePubkey
		finishedSyncSignatures = append(finishedSyncSignatures, signature)
		finishedSyncValidators = append(finishedSyncValidators, cpk)
	}
	if len(finishedSyncValidators) == 0 {
		return
	}
	msg := wire.NewMessageFinishSync(finishedSyncValidators, finishedSyncSignatures, sid)
	s.GetBlockchain().AddFinishedSyncValidators(msg.CommitteePublicKey, msg.Signature, msg.ShardID)
}

func (s *NodeEngine) GetUntriggerFeature(afterCheckPoint bool) []string {
	curView := s.bc.GetBeaconBestState()
	unTriggerFeatures := []string{}
	for f, _ := range config.Param().AutoEnableFeature {
		if config.Param().AutoEnableFeature[f].MinTriggerBlockHeight == 0 {
			//skip default value
			continue
		}
		if curView.TriggeredFeature == nil || curView.TriggeredFeature[f] == 0 {
			if afterCheckPoint {
				if curView.BeaconHeight > uint64(config.Param().AutoEnableFeature[f].MinTriggerBlockHeight) {
					unTriggerFeatures = append(unTriggerFeatures, f)
				}
			} else {
				unTriggerFeatures = append(unTriggerFeatures, f)
			}

		}
	}
	return unTriggerFeatures
}

func (s *NodeEngine) SendFeatureStat(accs []account.Account, unTriggerFeatures []string) {
	unTriggerFeatures = append(unTriggerFeatures, s.GetUntriggerFeature(false)...)
	// unTriggerFeatures = append(unTriggerFeatures, s.GetUntriggerFeature(true)...)
	featureSyncValidators := []string{}
	featureSyncSignatures := [][]byte{}
	// fmt.Printf("%v\n", len(accs))

	signBytes := []byte{}
	for _, v := range unTriggerFeatures {
		signBytes = append([]byte(wire.CmdMsgFeatureStat), []byte(v)...)
	}
	timestamp := time.Now().Unix()
	timestampStr := fmt.Sprintf("%v", timestamp)
	signBytes = append(signBytes, []byte(timestampStr)...)
	for _, v := range accs {
		dataSign := signBytes[:]
		dataSign = append(dataSign, []byte(v.SelfCommitteePubkey)...)
		signature, err := v.MiningKeySet.BriSignData(dataSign)
		if err != err {
			continue
		}
		featureSyncSignatures = append(featureSyncSignatures, signature)
		featureSyncValidators = append(featureSyncValidators, v.SelfCommitteePubkey)
	}
	if len(featureSyncValidators) == 0 {
		return
	}
	// fmt.Println("Send ", featureSyncValidators, unTriggerFeatures)
	s.GetBlockchain().ReceiveFeatureReport(int(timestamp), featureSyncValidators, featureSyncSignatures, unTriggerFeatures)
}

func (s *NodeEngine) GetBlockVersion(chainID int) int {
	chainEpoch := uint64(1)
	chainHeight := uint64(1)
	triggerFeature := make(map[string]uint64)
	if chainID == -1 {
		chainEpoch = s.bc.BeaconChain.GetEpoch()
		chainHeight = s.bc.BeaconChain.GetBestViewHeight()
		triggerFeature = s.bc.BeaconChain.GetFinalView().(*blockchain.BeaconBestState).TriggeredFeature
	} else {
		chainEpoch = s.bc.ShardChain[chainID].GetEpoch()
		chainHeight = s.bc.ShardChain[chainID].GetBestView().GetBeaconHeight()
		triggerFeature = s.bc.ShardChain[chainID].GetFinalView().(*blockchain.ShardBestState).TriggeredFeature
	}

	//get last trigger feature that change block version
	latestFeature := ""
	latestTriggerHeight := uint64(0)
	for f, h := range triggerFeature {
		if _, ok := config.Param().FeatureVersion[f]; ok {
			if latestTriggerHeight < h {
				latestTriggerHeight = h
				latestFeature = f
			}
		}
	}
	if version, ok := config.Param().FeatureVersion[latestFeature]; ok {
		return int(version)
	}

	//legacy flow
	if triggerFeature[blockchain.INSTANT_FINALITY_FEATURE] != 0 {
		return types.INSTANT_FINALITY_VERSION
	}

	if chainHeight >= config.Param().ConsensusParam.BlockProducingV3Height {
		return types.BLOCK_PRODUCINGV3_VERSION
	}

	if chainHeight >= config.Param().ConsensusParam.Lemma2Height {
		return types.LEMMA2_VERSION
	}

	if chainHeight >= config.Param().ConsensusParam.StakingFlowV3Height {
		return types.SHARD_SFV3_VERSION
	}

	if chainHeight >= config.Param().ConsensusParam.StakingFlowV2Height {
		return types.SHARD_SFV2_VERSION
	}

	if chainEpoch >= config.Param().ConsensusParam.ConsensusV2Epoch {
		return types.MULTI_VIEW_VERSION
	}

	return types.BFT_VERSION
}

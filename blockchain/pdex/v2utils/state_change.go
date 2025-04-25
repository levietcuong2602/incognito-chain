package v2utils

type StateChange struct {
	PoolPairs    map[string]*PoolPairChange
	StakingPools map[string]*StakingPoolChange
}

func NewStateChange() *StateChange {
	return &StateChange{
		PoolPairs:    make(map[string]*PoolPairChange),
		StakingPools: make(map[string]*StakingPoolChange),
	}
}

type StakingPoolChange struct {
	Stakers         map[string]*StakerChange
	RewardsPerShare map[string]bool
}

type StakerChange struct {
	IsChanged           bool
	Rewards             map[string]bool
	LastRewardsPerShare map[string]bool
}

func NewStakingChange() *StakingPoolChange {
	return &StakingPoolChange{
		Stakers:         make(map[string]*StakerChange),
		RewardsPerShare: make(map[string]bool),
	}
}

func NewStakerChange() *StakerChange {
	return &StakerChange{
		Rewards:             make(map[string]bool),
		LastRewardsPerShare: make(map[string]bool),
	}
}

type PoolPairChange struct {
	IsChanged       bool
	Shares          map[string]*ShareChange
	OrderIDs        map[string]bool
	LpFeesPerShare  map[string]bool
	ProtocolFees    map[string]bool
	StakingPoolFees map[string]bool
	MakingVolume    map[string]*MakingVolumeChange
	OrderRewards    map[string]*OrderRewardChange
}

func NewPoolPairChange() *PoolPairChange {
	return &PoolPairChange{
		Shares:          make(map[string]*ShareChange),
		OrderIDs:        make(map[string]bool),
		LpFeesPerShare:  make(map[string]bool),
		ProtocolFees:    make(map[string]bool),
		StakingPoolFees: make(map[string]bool),
		MakingVolume:    make(map[string]*MakingVolumeChange),
		OrderRewards:    make(map[string]*OrderRewardChange),
	}
}

type ShareChange struct {
	IsChanged          bool
	TradingFees        map[string]bool
	LastLPFeesPerShare map[string]bool
}

func NewShareChange() *ShareChange {
	return &ShareChange{
		TradingFees:        make(map[string]bool),
		LastLPFeesPerShare: make(map[string]bool),
	}
}

type MakingVolumeChange struct {
	Volume map[string]bool
}

func NewMakingVolumeChange() *MakingVolumeChange {
	return &MakingVolumeChange{
		Volume: make(map[string]bool),
	}
}

type OrderRewardChange struct {
	UncollectedReward map[string]bool
}

func NewOrderRewardChange() *OrderRewardChange {
	return &OrderRewardChange{
		UncollectedReward: make(map[string]bool),
	}
}

package client

const (
	// ReadersPolicyKey is the key used for the read policy
	ReadersPolicyKey = "Readers"

	// WritersPolicyKey is the key used for the read policy
	WritersPolicyKey = "Writers"

	// AdminsPolicyKey is the key used for the read policy
	AdminsPolicyKey = "Admins"

	// BlockValidationPolicyKey TODO
	BlockValidationPolicyKey = "BlockValidation"

	EndorsementPolicyKey = "Endorsement"

	EndpointsKey = "Endpoints"

	// AnchorPeersKey is the key name for the AnchorPeers ConfigValue
	AnchorPeersKey = "AnchorPeers"

	MSPKey = "MSP"

	// SignaturePolicyType is the 'Type' string for signature policies
	SignaturePolicyType = "Signature"

	// ImplicitMetaPolicyType is the 'Type' string for implicit meta policies
	ImplicitMetaPolicyType = "ImplicitMeta"

	// HashingAlgorithmKey is the cb.ConfigItem type key name for the HashingAlgorithm message
	HashingAlgorithmKey = "HashingAlgorithm"

	// BlockDataHashingStructureKey is the cb.ConfigItem type key name for the BlockDataHashingStructure message
	BlockDataHashingStructureKey = "BlockDataHashingStructure"

	// CapabilitiesKey is the name of the key which refers to capabilities, it appears at the channel,
	// application, and orderer levels and this constant is used for all three.
	CapabilitiesKey = "Capabilities"

	// BatchSizeKey is the cb.ConfigItem type key name for the BatchSize message.
	BatchSizeKey = "BatchSize"

	// BatchTimeoutKey is the cb.ConfigItem type key name for the BatchTimeout message.
	BatchTimeoutKey = "BatchTimeout"

	// ChannelRestrictionsKey is the key name for the ChannelRestrictions message.
	ChannelRestrictionsKey = "ChannelRestrictions"

	// ConsensusTypeKey is the cb.ConfigItem type key name for the ConsensusType message.
	ConsensusTypeKey = "ConsensusType"

	ConsensusTypeEtcdRaft = "etcdraft"

	// OrdererGroupKey is the group name for the orderer config.
	OrdererGroupKey = "Orderer"

	ordererAdminsPolicyName = "/Channel/Orderer/Admins"

	// ChannelCreationPolicyKey is the key used in the consortium config to denote the policy
	// to be used in evaluating whether a channel creation request is authorized
	ChannelCreationPolicyKey = "ChannelCreationPolicy"

	// ConsortiumsGroupKey is the group name for the consortiums config
	ConsortiumsGroupKey = "Consortiums"

	// ConsortiumKey is the key for the cb.ConfigValue for the Consortium message
	ConsortiumKey = "Consortium"

	// OrdererAddressesKey is the cb.ConfigItem type key name for the OrdererAddresses message
	OrdererAddressesKey = "OrdererAddresses"

	// ConfigChannelName 配置通道名称，区块链网络创世通道
	ConfigChannelName = "genesis"

	// ApplicationGroupKey is the group name for the Application config
	ApplicationGroupKey = "Application"

	// ACLsKey is the name of the ACLs config
	ACLsKey = "ACLs"
)

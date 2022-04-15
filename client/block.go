package client

import (
	"bytes"
	"crypto/sha256"
	"github.com/feng081212/fabric-protos-go/common"
)

func CreateGenesisBlock(channelID string, configGroup *common.ConfigGroup) (*common.Block, error) {

	data := ProtoMarshalIgnoreError(&common.ConfigEnvelope{Config: &common.Config{ChannelGroup: configGroup}})

	payload := CreatePayload(common.HeaderType_CONFIG, channelID, sha256.New(), nil, data, func(header *common.ChannelHeader) {
		header.Version = 1
	})

	envelope := &common.Envelope{Payload: ProtoMarshalIgnoreError(payload), Signature: nil}

	block := NewBlock(0, nil)
	block.Data = &common.BlockData{Data: [][]byte{ProtoMarshalIgnoreError(envelope)}}
	block.Header.DataHash = BlockDataHash(block.Data)
	block.Metadata.Metadata[common.BlockMetadataIndex_LAST_CONFIG] = ProtoMarshalIgnoreError(&common.Metadata{
		Value: ProtoMarshalIgnoreError(&common.LastConfig{Index: 0}),
	})
	block.Metadata.Metadata[common.BlockMetadataIndex_SIGNATURES] = ProtoMarshalIgnoreError(&common.Metadata{
		Value: ProtoMarshalIgnoreError(&common.OrdererBlockMetadata{
			LastConfig: &common.LastConfig{Index: 0},
		}),
	})
	return block, nil
}

// NewBlock constructs a block with no data and no metadata.
func NewBlock(seqNum uint64, previousHash []byte) *common.Block {
	block := &common.Block{}
	block.Header = &common.BlockHeader{}
	block.Header.Number = seqNum
	block.Header.PreviousHash = previousHash
	block.Header.DataHash = []byte{}
	block.Data = &common.BlockData{}

	var metadataContents [][]byte
	for i := 0; i < len(common.BlockMetadataIndex_name); i++ {
		metadataContents = append(metadataContents, []byte{})
	}
	block.Metadata = &common.BlockMetadata{Metadata: metadataContents}

	return block
}

func BlockDataHash(b *common.BlockData) []byte {
	sum := sha256.Sum256(bytes.Join(b.Data, nil))
	return sum[:]
}

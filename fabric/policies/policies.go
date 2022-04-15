package policies

import (
	"github.com/hyperledger/fabric-protos-go/common"
	"github.com/pkg/errors"
	"strings"
)

func ImplicitMetaFromString(input string) (*common.ImplicitMetaPolicy, error) {
	args := strings.Split(input, " ")
	if len(args) != 2 {
		return nil, errors.Errorf("expected two space separated tokens, but got %d", len(args))
	}

	res := &common.ImplicitMetaPolicy{
		SubPolicy: args[1],
	}

	switch args[0] {
	case common.ImplicitMetaPolicy_ANY.String():
		res.Rule = common.ImplicitMetaPolicy_ANY
	case common.ImplicitMetaPolicy_ALL.String():
		res.Rule = common.ImplicitMetaPolicy_ALL
	case common.ImplicitMetaPolicy_MAJORITY.String():
		res.Rule = common.ImplicitMetaPolicy_MAJORITY
	default:
		return nil, errors.Errorf("unknown rule type '%s', expected ALL, ANY, or MAJORITY", args[0])
	}

	return res, nil
}
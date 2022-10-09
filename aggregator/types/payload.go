package types

import (
	"fmt"
)

const ModuleAddress = "0x89576037b3cc0b89645ea393a47787bb348272c76d6941c574b053672b848039"

type EntryFunctionPayload struct {
	Function string
	TypeArgs []string
	Args     []interface{}
}

func buildPayloadOneStepRoute(
	firstDexType uint8,
	firstPoolType uint64,
	firstIsXToY bool,
	xIn uint64,
	yMinOut uint64,
	p []TokenType,
) EntryFunctionPayload {
	// todo abi
	typeArgs := make([]string, len(p))
	for _, item := range p {
		typeArgs = append(typeArgs, item.FullName())
	}
	return EntryFunctionPayload{
		Function: fmt.Sprintf("%s::%s::%s", ModuleAddress, "aggregator", "one_step_route"),
		TypeArgs: typeArgs,
		Args: []interface{}{
			firstDexType,
			firstPoolType,
			firstIsXToY,
			xIn,
			yMinOut,
		},
	}
}

func buildPayloadTwoStepRoute(
	firstDexType uint8,
	firstPoolType uint64,
	firstIsXToY bool,
	secondDexType uint8,
	secondPoolType uint64,
	secondIsXToY bool,
	xIn uint64,
	zMinOut uint64,
	p []TokenType,
) EntryFunctionPayload {
	// todo abi
	typeArgs := make([]string, len(p))
	for _, item := range p {
		typeArgs = append(typeArgs, item.FullName())
	}
	return EntryFunctionPayload{
		Function: fmt.Sprintf("%s::%s::%s", ModuleAddress, "aggregator", "two_step_route"),
		TypeArgs: typeArgs,
		Args: []interface{}{
			firstDexType,
			firstPoolType,
			firstIsXToY,
			secondDexType,
			secondPoolType,
			secondIsXToY,
			xIn,
			zMinOut,
		},
	}
}

func buildPayloadThreeStepRoute(
	firstDexType uint8,
	firstPoolType uint64,
	firstIsXToY bool,
	secondDexType uint8,
	secondPoolType uint64,
	secondIsXToY bool,
	thirdDexType uint8,
	thirdPoolType uint64,
	thirdIsXToY bool,
	xIn uint64,
	mMinOut uint64,
	p []TokenType,
) EntryFunctionPayload {
	// todo abi
	typeArgs := make([]string, len(p))
	for _, item := range p {
		typeArgs = append(typeArgs, item.FullName())
	}
	return EntryFunctionPayload{
		Function: fmt.Sprintf("%s::%s::%s", ModuleAddress, "aggregator", "three_step_route"),
		TypeArgs: typeArgs,
		Args: []interface{}{
			firstDexType,
			firstPoolType,
			firstIsXToY,
			secondDexType,
			secondPoolType,
			secondIsXToY,
			thirdDexType,
			thirdPoolType,
			thirdIsXToY,
			xIn,
			mMinOut,
		},
	}
}

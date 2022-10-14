package types

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

const (
	Bool    = "bool"
	U8      = "u8"
	U64     = "u64"
	U128    = "u128"
	Address = "address"
	Signer  = "signer"
)

var atomicTypeTags = []string{Bool, U8, U64, U128, Address, Signer}
var structTagNameReg *regexp.Regexp = regexp.MustCompile("^[a-zA-Z_][a-zA-Z_0-9]*<")

type TypeTag struct {
	AtomicTypeTag *AtomicTypeTag
	StructTag     *StructTag
	VectorTag     *VectorTag
	TypeParamIdx  *TypeParamIdx
}

func (t TypeTag) isEmpty() bool {
	return t.AtomicTypeTag == nil &&
		t.StructTag == nil &&
		t.VectorTag == nil &&
		t.TypeParamIdx == nil
}

type AtomicTypeTag struct {
	Name string
}

type VectorTag struct {
	TypeParam TypeTag
}

type TypeParamIdx struct {
	ParamIdx int
}

type StructTag struct {
	Address    string
	Module     string
	Name       string
	TypeParams []TypeTag
}

func (t *StructTag) GetFullName() string {
	typeParamStr := getTypeParamsString(t.TypeParams)
	return fmt.Sprintf("%s::%s::%s%s", t.Address, t.Module, t.Name, typeParamStr)
}

func getTypeTagFullName(typeTag TypeTag) string {
	if typeTag.VectorTag != nil {
		return fmt.Sprintf("vector<%s>", getTypeTagFullName(typeTag.VectorTag.TypeParam))
	} else if typeTag.StructTag != nil {
		return typeTag.StructTag.GetFullName()
	} else if typeTag.TypeParamIdx != nil {
		return fmt.Sprintf("$tv%d", typeTag.TypeParamIdx.ParamIdx)
	} else if typeTag.AtomicTypeTag != nil {
		return typeTag.AtomicTypeTag.Name
	} else {
		return ""
	}
}

func getTypeParamsString(typeParams []TypeTag) string {
	if nil == typeParams || len(typeParams) == 0 {
		return ""
	}

	strs := make([]string, len(typeParams))
	for i, tag := range typeParams {
		strs[i] = getTypeTagFullName(tag)
	}
	return "<" + strings.Join(strs, ", ") + ">"
}

func ParseMoveStructTag(tag string) (StructTag, error) {
	result, err := parseTypeTagOrError(tag)
	if err != nil {
		return StructTag{}, err
	}
	if result.StructTag == nil {
		return StructTag{}, errors.New("not struct tag")
	}
	return *result.StructTag, nil
}

func parseTypeTagOrError(name string) (TypeTag, error) {
	tag, remaining, err := parseTypeTag(name)
	if err != nil {
		return TypeTag{}, err
	}
	if tag.isEmpty() || len(remaining) > 0 {
		return TypeTag{}, fmt.Errorf("invalid type tag: %s", name)
	}
	return tag, nil
}

func parseTypeTag(name string) (result TypeTag, remain string, err error) {
	var (
		atomicResult *AtomicTypeTag
		vectorResult *VectorTag
		structResult *StructTag
		tvResult     *TypeParamIdx
	)
	atomicResult, remain = parseAtomicTag(name)
	if atomicResult != nil {
		result.AtomicTypeTag = atomicResult
		return
	}

	vectorResult, remain, err = parseVectorTag(name)
	if err != nil {
		return
	}
	if vectorResult != nil {
		result.VectorTag = vectorResult
		return
	}

	structResult, remain, err = parseQualifiedStructTag(name)
	if err != nil {
		return
	}
	if structResult != nil {
		result.StructTag = structResult
		return
	}

	tvResult, remain, err = parseTypeParameter(name)
	if err != nil {
		return
	}
	if tvResult != nil {
		result.TypeParamIdx = tvResult
		return
	}

	err = fmt.Errorf("bad typetag: %s", name)
	return
}

func parseAtomicTag(name string) (*AtomicTypeTag, string) {
	for _, tag := range atomicTypeTags {
		if strings.HasPrefix(name, tag) {
			if len(name) == len(tag) {
				return &AtomicTypeTag{
					Name: tag,
				}, ""
			} else if name[len(tag)] == ',' || name[len(tag)] == '>' {
				return &AtomicTypeTag{
					Name: tag,
				}, name[len(tag):]
			}
		}
	}
	return nil, name
}

func parseVectorTag(name string) (*VectorTag, string, error) {
	if !strings.HasPrefix(name, "vector<") {
		return nil, name, nil
	}

	elementType, remain, err := parseTypeTag(name[7:])
	if err != nil {
		return nil, "", err
	}
	if elementType.isEmpty() || !strings.HasPrefix(remain, ">") {
		return nil, "", fmt.Errorf("badly formatted vector type name: %s", name)
	}
	return &VectorTag{
		TypeParam: elementType,
	}, remain[1:], nil
}

func parseQualifiedStructTag(name string) (*StructTag, string, error) {
	if !strings.Contains(name, "::") {
		return nil, name, nil
	}
	address, withoutAddress := splitByDoubleColon(name)
	module, withoutModule := splitByDoubleColon(withoutAddress)
	if structTagNameReg.Match([]byte(withoutModule)) {
		leftBracketIdx := strings.Index(withoutModule, "<")
		structName := withoutModule[0:leftBracketIdx]
		afterLeftBracket := withoutModule[leftBracketIdx+1:]
		typeParams := []TypeTag{}
		result, remain, err := parseTypeTag(afterLeftBracket)
		if err != nil {
			return nil, "", err
		}
		for {
			if result.isEmpty() {
				return nil, "", fmt.Errorf("badly formatted struct name: %s", name)
			}
			typeParams = append(typeParams, result)
			if strings.HasPrefix(remain, ">") {
				return &StructTag{
					Address:    address,
					Module:     module,
					Name:       structName,
					TypeParams: typeParams,
				}, remain[1:], nil
			} else if strings.HasPrefix(remain, ", ") {
				result, remain, err = parseTypeTag(remain[2:])
				if err != nil {
					return nil, "", err
				}
			} else if strings.HasPrefix(remain, ",") {
				result, remain, err = parseTypeTag(remain[1:])
				if err != nil {
					return nil, "", err
				}
			} else {
				return nil, "", fmt.Errorf("badly formatted struct name: %s", name)
			}
		}
	} else {
		commaIdx := strings.Index(withoutModule, ",")
		brackIdx := strings.Index(withoutModule, ">")
		if commaIdx == -1 && brackIdx == -1 {
			return &StructTag{
				Address:    address,
				Module:     module,
				Name:       withoutModule,
				TypeParams: []TypeTag{},
			}, "", nil
		}
		separatorIdx := brackIdx
		if commaIdx > -1 {
			separatorIdx = min(commaIdx, brackIdx)
		}
		return &StructTag{
			Address:    address,
			Module:     module,
			Name:       withoutModule[0:separatorIdx],
			TypeParams: []TypeTag{},
		}, withoutModule[separatorIdx:], nil
	}
}

func parseTypeParameter(name string) (*TypeParamIdx, string, error) {
	if !strings.HasPrefix(name, "$tv") {
		return nil, name, nil
	}

	idx := 3
	for ; idx < len(name); idx++ {
		if name[idx] >= '0' && name[idx] <= '9' {
			continue
		}
		break
	}
	if idx == 3 {
		return nil, name, fmt.Errorf("failed to find number after $tv in: %s", name)
	}
	paramIdx, err := strconv.ParseInt(name[3:idx-3], 10, 64)
	if err != nil {
		return nil, name, fmt.Errorf("paramIdx is not integer: %s", name[3:idx-3])
	}
	return &TypeParamIdx{
		ParamIdx: int(paramIdx),
	}, name[idx:], nil
}

func splitByDoubleColon(name string) (string, string) {
	endIdx := strings.Index(name, "::")
	return name[0:endIdx], name[endIdx+2:]
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

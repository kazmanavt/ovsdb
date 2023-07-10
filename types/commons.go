package types

const (
	ZeroUUID = "00000000-0000-0000-0000-000000000000"
)

// AtomicType is the interface that all types that can be used as elements in a Set or Map must implement
type AtomicType interface {
	string | int | bool | float64 | UUIDType
}

type SetType interface {
	Set[string] | Set[int] | Set[bool] | Set[float64] | Set[UUIDType]
}

type MapType interface {
	Map[string, string] | Map[string, int] | Map[string, bool] | Map[string, float64] | Map[string, UUIDType] |
		Map[int, string] | Map[int, int] | Map[int, bool] | Map[int, float64] | Map[int, UUIDType] |
		Map[bool, string] | Map[bool, int] | Map[bool, bool] | Map[bool, float64] | Map[bool, UUIDType] |
		Map[float64, string] | Map[float64, int] | Map[float64, bool] | Map[float64, float64] | Map[float64, UUIDType] |
		Map[UUIDType, string] | Map[UUIDType, int] | Map[UUIDType, bool] | Map[UUIDType, float64] | Map[UUIDType, UUIDType]
}

type BaseType interface {
	AtomicType | SetType | MapType
}

type Updater2 interface {
	Update2(other any) (any, error)
}

func IsSetType(t any) bool {
	switch t.(type) {
	case Set[string], Set[int], Set[bool], Set[float64], Set[UUIDType]:
		return true
	}
	return false
}

func IsMapType(t any) bool {
	switch t.(type) {
	case Map[string, string], Map[string, int], Map[string, bool], Map[string, float64], Map[string, UUIDType],
		Map[int, string], Map[int, int], Map[int, bool], Map[int, float64], Map[int, UUIDType],
		Map[bool, string], Map[bool, int], Map[bool, bool], Map[bool, float64], Map[bool, UUIDType],
		Map[float64, string], Map[float64, int], Map[float64, bool], Map[float64, float64], Map[float64, UUIDType],
		Map[UUIDType, string], Map[UUIDType, int], Map[UUIDType, bool], Map[UUIDType, float64], Map[UUIDType, UUIDType]:
		return true
	}
	return false
}

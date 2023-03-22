package types

const (
	ZeroUUID = "00000000-0000-0000-0000-000000000000"
)

// AtomicType is the interface that all types that can be used as elements in a Set or Map must implement
type AtomicType interface {
	string | int | bool | float64 | UUID | NamedUUID
}

type SetType interface {
	Set[string] | Set[int] | Set[bool] | Set[float64] | Set[UUID] | Set[NamedUUID]
}

type MapType interface {
	Map[string, string] | Map[string, int] | Map[string, bool] | Map[string, float64] | Map[string, UUID] | Map[string, NamedUUID] |
		Map[int, string] | Map[int, int] | Map[int, bool] | Map[int, float64] | Map[int, UUID] | Map[int, NamedUUID] |
		Map[bool, string] | Map[bool, int] | Map[bool, bool] | Map[bool, float64] | Map[bool, UUID] | Map[bool, NamedUUID] |
		Map[float64, string] | Map[float64, int] | Map[float64, bool] | Map[float64, float64] | Map[float64, UUID] | Map[float64, NamedUUID] |
		Map[UUID, string] | Map[UUID, int] | Map[UUID, bool] | Map[UUID, float64] | Map[UUID, UUID] | Map[UUID, NamedUUID] |
		Map[NamedUUID, string] | Map[NamedUUID, int] | Map[NamedUUID, bool] | Map[NamedUUID, float64] | Map[NamedUUID, UUID] | Map[NamedUUID, NamedUUID]
}

type BaseType interface {
	AtomicType | SetType | MapType
}

type Updater2 interface {
	Update2(other any) error
}

//func IsAtomicType(t any) bool {
//	switch t.(type) {
//	case string, int, bool, float64, UUID, NamedUUID:
//		return true
//	}
//	return false
//}

func IsSetType(t any) bool {
	switch t.(type) {
	case Set[string], Set[int], Set[bool], Set[float64], Set[UUID], Set[NamedUUID]:
		return true
	}
	return false
}

func IsMapType(t any) bool {
	switch t.(type) {
	case Map[string, string], Map[string, int], Map[string, bool], Map[string, float64], Map[string, UUID], Map[string, NamedUUID],
		Map[int, string], Map[int, int], Map[int, bool], Map[int, float64], Map[int, UUID], Map[int, NamedUUID],
		Map[bool, string], Map[bool, int], Map[bool, bool], Map[bool, float64], Map[bool, UUID], Map[bool, NamedUUID],
		Map[float64, string], Map[float64, int], Map[float64, bool], Map[float64, float64], Map[float64, UUID], Map[float64, NamedUUID],
		Map[UUID, string], Map[UUID, int], Map[UUID, bool], Map[UUID, float64], Map[UUID, UUID], Map[UUID, NamedUUID],
		Map[NamedUUID, string], Map[NamedUUID, int], Map[NamedUUID, bool], Map[NamedUUID, float64], Map[NamedUUID, UUID], Map[NamedUUID, NamedUUID]:
		return true
	}
	return false
}

//func IsBaseType(t any) bool {
//	return IsAtomicType(t) || IsSetType(t) || IsMapType(t)
//}

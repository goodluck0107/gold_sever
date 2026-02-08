package cast

import (
	"fmt"
	"math"
	"math/big"
	"strconv"
	"strings"
	"time"
)

// ToInt casts an interface to an int type.
func ToInt(i any) int {
	v, _ := ToIntE(i)
	return v
}

// ToIntE casts an interface to an int type.
func ToIntE(a any) (int, error) {
	a = indirectToStringerOrError(a)

	switch v := a.(type) {
	case int:
		return v, nil
	case int8:
		return int(v), nil
	case int16:
		return int(v), nil
	case int32:
		return int(v), nil
	case int64:
		return int(v), nil
	case uint:
		return int(v), nil
	case uint8:
		return int(v), nil
	case uint16:
		return int(v), nil
	case uint32:
		return int(v), nil
	case uint64:
		return int(v), nil
	case float32:
		return int(v), nil
	case float64:
		return int(v), nil
	case *big.Int:
		if v == nil {
			return 0, fmt.Errorf("unable to cast %#v of type %T to int", a, a)
		}
		return int(v.Int64()), nil
	case *big.Float:
		if v == nil {
			return 0, fmt.Errorf("unable to cast %#v of type %T to int", a, a)
		}
		n, _ := v.Int64()
		return int(n), nil
	case *big.Rat:
		if v == nil {
			return 0, fmt.Errorf("unable to cast %#v of type %T to int", a, a)
		}
		n, _ := v.Float64()
		return int(n), nil
	case complex64:
		return int(real(v)), nil
	case complex128:
		return int(real(v)), nil
	case bool:
		if v {
			return 1, nil
		}
		return 0, nil
	case time.Duration:
		return int(v), nil
	case time.Location:
		_, offset := time.Now().In(&v).Zone()
		return int(offset), nil
	case *time.Location:
		_, offset := time.Now().In(v).Zone()
		return int(offset), nil
	case string:
		n, err := dec.ToInt(v)
		if err == nil {
			return int(n), nil
		}
		return 0, fmt.Errorf("unable to cast %#v of type %T to int", a, a)
	case []byte:
		n, err := dec.ToInt(string(v))
		if err == nil {
			return int(n), nil
		}
		return 0, fmt.Errorf("unable to cast %#v of type %T to int", a, a)
	case fmt.Stringer:
		n, err := dec.ToInt(v.String())
		if err == nil {
			return int(n), nil
		}
		return 0, fmt.Errorf("unable to cast %#v of type %T to int", a, a)
	case error:
		n, err := dec.ToInt(v.Error())
		if err == nil {
			return int(n), nil
		}
		return 0, fmt.Errorf("unable to cast %#v of type %T to int", a, a)
	case nil:
		return 0, nil
	default:
		return 0, fmt.Errorf("unable to cast %#v of type %T to int", a, a)
	}
}

// ToInt8 casts an interface to an int8 type.
func ToInt8(i any) int8 {
	v, _ := ToInt8E(i)
	return v
}

// ToInt8E casts an interface to an int8 type.
func ToInt8E(a any) (int8, error) {
	a = indirectToStringerOrError(a)

	switch v := a.(type) {
	case int:
		return int8(v), nil
	case int8:
		return v, nil
	case int16:
		return int8(v), nil
	case int32:
		return int8(v), nil
	case int64:
		return int8(v), nil
	case uint:
		return int8(v), nil
	case uint8:
		return int8(v), nil
	case uint16:
		return int8(v), nil
	case uint32:
		return int8(v), nil
	case uint64:
		return int8(v), nil
	case float32:
		return int8(v), nil
	case float64:
		return int8(v), nil
	case *big.Int:
		if v == nil {
			return 0, fmt.Errorf("unable to cast %#v of type %T to int8", v, v)
		}
		return int8(v.Int64()), nil
	case *big.Float:
		if v == nil {
			return 0, fmt.Errorf("unable to cast %#v of type %T to int8", v, v)
		}
		n, _ := v.Int64()
		return int8(n), nil
	case *big.Rat:
		if v == nil {
			return 0, fmt.Errorf("unable to cast %#v of type %T to int8", v, v)
		}
		n, _ := v.Float64()
		return int8(n), nil
	case complex64:
		return int8(real(v)), nil
	case complex128:
		return int8(real(v)), nil
	case bool:
		if v {
			return 1, nil
		}
		return 0, nil
	case time.Duration:
		return int8(v), nil
	case time.Location:
		_, offset := time.Now().In(&v).Zone()
		return int8(offset), nil
	case *time.Location:
		_, offset := time.Now().In(v).Zone()
		return int8(offset), nil
	case string:
		n, err := dec.ToInt(v)
		if err == nil {
			return int8(n), nil
		}
		return 0, fmt.Errorf("unable to cast %#v of type %T to int8", v, v)
	case []byte:
		n, err := dec.ToInt(string(v))
		if err == nil {
			return int8(n), nil
		}
		return 0, fmt.Errorf("unable to cast %#v of type %T to int8", v, v)
	case fmt.Stringer:
		n, err := dec.ToInt(v.String())
		if err == nil {
			return int8(n), nil
		}
		return 0, fmt.Errorf("unable to cast %#v of type %T to int8", v, v)
	case error:
		n, err := dec.ToInt(v.Error())
		if err == nil {
			return int8(n), nil
		}
		return 0, fmt.Errorf("unable to cast %#v of type %T to int8", v, v)
	case nil:
		return 0, nil
	default:
		return 0, fmt.Errorf("unable to cast %#v of type %T to int8", v, v)
	}
}

// ToInt16 casts an interface to an int16 type.
func ToInt16(i any) int16 {
	v, _ := ToInt16E(i)
	return v
}

// ToInt16E casts an interface to an int16 type.
func ToInt16E(a any) (int16, error) {
	a = indirectToStringerOrError(a)

	switch v := a.(type) {
	case int:
		return int16(v), nil
	case int8:
		return int16(v), nil
	case int16:
		return v, nil
	case int32:
		return int16(v), nil
	case int64:
		return int16(v), nil
	case uint:
		return int16(v), nil
	case uint8:
		return int16(v), nil
	case uint16:
		return int16(v), nil
	case uint32:
		return int16(v), nil
	case uint64:
		return int16(v), nil
	case float32:
		return int16(v), nil
	case float64:
		return int16(v), nil
	case *big.Int:
		if v == nil {
			return 0, fmt.Errorf("unable to cast %#v of type %T to int16", a, a)
		}
		return int16(v.Int64()), nil
	case *big.Float:
		if v == nil {
			return 0, fmt.Errorf("unable to cast %#v of type %T to int16", a, a)
		}
		n, _ := v.Int64()
		return int16(n), nil
	case *big.Rat:
		if v == nil {
			return 0, fmt.Errorf("unable to cast %#v of type %T to int16", a, a)
		}
		n, _ := v.Float64()
		return int16(n), nil
	case complex64:
		return int16(real(v)), nil
	case complex128:
		return int16(real(v)), nil
	case bool:
		if v {
			return 1, nil
		}
		return 0, nil
	case time.Duration:
		return int16(v), nil
	case time.Location:
		_, offset := time.Now().In(&v).Zone()
		return int16(offset), nil
	case *time.Location:
		_, offset := time.Now().In(v).Zone()
		return int16(offset), nil
	case string:
		n, err := dec.ToInt(v)
		if err == nil {
			return int16(n), nil
		}
		return 0, fmt.Errorf("unable to cast %#v of type %T to int16", a, a)
	case []byte:
		n, err := dec.ToInt(string(v))
		if err == nil {
			return int16(n), nil
		}
		return 0, fmt.Errorf("unable to cast %#v of type %T to int16", a, a)
	case fmt.Stringer:
		n, err := dec.ToInt(v.String())
		if err == nil {
			return int16(n), nil
		}
		return 0, fmt.Errorf("unable to cast %#v of type %T to int16", a, a)
	case error:
		n, err := dec.ToInt(v.Error())
		if err == nil {
			return int16(n), nil
		}
		return 0, fmt.Errorf("unable to cast %#v of type %T to int16", a, a)
	case nil:
		return 0, nil
	default:
		return 0, fmt.Errorf("unable to cast %#v of type %T to int16", a, a)
	}
}

// ToInt32 casts an interface to an int32 type.
func ToInt32(i any) int32 {
	v, _ := ToInt32E(i)
	return v
}

// ToInt32E casts an interface to an int32 type.
func ToInt32E(a any) (int32, error) {
	a = indirectToStringerOrError(a)

	switch v := a.(type) {
	case int:
		return int32(v), nil
	case int8:
		return int32(v), nil
	case int16:
		return int32(v), nil
	case int32:
		return v, nil
	case int64:
		return int32(v), nil
	case uint:
		return int32(v), nil
	case uint8:
		return int32(v), nil
	case uint16:
		return int32(v), nil
	case uint32:
		return int32(v), nil
	case uint64:
		return int32(v), nil
	case float32:
		return int32(v), nil
	case float64:
		return int32(v), nil
	case *big.Int:
		if v == nil {
			return 0, fmt.Errorf("unable to cast %#v of type %T to int32", a, a)
		}
		return int32(v.Int64()), nil
	case *big.Float:
		if v == nil {
			return 0, fmt.Errorf("unable to cast %#v of type %T to int32", a, a)
		}
		n, _ := v.Int64()
		return int32(n), nil
	case *big.Rat:
		if v == nil {
			return 0, fmt.Errorf("unable to cast %#v of type %T to int32", a, a)
		}
		n, _ := v.Float64()
		return int32(n), nil
	case complex64:
		return int32(real(v)), nil
	case complex128:
		return int32(real(v)), nil
	case bool:
		if v {
			return 1, nil
		}
		return 0, nil
	case time.Duration:
		return int32(v), nil
	case time.Location:
		_, offset := time.Now().In(&v).Zone()
		return int32(offset), nil
	case *time.Location:
		_, offset := time.Now().In(v).Zone()
		return int32(offset), nil
	case string:
		n, err := dec.ToInt(v)
		if err == nil {
			return int32(n), nil
		}
		return 0, fmt.Errorf("unable to cast %#v of type %T to int32", a, a)
	case []byte:
		n, err := dec.ToInt(string(v))
		if err == nil {
			return int32(n), nil
		}
		return 0, fmt.Errorf("unable to cast %#v of type %T to int32", a, a)
	case fmt.Stringer:
		n, err := dec.ToInt(v.String())
		if err == nil {
			return int32(n), nil
		}
		return 0, fmt.Errorf("unable to cast %#v of type %T to int32", a, a)
	case error:
		n, err := dec.ToInt(v.Error())
		if err == nil {
			return int32(n), nil
		}
		return 0, fmt.Errorf("unable to cast %#v of type %T to int32", a, a)
	case nil:
		return 0, nil
	default:
		return 0, fmt.Errorf("unable to cast %#v of type %T to int32", a, a)
	}
}

// ToInt64 casts an interface to an int64 type.
func ToInt64(i any) int64 {
	v, _ := ToInt64E(i)
	return v
}

// ToInt64E casts an interface to an int64 type.
func ToInt64E(a any) (int64, error) {
	a = indirectToStringerOrError(a)

	switch v := a.(type) {
	case int:
		return int64(v), nil
	case int8:
		return int64(v), nil
	case int16:
		return int64(v), nil
	case int32:
		return int64(v), nil
	case int64:
		return v, nil
	case uint:
		return int64(v), nil
	case uint8:
		return int64(v), nil
	case uint16:
		return int64(v), nil
	case uint32:
		return int64(v), nil
	case uint64:
		return int64(v), nil
	case float32:
		return int64(v), nil
	case float64:
		return int64(v), nil
	case *big.Int:
		if v == nil {
			return 0, fmt.Errorf("unable to cast %#v of type %T to int64", a, a)
		}
		return v.Int64(), nil
	case *big.Float:
		if v == nil {
			return 0, fmt.Errorf("unable to cast %#v of type %T to int64", a, a)
		}
		n, _ := v.Int64()
		return n, nil
	case *big.Rat:
		if v == nil {
			return 0, fmt.Errorf("unable to cast %#v of type %T to int64", a, a)
		}
		n, _ := v.Float64()
		return int64(n), nil
	case complex64:
		return int64(real(v)), nil
	case complex128:
		return int64(real(v)), nil
	case bool:
		if v {
			return 1, nil
		}
		return 0, nil
	case time.Duration:
		return int64(v), nil
	case time.Location:
		_, offset := time.Now().In(&v).Zone()
		return int64(offset), nil
	case *time.Location:
		_, offset := time.Now().In(v).Zone()
		return int64(offset), nil
	case string:
		n, err := dec.ToInt(v)
		if err == nil {
			return n, nil
		}
		return 0, fmt.Errorf("unable to cast %#v of type %T to int64", a, a)
	case []byte:
		n, err := dec.ToInt(string(v))
		if err == nil {
			return n, nil
		}
		return 0, fmt.Errorf("unable to cast %#v of type %T to int64", a, a)
	case fmt.Stringer:
		n, err := dec.ToInt(v.String())
		if err == nil {
			return n, nil
		}
		return 0, fmt.Errorf("unable to cast %#v of type %T to int64", a, a)
	case error:
		n, err := dec.ToInt(v.Error())
		if err == nil {
			return n, nil
		}
		return 0, fmt.Errorf("unable to cast %#v of type %T to int64", a, a)
	case nil:
		return 0, nil
	default:
		return 0, fmt.Errorf("unable to cast %#v of type %T to int64", a, a)
	}
}

// ToUint casts an interface to a uint type.
func ToUint(i any) uint {
	v, _ := ToUintE(i)
	return v
}

// ToUintE casts an interface to a uint type.
func ToUintE(a any) (uint, error) {
	a = indirectToStringerOrError(a)

	switch v := a.(type) {
	case int:
		if v < 0 {
			return 0, fmt.Errorf("unable to cast %#v of type %T to uint", v, v)
		}
		return uint(v), nil
	case int8:
		if v < 0 {
			return 0, fmt.Errorf("unable to cast %#v of type %T to uint", v, v)
		}
		return uint(v), nil
	case int16:
		if v < 0 {
			return 0, fmt.Errorf("unable to cast %#v of type %T to uint", v, v)
		}
		return uint(v), nil
	case int32:
		if v < 0 {
			return 0, fmt.Errorf("unable to cast %#v of type %T to uint", v, v)
		}
		return uint(v), nil
	case int64:
		if v < 0 {
			return 0, fmt.Errorf("unable to cast %#v of type %T to uint", v, v)
		}
		return uint(v), nil
	case uint:
		return v, nil
	case uint8:
		return uint(v), nil
	case uint16:
		return uint(v), nil
	case uint32:
		return uint(v), nil
	case uint64:
		return uint(v), nil
	case float32:
		if v < 0 {
			return 0, fmt.Errorf("unable to cast %#v of type %T to uint", v, v)
		}
		return uint(v), nil
	case float64:
		if v < 0 {
			return 0, fmt.Errorf("unable to cast %#v of type %T to uint", v, v)
		}
		return uint(v), nil
	case *big.Int:
		if v == nil {
			return 0, fmt.Errorf("unable to cast %#v of type %T to uint", a, a)
		}
		if v.Sign() < 0 {
			return 0, fmt.Errorf("unable to cast %#v of type %T to uint", v, v)
		}
		return uint(v.Uint64()), nil
	case *big.Float:
		if v == nil {
			return 0, fmt.Errorf("unable to cast %#v of type %T to uint", a, a)
		}
		if v.Sign() < 0 {
			return 0, fmt.Errorf("unable to cast %#v of type %T to uint", v, v)
		}
		n, _ := v.Uint64()
		return uint(n), nil
	case *big.Rat:
		if v == nil {
			return 0, fmt.Errorf("unable to cast %#v of type %T to uint", a, a)
		}
		if v.Sign() < 0 {
			return 0, fmt.Errorf("unable to cast %#v of type %T to uint", v, v)
		}
		n, _ := v.Float64()
		return uint(n), nil
	case complex64:
		if real(v) < 0 {
			return 0, fmt.Errorf("unable to cast %#v of type %T to uint", v, v)
		}
		return uint(real(v)), nil
	case complex128:
		if real(v) < 0 {
			return 0, fmt.Errorf("unable to cast %#v of type %T to uint", v, v)
		}
		return uint(real(v)), nil
	case bool:
		if v {
			return 1, nil
		}
		return 0, nil
	case time.Duration:
		return uint(v), nil
	case time.Location:
		_, offset := time.Now().In(&v).Zone()
		return uint(offset), nil
	case *time.Location:
		_, offset := time.Now().In(v).Zone()
		return uint(offset), nil
	case string:
		n, err := dec.ToUint(v)
		if err == nil {
			return uint(n), nil
		}
		return 0, fmt.Errorf("unable to cast %#v of type %T to uint", a, a)
	case []byte:
		n, err := dec.ToUint(string(v))
		if err == nil {
			return uint(n), nil
		}
		return 0, fmt.Errorf("unable to cast %#v of type %T to uint", a, a)
	case fmt.Stringer:
		n, err := dec.ToUint(v.String())
		if err == nil {
			return uint(n), nil
		}
		return 0, fmt.Errorf("unable to cast %#v of type %T to uint", a, a)
	case error:
		n, err := dec.ToUint(v.Error())
		if err == nil {
			return uint(n), nil
		}
		return 0, fmt.Errorf("unable to cast %#v of type %T to uint", a, a)
	case nil:
		return 0, nil
	default:
		return 0, fmt.Errorf("unable to cast %#v of type %T to uint", a, a)
	}
}

// ToUint8 casts an interface to a uint8 type.
func ToUint8(i any) uint8 {
	v, _ := ToUint8E(i)
	return v
}

// ToUint8E casts an interface to a uint type.
func ToUint8E(a any) (uint8, error) {
	a = indirectToStringerOrError(a)

	switch v := a.(type) {
	case int:
		if v < 0 {
			return 0, fmt.Errorf("unable to cast %#v of type %T to uint8", v, v)
		}
		return uint8(v), nil
	case int8:
		if v < 0 {
			return 0, fmt.Errorf("unable to cast %#v of type %T to uint8", v, v)
		}
		return uint8(v), nil
	case int16:
		if v < 0 {
			return 0, fmt.Errorf("unable to cast %#v of type %T to uint8", v, v)
		}
		return uint8(v), nil
	case int32:
		if v < 0 {
			return 0, fmt.Errorf("unable to cast %#v of type %T to uint8", v, v)
		}
		return uint8(v), nil
	case int64:
		if v < 0 {
			return 0, fmt.Errorf("unable to cast %#v of type %T to uint8", v, v)
		}
		return uint8(v), nil
	case uint:
		return uint8(v), nil
	case uint8:
		return v, nil
	case uint16:
		return uint8(v), nil
	case uint32:
		return uint8(v), nil
	case uint64:
		return uint8(v), nil
	case float32:
		if v < 0 {
			return 0, fmt.Errorf("unable to cast %#v of type %T to uint8", v, v)
		}
		return uint8(v), nil
	case float64:
		if v < 0 {
			return 0, fmt.Errorf("unable to cast %#v of type %T to uint8", v, v)
		}
		return uint8(v), nil
	case *big.Int:
		if v == nil {
			return 0, fmt.Errorf("unable to cast %#v of type %T to uint8", a, a)
		}
		if v.Sign() < 0 {
			return 0, fmt.Errorf("unable to cast %#v of type %T to uint8", v, v)
		}
		return uint8(v.Uint64()), nil
	case *big.Float:
		if v == nil {
			return 0, fmt.Errorf("unable to cast %#v of type %T to uint8", a, a)
		}
		if v.Sign() < 0 {
			return 0, fmt.Errorf("unable to cast %#v of type %T to uint8", v, v)
		}
		n, _ := v.Uint64()
		return uint8(n), nil
	case *big.Rat:
		if v == nil {
			return 0, fmt.Errorf("unable to cast %#v of type %T to uint8", a, a)
		}
		if v.Sign() < 0 {
			return 0, fmt.Errorf("unable to cast %#v of type %T to uint8", v, v)
		}
		n, _ := v.Float64()
		return uint8(n), nil
	case complex64:
		if real(v) < 0 {
			return 0, fmt.Errorf("unable to cast %#v of type %T to uint8", v, v)
		}
		return uint8(real(v)), nil
	case complex128:
		if real(v) < 0 {
			return 0, fmt.Errorf("unable to cast %#v of type %T to uint8", v, v)
		}
		return uint8(real(v)), nil
	case bool:
		if v {
			return 1, nil
		}
		return 0, nil
	case time.Duration:
		return uint8(v), nil
	case time.Location:
		_, offset := time.Now().In(&v).Zone()
		return uint8(offset), nil
	case *time.Location:
		_, offset := time.Now().In(v).Zone()
		return uint8(offset), nil
	case string:
		n, err := dec.ToUint(v)
		if err == nil {
			return uint8(n), nil
		}
		return 0, fmt.Errorf("unable to cast %#v of type %T to uint8", a, a)
	case []byte:
		n, err := dec.ToUint(string(v))
		if err == nil {
			return uint8(n), nil
		}
		return 0, fmt.Errorf("unable to cast %#v of type %T to uint8", a, a)
	case fmt.Stringer:
		n, err := dec.ToUint(v.String())
		if err == nil {
			return uint8(n), nil
		}
		return 0, fmt.Errorf("unable to cast %#v of type %T to uint8", a, a)
	case error:
		n, err := dec.ToUint(v.Error())
		if err == nil {
			return uint8(n), nil
		}
		return 0, fmt.Errorf("unable to cast %#v of type %T to uint8", a, a)
	case nil:
		return 0, nil
	default:
		return 0, fmt.Errorf("unable to cast %#v of type %T to uint8", a, a)
	}
}

// ToUint16 casts an interface to a uint16 type.
func ToUint16(i any) uint16 {
	v, _ := ToUint16E(i)
	return v
}

// ToUint16E casts an interface to a uint16 type.
func ToUint16E(a any) (uint16, error) {
	a = indirectToStringerOrError(a)

	switch v := a.(type) {
	case int:
		if v < 0 {
			return 0, fmt.Errorf("unable to cast %#v of type %T to uint16", v, v)
		}
		return uint16(v), nil
	case int8:
		if v < 0 {
			return 0, fmt.Errorf("unable to cast %#v of type %T to uint16", v, v)
		}
		return uint16(v), nil
	case int16:
		if v < 0 {
			return 0, fmt.Errorf("unable to cast %#v of type %T to uint16", v, v)
		}
		return uint16(v), nil
	case int32:
		if v < 0 {
			return 0, fmt.Errorf("unable to cast %#v of type %T to uint16", v, v)
		}
		return uint16(v), nil
	case int64:
		if v < 0 {
			return 0, fmt.Errorf("unable to cast %#v of type %T to uint16", v, v)
		}
		return uint16(v), nil
	case uint:
		return uint16(v), nil
	case uint8:
		return uint16(v), nil
	case uint16:
		return v, nil
	case uint32:
		return uint16(v), nil
	case uint64:
		return uint16(v), nil
	case float32:
		if v < 0 {
			return 0, fmt.Errorf("unable to cast %#v of type %T to uint16", v, v)
		}
		return uint16(v), nil
	case float64:
		if v < 0 {
			return 0, fmt.Errorf("unable to cast %#v of type %T to uint16", v, v)
		}
		return uint16(v), nil
	case *big.Int:
		if v == nil {
			return 0, fmt.Errorf("unable to cast %#v of type %T to uint16", a, a)
		}
		if v.Sign() < 0 {
			return 0, fmt.Errorf("unable to cast %#v of type %T to uint16", v, v)
		}
		return uint16(v.Uint64()), nil
	case *big.Float:
		if v == nil {
			return 0, fmt.Errorf("unable to cast %#v of type %T to uint16", a, a)
		}
		if v.Sign() < 0 {
			return 0, fmt.Errorf("unable to cast %#v of type %T to uint16", v, v)
		}
		n, _ := v.Uint64()
		return uint16(n), nil
	case *big.Rat:
		if v == nil {
			return 0, fmt.Errorf("unable to cast %#v of type %T to uint16", a, a)
		}
		if v.Sign() < 0 {
			return 0, fmt.Errorf("unable to cast %#v of type %T to uint16", v, v)
		}
		n, _ := v.Float64()
		return uint16(n), nil
	case complex64:
		if real(v) < 0 {
			return 0, fmt.Errorf("unable to cast %#v of type %T to uint16", v, v)
		}
		return uint16(real(v)), nil
	case complex128:
		if real(v) < 0 {
			return 0, fmt.Errorf("unable to cast %#v of type %T to uint16", v, v)
		}
		return uint16(real(v)), nil
	case bool:
		if v {
			return 1, nil
		}
		return 0, nil
	case time.Duration:
		return uint16(v), nil
	case time.Location:
		_, offset := time.Now().In(&v).Zone()
		return uint16(offset), nil
	case *time.Location:
		_, offset := time.Now().In(v).Zone()
		return uint16(offset), nil
	case string:
		n, err := dec.ToUint(v)
		if err == nil {
			return uint16(n), nil
		}
		return 0, fmt.Errorf("unable to cast %#v of type %T to uint16", a, a)
	case []byte:
		n, err := dec.ToUint(string(v))
		if err == nil {
			return uint16(n), nil
		}
		return 0, fmt.Errorf("unable to cast %#v of type %T to uint16", a, a)
	case fmt.Stringer:
		n, err := dec.ToUint(v.String())
		if err == nil {
			return uint16(n), nil
		}
		return 0, fmt.Errorf("unable to cast %#v of type %T to uint16", a, a)
	case error:
		n, err := dec.ToUint(v.Error())
		if err == nil {
			return uint16(n), nil
		}
		return 0, fmt.Errorf("unable to cast %#v of type %T to uint16", a, a)
	case nil:
		return 0, nil
	default:
		return 0, fmt.Errorf("unable to cast %#v of type %T to uint16", a, a)
	}
}

// ToUint32 casts an interface to a uint32 type.
func ToUint32(i any) uint32 {
	v, _ := ToUint32E(i)
	return v
}

// ToUint32E casts an interface to a uint32 type.
func ToUint32E(a any) (uint32, error) {
	a = indirectToStringerOrError(a)

	switch v := a.(type) {
	case int:
		if v < 0 {
			return 0, fmt.Errorf("unable to cast %#v of type %T to uint32", v, v)
		}
		return uint32(v), nil
	case int8:
		if v < 0 {
			return 0, fmt.Errorf("unable to cast %#v of type %T to uint32", v, v)
		}
		return uint32(v), nil
	case int16:
		if v < 0 {
			return 0, fmt.Errorf("unable to cast %#v of type %T to uint32", v, v)
		}
		return uint32(v), nil
	case int32:
		if v < 0 {
			return 0, fmt.Errorf("unable to cast %#v of type %T to uint32", v, v)
		}
		return uint32(v), nil
	case int64:
		if v < 0 {
			return 0, fmt.Errorf("unable to cast %#v of type %T to uint32", v, v)
		}
		return uint32(v), nil
	case uint:
		return uint32(v), nil
	case uint8:
		return uint32(v), nil
	case uint16:
		return uint32(v), nil
	case uint32:
		return v, nil
	case uint64:
		return uint32(v), nil
	case float32:
		if v < 0 {
			return 0, fmt.Errorf("unable to cast %#v of type %T to uint32", v, v)
		}
		return uint32(v), nil
	case float64:
		if v < 0 {
			return 0, fmt.Errorf("unable to cast %#v of type %T to uint32", v, v)
		}
		return uint32(v), nil
	case *big.Int:
		if v == nil {
			return 0, fmt.Errorf("unable to cast %#v of type %T to uint32", a, a)
		}
		if v.Sign() < 0 {
			return 0, fmt.Errorf("unable to cast %#v of type %T to uint32", v, v)
		}
		return uint32(v.Uint64()), nil
	case *big.Float:
		if v == nil {
			return 0, fmt.Errorf("unable to cast %#v of type %T to uint32", a, a)
		}
		if v.Sign() < 0 {
			return 0, fmt.Errorf("unable to cast %#v of type %T to uint32", v, v)
		}
		n, _ := v.Uint64()
		return uint32(n), nil
	case *big.Rat:
		if v == nil {
			return 0, fmt.Errorf("unable to cast %#v of type %T to uint32", a, a)
		}
		if v.Sign() < 0 {
			return 0, fmt.Errorf("unable to cast %#v of type %T to uint32", v, v)
		}
		n, _ := v.Float64()
		return uint32(n), nil
	case complex64:
		if real(v) < 0 {
			return 0, fmt.Errorf("unable to cast %#v of type %T to uint32", v, v)
		}
		return uint32(real(v)), nil
	case complex128:
		if real(v) < 0 {
			return 0, fmt.Errorf("unable to cast %#v of type %T to uint32", v, v)
		}
		return uint32(real(v)), nil
	case bool:
		if v {
			return 1, nil
		}
		return 0, nil
	case time.Duration:
		return uint32(v), nil
	case time.Location:
		_, offset := time.Now().In(&v).Zone()
		return uint32(offset), nil
	case *time.Location:
		_, offset := time.Now().In(v).Zone()
		return uint32(offset), nil
	case string:
		n, err := dec.ToUint(v)
		if err == nil {
			return uint32(n), nil
		}
		return 0, fmt.Errorf("unable to cast %#v of type %T to uint32", a, a)
	case []byte:
		n, err := dec.ToUint(string(v))
		if err == nil {
			return uint32(n), nil
		}
		return 0, fmt.Errorf("unable to cast %#v of type %T to uint32", a, a)
	case fmt.Stringer:
		n, err := dec.ToUint(v.String())
		if err == nil {
			return uint32(n), nil
		}
		return 0, fmt.Errorf("unable to cast %#v of type %T to uint32", a, a)
	case error:
		n, err := dec.ToUint(v.Error())
		if err == nil {
			return uint32(n), nil
		}
		return 0, fmt.Errorf("unable to cast %#v of type %T to uint32", a, a)
	case nil:
		return 0, nil
	default:
		return 0, fmt.Errorf("unable to cast %#v of type %T to uint32", a, a)
	}
}

// ToUint64 casts an interface to a uint64 type.
func ToUint64(i any) uint64 {
	v, _ := ToUint64E(i)
	return v
}

// ToUint64E casts an interface to a uint64 type.
func ToUint64E(a any) (uint64, error) {
	a = indirectToStringerOrError(a)

	switch v := a.(type) {
	case int:
		if v < 0 {
			return 0, fmt.Errorf("unable to cast %#v of type %T to uint64", v, v)
		}
		return uint64(v), nil
	case int8:
		if v < 0 {
			return 0, fmt.Errorf("unable to cast %#v of type %T to uint64", v, v)
		}
		return uint64(v), nil
	case int16:
		if v < 0 {
			return 0, fmt.Errorf("unable to cast %#v of type %T to uint64", v, v)
		}
		return uint64(v), nil
	case int32:
		if v < 0 {
			return 0, fmt.Errorf("unable to cast %#v of type %T to uint64", v, v)
		}
		return uint64(v), nil
	case int64:
		if v < 0 {
			return 0, fmt.Errorf("unable to cast %#v of type %T to uint64", v, v)
		}
		return uint64(v), nil
	case uint:
		return uint64(v), nil
	case uint8:
		return uint64(v), nil
	case uint16:
		return uint64(v), nil
	case uint32:
		return uint64(v), nil
	case uint64:
		return v, nil
	case float32:
		if v < 0 {
			return 0, fmt.Errorf("unable to cast %#v of type %T to uint64", v, v)
		}
		return uint64(v), nil
	case float64:
		if v < 0 {
			return 0, fmt.Errorf("unable to cast %#v of type %T to uint64", v, v)
		}
		return uint64(v), nil
	case *big.Int:
		if v == nil {
			return 0, fmt.Errorf("unable to cast %#v of type %T to uint64", a, a)
		}
		if v.Sign() < 0 {
			return 0, fmt.Errorf("unable to cast %#v of type %T to uint64", v, v)
		}
		return v.Uint64(), nil
	case *big.Float:
		if v == nil {
			return 0, fmt.Errorf("unable to cast %#v of type %T to uint64", a, a)
		}
		if v.Sign() < 0 {
			return 0, fmt.Errorf("unable to cast %#v of type %T to uint64", v, v)
		}
		n, _ := v.Uint64()
		return n, nil
	case *big.Rat:
		if v == nil {
			return 0, fmt.Errorf("unable to cast %#v of type %T to uint64", a, a)
		}
		if v.Sign() < 0 {
			return 0, fmt.Errorf("unable to cast %#v of type %T to uint64", v, v)
		}
		n, _ := v.Float64()
		return uint64(n), nil
	case complex64:
		if real(v) < 0 {
			return 0, fmt.Errorf("unable to cast %#v of type %T to uint64", v, v)
		}
		return uint64(real(v)), nil
	case complex128:
		if real(v) < 0 {
			return 0, fmt.Errorf("unable to cast %#v of type %T to uint64", v, v)
		}
		return uint64(real(v)), nil
	case bool:
		if v {
			return 1, nil
		}
		return 0, nil
	case time.Duration:
		return uint64(v), nil
	case time.Location:
		_, offset := time.Now().In(&v).Zone()
		return uint64(offset), nil
	case *time.Location:
		_, offset := time.Now().In(v).Zone()
		return uint64(offset), nil
	case string:
		n, err := dec.ToUint(v)
		if err == nil {
			return n, nil
		}
		return 0, fmt.Errorf("unable to cast %#v of type %T to uint64", a, a)
	case []byte:
		n, err := dec.ToUint(string(v))
		if err == nil {
			return n, nil
		}
		return 0, fmt.Errorf("unable to cast %#v of type %T to uint64", a, a)
	case fmt.Stringer:
		n, err := dec.ToUint(v.String())
		if err == nil {
			return n, nil
		}
		return 0, fmt.Errorf("unable to cast %#v of type %T to uint64", a, a)
	case error:
		n, err := dec.ToUint(v.Error())
		if err == nil {
			return n, nil
		}
		return 0, fmt.Errorf("unable to cast %#v of type %T to uint64", a, a)
	case nil:
		return 0, nil
	default:
		return 0, fmt.Errorf("unable to cast %#v of type %T to uint64", a, a)
	}
}

// ToFloat32 casts an interface to a float32 type.
func ToFloat32(i any) float32 {
	v, _ := ToFloat32E(i)
	return v
}

// ToFloat32E casts an interface to a float32 type.
func ToFloat32E(a any) (float32, error) {
	a = indirectToStringerOrError(a)

	switch v := a.(type) {
	case int:
		return float32(v), nil
	case int8:
		return float32(v), nil
	case int16:
		return float32(v), nil
	case int32:
		return float32(v), nil
	case int64:
		return float32(v), nil
	case uint:
		return float32(v), nil
	case uint8:
		return float32(v), nil
	case uint16:
		return float32(v), nil
	case uint32:
		return float32(v), nil
	case uint64:
		return float32(v), nil
	case float32:
		return v, nil
	case float64:
		return float32(v), nil
	case *big.Int:
		if v == nil {
			return 0, fmt.Errorf("unable to cast %#v of type %T to float32", a, a)
		}
		n, _ := new(big.Float).SetInt(v).Float32()
		return n, nil
	case *big.Float:
		if v == nil {
			return 0, fmt.Errorf("unable to cast %#v of type %T to float32", a, a)
		}
		n, _ := v.Float32()
		return n, nil
	case *big.Rat:
		if v == nil {
			return 0, fmt.Errorf("unable to cast %#v of type %T to float32", a, a)
		}
		n, _ := v.Float32()
		return n, nil
	case complex64:
		return real(v), nil
	case complex128:
		return float32(real(v)), nil
	case bool:
		if v {
			return 1, nil
		}
		return 0, nil
	case time.Duration:
		return float32(v), nil
	case time.Location:
		_, offset := time.Now().In(&v).Zone()
		return float32(offset), nil
	case *time.Location:
		_, offset := time.Now().In(v).Zone()
		return float32(offset), nil
	case string:
		n, err := dec.ToFloat32(v)
		if err == nil {
			return n, nil
		}
		return 0, fmt.Errorf("unable to cast %#v of type %T to float32", a, a)
	case []byte:
		n, err := dec.ToFloat32(string(v))
		if err == nil {
			return n, nil
		}
		return 0, fmt.Errorf("unable to cast %#v of type %T to float32", a, a)
	case fmt.Stringer:
		n, err := dec.ToFloat32(v.String())
		if err == nil {
			return n, nil
		}
		return 0, fmt.Errorf("unable to cast %#v of type %T to float32", a, a)
	case error:
		n, err := dec.ToFloat32(v.Error())
		if err == nil {
			return n, nil
		}
		return 0, fmt.Errorf("unable to cast %#v of type %T to float32", a, a)
	case nil:
		return 0, nil
	default:
		return 0, fmt.Errorf("unable to cast %#v of type %T to float32", a, a)
	}
}

// ToFloat64 casts an interface to a float64 type.
func ToFloat64(i any) float64 {
	v, _ := ToFloat64E(i)
	return v
}

// ToFloat64E casts an interface to a float64 type.
func ToFloat64E(a any) (float64, error) {
	a = indirectToStringerOrError(a)

	switch v := a.(type) {
	case int:
		return float64(v), nil
	case int8:
		return float64(v), nil
	case int16:
		return float64(v), nil
	case int32:
		return float64(v), nil
	case int64:
		return float64(v), nil
	case uint:
		return float64(v), nil
	case uint8:
		return float64(v), nil
	case uint16:
		return float64(v), nil
	case uint32:
		return float64(v), nil
	case uint64:
		return float64(v), nil
	case float32:
		return float64(v), nil
	case float64:
		return v, nil
	case *big.Int:
		if v == nil {
			return 0, fmt.Errorf("unable to cast %#v of type %T to float64", a, a)
		}
		n, _ := new(big.Float).SetInt(v).Float64()
		return n, nil
	case *big.Float:
		if v == nil {
			return 0, fmt.Errorf("unable to cast %#v of type %T to float64", a, a)
		}
		n, _ := v.Float64()
		return n, nil
	case *big.Rat:
		if v == nil {
			return 0, fmt.Errorf("unable to cast %#v of type %T to float64", a, a)
		}
		n, _ := v.Float64()
		return n, nil
	case complex64:
		return float64(real(v)), nil
	case complex128:
		return real(v), nil
	case bool:
		if v {
			return 1, nil
		}
		return 0, nil
	case time.Duration:
		return float64(v), nil
	case time.Location:
		_, offset := time.Now().In(&v).Zone()
		return float64(offset), nil
	case *time.Location:
		_, offset := time.Now().In(v).Zone()
		return float64(offset), nil
	case string:
		n, err := dec.ToFloat64(v)
		if err == nil {
			return n, nil
		}
		return 0, fmt.Errorf("unable to cast %#v of type %T to float64", a, a)
	case []byte:
		n, err := dec.ToFloat64(string(v))
		if err == nil {
			return n, nil
		}
		return 0, fmt.Errorf("unable to cast %#v of type %T to float64", a, a)
	case fmt.Stringer:
		n, err := dec.ToFloat64(v.String())
		if err == nil {
			return n, nil
		}
		return 0, fmt.Errorf("unable to cast %#v of type %T to float64", a, a)
	case error:
		n, err := dec.ToFloat64(v.Error())
		if err == nil {
			return n, nil
		}
		return 0, fmt.Errorf("unable to cast %#v of type %T to float64", a, a)
	case nil:
		return 0, nil
	default:
		return 0, fmt.Errorf("unable to cast %#v of type %T to float64", a, a)
	}
}

// ToBigInt casts an interface to a *big.Int type.
func ToBigInt(i any) *big.Int {
	v, _ := ToBigIntE(i)
	return v
}

// ToBigIntE casts an interface to a *big.Int type.
func ToBigIntE(a any) (*big.Int, error) {
	a = indirectToStringerOrError(a)

	switch v := a.(type) {
	case int:
		return big.NewInt(0).SetInt64(int64(v)), nil
	case int8:
		return big.NewInt(0).SetInt64(int64(v)), nil
	case int16:
		return big.NewInt(0).SetInt64(int64(v)), nil
	case int32:
		return big.NewInt(0).SetInt64(int64(v)), nil
	case int64:
		return big.NewInt(0).SetInt64(v), nil
	case uint:
		return big.NewInt(0).SetUint64(uint64(v)), nil
	case uint8:
		return big.NewInt(0).SetUint64(uint64(v)), nil
	case uint16:
		return big.NewInt(0).SetUint64(uint64(v)), nil
	case uint32:
		return big.NewInt(0).SetUint64(uint64(v)), nil
	case uint64:
		return big.NewInt(0).SetUint64(v), nil
	case float32:
		if math.IsInf(float64(v), 0) || math.IsNaN(float64(v)) {
			return big.NewInt(0), fmt.Errorf("unable to cast %#v of type %T to big.Float", a, a)
		}
		n, _ := big.NewFloat(float64(v)).Int(nil)
		return n, nil
	case float64:
		if math.IsInf(v, 0) || math.IsNaN(v) {
			return big.NewInt(0), fmt.Errorf("unable to cast %#v of type %T to big.Float", a, a)
		}
		n, _ := big.NewFloat(v).Int(nil)
		return n, nil
	case *big.Int:
		if v == nil {
			return big.NewInt(0), fmt.Errorf("unable to cast %#v of type %T to *big.Int", a, a)
		}
		return v, nil
	case *big.Float:
		if v == nil {
			return big.NewInt(0), fmt.Errorf("unable to cast %#v of type %T to *big.Int", a, a)
		}
		n, _ := v.Int(nil)
		return n, nil
	case *big.Rat:
		if v == nil {
			return big.NewInt(0), fmt.Errorf("unable to cast %#v of type %T to *big.Int", a, a)
		}
		n, _ := v.Float64()
		return big.NewInt(int64(n)), nil
	case complex64:
		return big.NewInt(0).SetInt64(int64(real(v))), nil
	case complex128:
		return big.NewInt(0).SetInt64(int64(real(v))), nil
	case bool:
		if v {
			return big.NewInt(1), nil
		}
		return big.NewInt(0), nil
	case time.Duration:
		return big.NewInt(0).SetInt64(int64(v)), nil
	case time.Location:
		_, offset := time.Now().In(&v).Zone()
		return big.NewInt(0).SetInt64(int64(offset)), nil
	case *time.Location:
		_, offset := time.Now().In(v).Zone()
		return big.NewInt(0).SetInt64(int64(offset)), nil
	case string:
		return dec.ToBigInt(v)
	case []byte:
		return dec.ToBigInt(string(v))
	case fmt.Stringer:
		return dec.ToBigInt(v.String())
	case error:
		return dec.ToBigInt(v.Error())
	case nil:
		return big.NewInt(0), nil
	default:
		return big.NewInt(0), fmt.Errorf("unable to cast %#v of type %T to *big.Int", a, a)
	}
}

// ToBigFloat casts an interface to a *big.Float type.
func ToBigFloat(i any) *big.Float {
	v, _ := ToBigFloatE(i)
	return v
}

// ToBigFloatE casts an interface to a *big.Float type.
func ToBigFloatE(a any) (*big.Float, error) {
	a = indirectToStringerOrError(a)

	switch v := a.(type) {
	case int:
		return big.NewFloat(0).SetInt64(int64(v)), nil
	case int8:
		return big.NewFloat(0).SetInt64(int64(v)), nil
	case int16:
		return big.NewFloat(0).SetInt64(int64(v)), nil
	case int32:
		return big.NewFloat(0).SetInt64(int64(v)), nil
	case int64:
		return big.NewFloat(0).SetInt64(v), nil
	case uint:
		return big.NewFloat(0).SetUint64(uint64(v)), nil
	case uint8:
		return big.NewFloat(0).SetUint64(uint64(v)), nil
	case uint16:
		return big.NewFloat(0).SetUint64(uint64(v)), nil
	case uint32:
		return big.NewFloat(0).SetUint64(uint64(v)), nil
	case uint64:
		return big.NewFloat(0).SetUint64(v), nil
	case float32:
		if math.IsInf(float64(v), 0) || math.IsNaN(float64(v)) {
			return big.NewFloat(0), fmt.Errorf("unable to cast %#v of type %T to big.Float", a, a)
		}
		return big.NewFloat(float64(v)), nil
	case float64:
		if math.IsInf(v, 0) || math.IsNaN(v) {
			return big.NewFloat(0), fmt.Errorf("unable to cast %#v of type %T to big.Float", a, a)
		}
		return big.NewFloat(v), nil
	case *big.Int:
		if v == nil {
			return big.NewFloat(0), fmt.Errorf("unable to cast %#v of type %T to *big.Float", a, a)
		}
		return big.NewFloat(0).SetInt(v), nil
	case *big.Float:
		if v == nil {
			return big.NewFloat(0), fmt.Errorf("unable to cast %#v of type %T to *big.Float", a, a)
		}
		return v, nil
	case *big.Rat:
		if v == nil {
			return big.NewFloat(0), fmt.Errorf("unable to cast %#v of type %T to *big.Float", a, a)
		}
		n, _ := v.Float64()
		return big.NewFloat(n), nil
	case complex64:
		return big.NewFloat(0).SetInt64(int64(real(v))), nil
	case complex128:
		return big.NewFloat(0).SetInt64(int64(real(v))), nil
	case bool:
		if v {
			return big.NewFloat(1), nil
		}
		return big.NewFloat(0), nil
	case time.Duration:
		return big.NewFloat(0).SetInt64(int64(v)), nil
	case time.Location:
		_, offset := time.Now().In(&v).Zone()
		return big.NewFloat(0).SetInt64(int64(offset)), nil
	case *time.Location:
		_, offset := time.Now().In(v).Zone()
		return big.NewFloat(0).SetInt64(int64(offset)), nil
	case string:
		return dec.ToBigFloat(v)
	case []byte:
		return dec.ToBigFloat(string(v))
	case fmt.Stringer:
		return dec.ToBigFloat(v.String())
	case error:
		return dec.ToBigFloat(v.Error())
	case nil:
		return big.NewFloat(0), nil
	default:
		return big.NewFloat(0), fmt.Errorf("unable to cast %#v of type %T to *big.Float", a, a)
	}
}

// ToBigRat casts an interface to a *big.Rat type.
func ToBigRat(i any) *big.Rat {
	v, _ := ToBigRatE(i)
	return v
}

// ToBigRatE casts an interface to a *big.Rat type.
func ToBigRatE(a any) (*big.Rat, error) {
	a = indirectToStringerOrError(a)

	switch v := a.(type) {
	case int:
		return big.NewRat(int64(v), 1), nil
	case int8:
		return big.NewRat(int64(v), 1), nil
	case int16:
		return big.NewRat(int64(v), 1), nil
	case int32:
		return big.NewRat(int64(v), 1), nil
	case int64:
		return big.NewRat(v, 1), nil
	case uint:
		return big.NewRat(int64(v), 1), nil
	case uint8:
		return big.NewRat(int64(v), 1), nil
	case uint16:
		return big.NewRat(int64(v), 1), nil
	case uint32:
		return big.NewRat(int64(v), 1), nil
	case uint64:
		return big.NewRat(int64(v), 1), nil
	case float32:
		if math.IsInf(float64(v), 0) || math.IsNaN(float64(v)) {
			return big.NewRat(0, 1), fmt.Errorf("unable to cast %#v of type %T to big.Rat", a, a)
		}
		return big.NewRat(0, 1).SetFloat64(float64(v)), nil
	case float64:
		if math.IsInf(v, 0) || math.IsNaN(v) {
			return big.NewRat(0, 1), fmt.Errorf("unable to cast %#v of type %T to big.Rat", a, a)
		}
		return big.NewRat(0, 1).SetFloat64(v), nil
	case *big.Int:
		if v == nil {
			return big.NewRat(0, 1), fmt.Errorf("unable to cast %#v of type %T to *big.Rat", a, a)
		}
		return big.NewRat(0, 1).SetFloat64(float64(v.Int64())), nil
	case *big.Float:
		if v == nil {
			return big.NewRat(0, 1), fmt.Errorf("unable to cast %#v of type %T to *big.Rat", a, a)
		}
		n, _ := v.Float64()
		return big.NewRat(0, 1).SetFloat64(n), nil
	case *big.Rat:
		if v == nil {
			return big.NewRat(0, 1), fmt.Errorf("unable to cast %#v of type %T to *big.Rat", a, a)
		}
		return v, nil
	case complex64:
		return big.NewRat(0, 1).SetFloat64(float64(real(v))), nil
	case complex128:
		return big.NewRat(0, 1).SetFloat64(real(v)), nil
	case bool:
		if v {
			return big.NewRat(1, 1), nil
		}
		return big.NewRat(0, 1), nil
	case time.Duration:
		return big.NewRat(0, 1).SetInt64(int64(v)), nil
	case time.Location:
		_, offset := time.Now().In(&v).Zone()
		return big.NewRat(0, 1).SetInt64(int64(offset)), nil
	case *time.Location:
		_, offset := time.Now().In(v).Zone()
		return big.NewRat(0, 1).SetInt64(int64(offset)), nil
	case string:
		return dec.ToBigRat(v)
	case []byte:
		return dec.ToBigRat(string(v))
	case fmt.Stringer:
		return dec.ToBigRat(v.String())
	case error:
		return dec.ToBigRat(v.Error())
	case nil:
		return big.NewRat(0, 1), nil
	default:
		return big.NewRat(0, 1), fmt.Errorf("unable to cast %#v of type %T to *big.Rat", a, a)
	}
}

// ToComplex64 casts an interface to a complex64 type.
func ToComplex64(i any) complex64 {
	v, _ := ToComplex64E(i)
	return v
}

// ToComplex64E casts an interface to a complex64 type.
func ToComplex64E(a any) (complex64, error) {
	a = indirectToStringerOrError(a)

	switch v := a.(type) {
	case int:
		return complex(float32(v), 0), nil
	case int8:
		return complex(float32(v), 0), nil
	case int16:
		return complex(float32(v), 0), nil
	case int32:
		return complex(float32(v), 0), nil
	case int64:
		return complex(float32(v), 0), nil
	case uint:
		return complex(float32(v), 0), nil
	case uint8:
		return complex(float32(v), 0), nil
	case uint16:
		return complex(float32(v), 0), nil
	case uint32:
		return complex(float32(v), 0), nil
	case uint64:
		return complex(float32(v), 0), nil
	case float32:
		return complex(v, 0), nil
	case float64:
		return complex(float32(v), 0), nil
	case *big.Int:
		if v == nil {
			return 0, fmt.Errorf("unable to cast %#v of type %T to complex64", a, a)
		}
		n, _ := new(big.Float).SetInt(v).Float32()
		return complex(n, 0), nil
	case *big.Float:
		if v == nil {
			return 0, fmt.Errorf("unable to cast %#v of type %T to complex64", a, a)
		}
		n, _ := v.Float32()
		return complex(n, 0), nil
	case *big.Rat:
		if v == nil {
			return 0, fmt.Errorf("unable to cast %#v of type %T to complex64", a, a)
		}
		n, _ := new(big.Float).SetRat(v).Float32()
		return complex(n, 0), nil
	case complex64:
		return v, nil
	case complex128:
		return complex64(v), nil
	case bool:
		if v {
			return complex(1, 0), nil
		}
		return complex(0, 0), nil
	case time.Duration:
		return complex(float32(v), 0), nil
	case time.Location:
		_, offset := time.Now().In(&v).Zone()
		return complex(float32(offset), 0), nil
	case *time.Location:
		_, offset := time.Now().In(v).Zone()
		return complex(float32(offset), 0), nil
	case string:
		return dec.ToComplex64(v)
	case []byte:
		return dec.ToComplex64(string(v))
	case fmt.Stringer:
		return dec.ToComplex64(v.String())
	case error:
		return dec.ToComplex64(v.Error())
	case nil:
		return complex(0, 0), nil
	default:
		return 0, fmt.Errorf("unable to cast %#v of type %T to complex64", a, a)
	}
}

// ToComplex128 casts an interface to a complex128 type.
func ToComplex128(i any) complex128 {
	v, _ := ToComplex128E(i)
	return v
}

// ToComplex128E casts an interface to a complex128 type.
func ToComplex128E(a any) (complex128, error) {
	a = indirectToStringerOrError(a)

	switch v := a.(type) {
	case int:
		return complex(float64(v), 0), nil
	case int8:
		return complex(float64(v), 0), nil
	case int16:
		return complex(float64(v), 0), nil
	case int32:
		return complex(float64(v), 0), nil
	case int64:
		return complex(float64(v), 0), nil
	case uint:
		return complex(float64(v), 0), nil
	case uint8:
		return complex(float64(v), 0), nil
	case uint16:
		return complex(float64(v), 0), nil
	case uint32:
		return complex(float64(v), 0), nil
	case uint64:
		return complex(float64(v), 0), nil
	case float32:
		f := float64(v)
		c := complex(f, 0)
		return c, nil
	case float64:
		return complex(v, 0), nil
	case *big.Int:
		if v == nil {
			return 0, fmt.Errorf("unable to cast %#v of type %T to complex128", a, a)
		}
		n, _ := new(big.Float).SetInt(v).Float64()
		return complex(n, 0), nil
	case *big.Float:
		if v == nil {
			return 0, fmt.Errorf("unable to cast %#v of type %T to complex128", a, a)
		}
		n, _ := v.Float64()
		return complex(n, 0), nil
	case *big.Rat:
		if v == nil {
			return 0, fmt.Errorf("unable to cast %#v of type %T to complex128", a, a)
		}
		n, _ := new(big.Float).SetRat(v).Float64()
		return complex(n, 0), nil
	case complex64:
		return complex128(v), nil
	case complex128:
		return v, nil
	case bool:
		if v {
			return complex(1, 0), nil
		}
		return complex(0, 0), nil
	case time.Duration:
		return complex(float64(v), 0), nil
	case time.Location:
		_, offset := time.Now().In(&v).Zone()
		return complex(float64(offset), 0), nil
	case *time.Location:
		_, offset := time.Now().In(v).Zone()
		return complex(float64(offset), 0), nil
	case string:
		return dec.ToComplex128(v)
	case []byte:
		return dec.ToComplex128(string(v))
	case fmt.Stringer:
		return dec.ToComplex128(v.String())
	case error:
		return dec.ToComplex128(v.Error())
	case nil:
		return complex(0, 0), nil
	default:
		return 0, fmt.Errorf("unable to cast %#v of type %T to complex128", a, a)
	}
}

// ToBool casts an interface to a bool type.
func ToBool(i any) bool {
	v, _ := ToBoolE(i)
	return v
}

// ToBoolE casts an interface to a bool type.
func ToBoolE(a any) (bool, error) {
	a = indirectToStringerOrError(a)

	switch v := a.(type) {
	case int:
		return v != 0, nil
	case int8:
		return v != 0, nil
	case int16:
		return v != 0, nil
	case int32:
		return v != 0, nil
	case int64:
		return v != 0, nil
	case uint:
		return v != 0, nil
	case uint8:
		return v != 0, nil
	case uint16:
		return v != 0, nil
	case uint32:
		return v != 0, nil
	case uint64:
		return v != 0, nil
	case float32:
		return v != 0, nil
	case float64:
		return v != 0, nil
	case *big.Int:
		if v == nil {
			return false, fmt.Errorf("unable to cast %#v of type %T to bool", a, a)
		}
		return v.Sign() != 0, nil
	case *big.Float:
		if v == nil {
			return false, fmt.Errorf("unable to cast %#v of type %T to bool", a, a)
		}
		return v.Sign() != 0, nil
	case *big.Rat:
		if v == nil {
			return false, fmt.Errorf("unable to cast %#v of type %T to bool", a, a)
		}
		return v.Sign() != 0, nil
	case complex64:
		return real(v) != 0 || imag(v) != 0, nil
	case complex128:
		return real(v) != 0 || imag(v) != 0, nil
	case bool:
		return v, nil
	case time.Duration:
		return int64(v) != 0, nil
	case time.Location:
		_, offset := time.Now().In(&v).Zone()
		return offset != 0, nil
	case *time.Location:
		_, offset := time.Now().In(v).Zone()
		return offset != 0, nil
	case string:
		return dec.ToBool(v)
	case []byte:
		return dec.ToBool(string(v))
	case fmt.Stringer:
		return dec.ToBool(v.String())
	case error:
		return dec.ToBool(v.Error())
	case nil:
		return false, nil
	default:
		return false, fmt.Errorf("unable to cast %#v of type %T to bool", a, a)
	}
}

type decimalParser struct{}

var dec = decimalParser{}

func (p decimalParser) ToInt(s string) (int64, error) {
	switch {
	case strings.ContainsAny(s, "tf"):
		b, err := strconv.ParseBool(s)
		if err != nil {
			return 0, err
		}
		if b {
			return 1, nil
		}
		return 0, nil
	case strings.ContainsAny(s, "(i)"):
		n, err := strconv.ParseComplex(s, 128)
		if err != nil {
			return 0, err
		}
		return int64(real(n)), nil
	case strings.Contains(s, "/"):
		n, ok := big.NewRat(0, 1).SetString(s)
		if !ok {
			return 0, fmt.Errorf("unable to cast %#v of type %T to *big.Rat", s, s)
		}
		f, _ := n.Float64()
		return int64(f), nil
	default:
		s = p.trimPointZeroOfIntString(s)
		n, ok := big.NewInt(0).SetString(s, 0)
		if !ok {
			return 0, fmt.Errorf("unable to cast %#v of type %T to *big.Int", s, s)
		}
		return n.Int64(), nil
	}
}

func (p decimalParser) ToUint(s string) (uint64, error) {
	switch {
	case strings.ContainsAny(s, "tf"):
		b, err := strconv.ParseBool(s)
		if err != nil {
			return 0, err
		}
		if b {
			return 1, nil
		}
		return 0, nil
	case strings.ContainsAny(s, "(i)"):
		n, err := strconv.ParseComplex(s, 128)
		if err != nil {
			return 0, err
		}
		return uint64(real(n)), nil
	case strings.Contains(s, "/"):
		n, ok := big.NewRat(0, 1).SetString(s)
		if !ok {
			return 0, fmt.Errorf("unable to cast %#v of type %T to *big.Rat", s, s)
		}
		f, _ := n.Float64()
		return uint64(f), nil
	default:
		s = p.trimPointZeroOfIntString(s)
		n, ok := big.NewInt(0).SetString(s, 0)
		if !ok {
			return 0, fmt.Errorf("unable to cast %#v of type %T to *big.Int", s, s)
		}
		if n.Sign() < 0 {
			return 0, fmt.Errorf("unable to cast %#v of type %T to uint64", s, s)
		}
		return n.Uint64(), nil
	}
}

func (p decimalParser) ToFloat32(s string) (float32, error) {
	switch {
	case strings.ContainsAny(s, "tf"):
		b, err := strconv.ParseBool(s)
		if err != nil {
			return 0, err
		}
		if b {
			return 1, nil
		}
		return 0, nil
	case strings.ContainsAny(s, "(i)"):
		n, err := strconv.ParseComplex(s, 128)
		if err != nil {
			return 0, err
		}
		return float32(real(n)), nil
	case strings.Contains(s, "/"):
		n, ok := big.NewRat(0, 1).SetString(s)
		if !ok {
			return 0, fmt.Errorf("unable to cast %#v of type %T to *big.Rat", s, s)
		}
		f, _ := n.Float32()
		return f, nil
	default:
		f, ok := big.NewFloat(0).SetString(s)
		if !ok {
			return 0, fmt.Errorf("unable to cast %#v of type %T to *big.Float", s, s)
		}
		f32, _ := f.Float32()
		return f32, nil
	}
}

func (p decimalParser) ToFloat64(s string) (float64, error) {
	switch {
	case strings.ContainsAny(s, "tf"):
		b, err := strconv.ParseBool(s)
		if err != nil {
			return 0, err
		}
		if b {
			return 1, nil
		}
		return 0, nil
	case strings.ContainsAny(s, "(i)"):
		n, err := strconv.ParseComplex(s, 128)
		if err != nil {
			return 0, err
		}
		return real(n), nil
	case strings.Contains(s, "/"):
		n, ok := big.NewRat(0, 1).SetString(s)
		if !ok {
			return 0, fmt.Errorf("unable to cast %#v of type %T to *big.Rat", s, s)
		}
		f, _ := n.Float64()
		return f, nil
	default:
		f, ok := big.NewFloat(0).SetString(s)
		if !ok {
			return 0, fmt.Errorf("unable to cast %#v of type %T to *big.Float", s, s)
		}
		f64, _ := f.Float64()
		return f64, nil
	}
}

func (p decimalParser) ToBigInt(s string) (*big.Int, error) {
	switch {
	case strings.ContainsAny(s, "tf"):
		b, err := strconv.ParseBool(s)
		if err != nil {
			return nil, err
		}
		if b {
			return big.NewInt(1), nil
		}
		return big.NewInt(0), nil
	case strings.ContainsAny(s, "(i)"):
		n, err := strconv.ParseComplex(s, 128)
		if err != nil {
			return nil, err
		}
		return big.NewInt(int64(real(n))), nil
	case strings.Contains(s, "/"):
		n, ok := big.NewRat(0, 1).SetString(s)
		if !ok {
			return nil, fmt.Errorf("unable to cast %#v of type %T to *big.Rat", s, s)
		}
		f, _ := n.Float64()
		return big.NewInt(int64(f)), nil
	default:
		s = p.trimPointZeroOfIntString(s)
		n, ok := big.NewInt(0).SetString(s, 0)
		if !ok {
			return nil, fmt.Errorf("unable to cast %#v of type %T to *big.Int", s, s)
		}
		return n, nil
	}
}

func (p decimalParser) ToBigFloat(s string) (*big.Float, error) {
	switch {
	case strings.ContainsAny(s, "tf"):
		b, err := strconv.ParseBool(s)
		if err != nil {
			return nil, err
		}
		if b {
			return big.NewFloat(1), nil
		}
		return big.NewFloat(0), nil
	case strings.ContainsAny(s, "(i)"):
		n, err := strconv.ParseComplex(s, 128)
		if err != nil {
			return nil, err
		}
		return big.NewFloat(real(n)), nil
	case strings.Contains(s, "/"):
		n, ok := big.NewRat(0, 1).SetString(s)
		if !ok {
			return nil, fmt.Errorf("unable to cast %#v of type %T to *big.Rat", s, s)
		}
		f, _ := n.Float64()
		return big.NewFloat(f), nil
	default:
		f, ok := big.NewFloat(0).SetString(s)
		if !ok {
			return nil, fmt.Errorf("unable to cast %#v of type %T to *big.Float", s, s)
		}
		return f, nil
	}
}

func (p decimalParser) ToBigRat(s string) (*big.Rat, error) {
	switch {
	case strings.ContainsAny(s, "tf"):
		b, err := strconv.ParseBool(s)
		if err != nil {
			return nil, err
		}
		if b {
			return big.NewRat(1, 1), nil
		}
		return big.NewRat(0, 1), nil
	case strings.ContainsAny(s, "(i)"):
		n, err := strconv.ParseComplex(s, 128)
		if err != nil {
			return nil, err
		}
		return big.NewRat(int64(real(n)), 1), nil
	case strings.Contains(s, "/"):
		n, ok := big.NewRat(0, 1).SetString(s)
		if !ok {
			return nil, fmt.Errorf("unable to cast %#v of type %T to *big.Rat", s, s)
		}
		return n, nil
	default:
		f, ok := big.NewRat(0, 1).SetString(s)
		if !ok {
			return nil, fmt.Errorf("unable to cast %#v of type %T to *big.Rat", s, s)
		}
		return f, nil
	}
}

func (p decimalParser) ToComplex64(s string) (complex64, error) {
	switch {
	case strings.ContainsAny(s, "tf"):
		b, err := strconv.ParseBool(s)
		if err != nil {
			return complex(0, 0), err
		}
		if b {
			return complex(1, 0), nil
		}
		return complex(0, 0), nil
	case strings.ContainsAny(s, "(i)"):
		n, err := strconv.ParseComplex(s, 64)
		if err != nil {
			return 0, err
		}
		return complex64(n), nil
	case strings.Contains(s, "/"):
		n, ok := big.NewRat(0, 1).SetString(s)
		if !ok {
			return 0, fmt.Errorf("unable to cast %#v of type %T to *big.Rat", s, s)
		}
		f, _ := n.Float32()
		return complex(f, 0), nil
	default:
		n, err := strconv.ParseComplex(s, 64)
		if err != nil {
			return 0, err
		}
		return complex64(n), nil
	}
}

func (p decimalParser) ToComplex128(s string) (complex128, error) {
	switch {
	case strings.ContainsAny(s, "tf"):
		b, err := strconv.ParseBool(s)
		if err != nil {
			return complex(0, 0), err
		}
		if b {
			return complex(1, 0), nil
		}
		return complex(0, 0), nil
	case strings.ContainsAny(s, "(i)"):
		n, err := strconv.ParseComplex(s, 128)
		if err != nil {
			return 0, err
		}
		return n, nil
	case strings.Contains(s, "/"):
		n, ok := big.NewRat(0, 1).SetString(s)
		if !ok {
			return 0, fmt.Errorf("unable to cast %#v of type %T to *big.Rat", s, s)
		}
		f, _ := n.Float64()
		return complex(f, 0), nil
	default:
		n, err := strconv.ParseComplex(s, 128)
		if err != nil {
			return 0, err
		}
		return n, nil
	}
}

func (p decimalParser) ToBool(s string) (bool, error) {
	switch {
	case strings.ContainsAny(s, "tf"):
		return strconv.ParseBool(s)
	case strings.ContainsAny(s, "(i)"):
		n, err := strconv.ParseComplex(s, 128)
		if err != nil {
			return false, err
		}
		return real(n) != 0, nil
	case strings.Contains(s, "/"):
		n, ok := big.NewRat(0, 1).SetString(s)
		if !ok {
			return false, fmt.Errorf("unable to cast %#v of type %T to *big.Rat", s, s)
		}
		f, _ := n.Float64()
		return f != 0, nil
	default:
		f, ok := big.NewFloat(0).SetString(s)
		if !ok {
			return false, fmt.Errorf("unable to cast %#v of type %T to *big.Float", s, s)
		}
		f64, _ := f.Float64()
		return f64 != 0, nil
	}
}

func (p decimalParser) trimPointZeroOfIntString(s string) string {
	var foundZero bool
	for i := len(s); i > 0; i-- {
		switch s[i-1] {
		case '.':
			if foundZero {
				s = s[:i-1]
				return s
			}
		case '0':
			foundZero = true
		default:
			return s
		}
	}
	return s
}

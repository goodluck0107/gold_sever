package cast

import (
	"errors"
	"fmt"
	"math/big"
	"reflect"
	"strconv"
	"unsafe"

	"github.com/shopspring/decimal"
)

// ToString casts an interface to a string type.
func ToString(a any) string {
	v, _ := ToStringE(a)
	return v
}

// ToStringE casts an interface to a string type.
func ToStringE(a any) (string, error) {
	a = indirectToStringerOrError(a)

	switch v := a.(type) {
	case int:
		return strconv.Itoa(v), nil
	case int8:
		return strconv.FormatInt(int64(v), 10), nil
	case int16:
		return strconv.FormatInt(int64(v), 10), nil
	case int32:
		return strconv.Itoa(int(v)), nil
	case int64:
		return strconv.FormatInt(v, 10), nil
	case uint:
		return strconv.FormatUint(uint64(v), 10), nil
	case uint8:
		return strconv.FormatUint(uint64(v), 10), nil
	case uint16:
		return strconv.FormatUint(uint64(v), 10), nil
	case uint32:
		return strconv.FormatUint(uint64(v), 10), nil
	case uint64:
		return strconv.FormatUint(v, 10), nil
	case float32:
		// Use decimal to fix precision issue, FormatFloat is unstable.
		// Optional:
		//		return strconv.FormatFloat(float64(s), 'f', -1, 32), nil
		return decimal.NewFromFloat32(v).String(), nil
	case float64:
		// Use decimal to fix precision issue, FormatFloat is unstable.
		// Optional:
		// 		return strconv.FormatFloat(s, 'f', -1, 64), nil
		return decimal.NewFromFloat(v).String(), nil
	case *big.Int:
		return v.String(), nil
	case *big.Float:
		return v.String(), nil
	case *big.Rat:
		return v.String(), nil
	case complex64:
		return fmt.Sprintf("(%v+%vi)", real(v), imag(v)), nil
	case complex128:
		return fmt.Sprintf("(%v+%vi)", real(v), imag(v)), nil
	case bool:
		return strconv.FormatBool(v), nil
	case string:
		return v, nil
	case []byte:
		return string(v), nil
	case fmt.Stringer:
		return v.String(), nil
	case error:
		return v.Error(), nil
	case nil:
		return "", nil
	default:
		return "", fmt.Errorf("unable to cast %#v of type %T to string", a, a)
	}
}

// ToBytes casts an interface to a []byte type.
func ToBytes(a any) []byte {
	v, _ := ToBytesE(a)
	return v
}

// ToBytesE casts an interface to a []byte type.
func ToBytesE(a any) ([]byte, error) {
	a = indirectToStringerOrError(a)

	switch v := a.(type) {
	case int:
		return []byte(strconv.Itoa(v)), nil
	case int8:
		return []byte(strconv.FormatInt(int64(v), 10)), nil
	case int16:
		return []byte(strconv.FormatInt(int64(v), 10)), nil
	case int32:
		return []byte(strconv.Itoa(int(v))), nil
	case int64:
		return []byte(strconv.FormatInt(v, 10)), nil
	case uint:
		return []byte(strconv.FormatUint(uint64(v), 10)), nil
	case uint8:
		return []byte(strconv.FormatUint(uint64(v), 10)), nil
	case uint16:
		return []byte(strconv.FormatUint(uint64(v), 10)), nil
	case uint32:
		return []byte(strconv.FormatUint(uint64(v), 10)), nil
	case uint64:
		return []byte(strconv.FormatUint(v, 10)), nil
	case float32:
		// Use decimal to fix precision issue, FormatFloat is unstable.
		// Optional:
		//		return []byte(strconv.FormatFloat(float64(s), 'f', -1, 32)), nil
		return []byte(decimal.NewFromFloat32(v).String()), nil
	case float64:
		// Use decimal to fix precision issue, FormatFloat is unstable.
		// Optional:
		// 		return []byte(strconv.FormatFloat(s, 'f', -1, 64)), nil
		return []byte(decimal.NewFromFloat(v).String()), nil
	case *big.Int:
		return []byte(v.String()), nil
	case *big.Float:
		return []byte(v.String()), nil
	case *big.Rat:
		return []byte(v.String()), nil
	case complex64:
		return []byte(fmt.Sprintf("(%v+%vi)", real(v), imag(v))), nil
	case complex128:
		return []byte(fmt.Sprintf("(%v+%vi)", real(v), imag(v))), nil
	case bool:
		return []byte(strconv.FormatBool(v)), nil
	case string:
		return []byte(v), nil
	case []byte:
		return v, nil
	case fmt.Stringer:
		return []byte(v.String()), nil
	case error:
		return []byte(v.Error()), nil
	case nil:
		return []byte{}, nil
	default:
		return []byte{}, fmt.Errorf("unable to cast %#v of type %T to []byte", a, a)
	}
}

// ToStringer casts an interface to a fmt.Stringer type.
func ToStringer(a any) fmt.Stringer {
	v, _ := ToStringerE(a)
	return v
}

// ToStringerE casts an interface to a fmt.Stringer type.
func ToStringerE(a any) (fmt.Stringer, error) {
	a = indirectToStringerOrError(a)

	switch v := a.(type) {
	case int:
		return stringer{strconv.Itoa(v)}, nil
	case int8:
		return stringer{strconv.FormatInt(int64(v), 10)}, nil
	case int16:
		return stringer{strconv.FormatInt(int64(v), 10)}, nil
	case int32:
		return stringer{strconv.Itoa(int(v))}, nil
	case int64:
		return stringer{strconv.FormatInt(v, 10)}, nil
	case uint:
		return stringer{strconv.FormatUint(uint64(v), 10)}, nil
	case uint8:
		return stringer{strconv.FormatUint(uint64(v), 10)}, nil
	case uint16:
		return stringer{strconv.FormatUint(uint64(v), 10)}, nil
	case uint32:
		return stringer{strconv.FormatUint(uint64(v), 10)}, nil
	case uint64:
		return stringer{strconv.FormatUint(v, 10)}, nil
	case float32:
		// Use decimal to fix precision issue, FormatFloat is unstable.
		// Optional:
		//		return stringer{strconv.FormatFloat(float64(s), 'f', -1, 32)}, nil
		return stringer{decimal.NewFromFloat32(v).String()}, nil
	case float64:
		// Use decimal to fix precision issue, FormatFloat is unstable.
		// Optional:
		// 		return stringer{strconv.FormatFloat(s, 'f', -1, 64)}, nil
		return stringer{decimal.NewFromFloat(v).String()}, nil
	case *big.Int:
		return stringer{v.String()}, nil
	case *big.Float:
		return stringer{v.String()}, nil
	case *big.Rat:
		return stringer{v.String()}, nil
	case complex64:
		return stringer{fmt.Sprintf("(%v+%vi)", real(v), imag(v))}, nil
	case complex128:
		return stringer{fmt.Sprintf("(%v+%vi)", real(v), imag(v))}, nil
	case bool:
		return stringer{strconv.FormatBool(v)}, nil
	case string:
		return stringer{v}, nil
	case []byte:
		return stringer{string(v)}, nil
	case fmt.Stringer:
		return v, nil
	case error:
		return stringer{v.Error()}, nil
	case nil:
		return nil, nil
	default:
		return nil, fmt.Errorf("unable to cast %#v of type %T to fmt.Stringer", a, a)
	}
}

// ToError casts an interface to an error type.
func ToError(a any) error {
	v, _ := ToErrorE(a)
	return v
}

// ToErrorE casts an interface to an error type.
func ToErrorE(a any) (error, error) {
	a = indirectToStringerOrError(a)

	switch v := a.(type) {
	case int:
		return errors.New(strconv.Itoa(v)), nil
	case int8:
		return errors.New(strconv.FormatInt(int64(v), 10)), nil
	case int16:
		return errors.New(strconv.FormatInt(int64(v), 10)), nil
	case int32:
		return errors.New(strconv.Itoa(int(v))), nil
	case int64:
		return errors.New(strconv.FormatInt(v, 10)), nil
	case uint:
		return errors.New(strconv.FormatUint(uint64(v), 10)), nil
	case uint8:
		return errors.New(strconv.FormatUint(uint64(v), 10)), nil
	case uint16:
		return errors.New(strconv.FormatUint(uint64(v), 10)), nil
	case uint32:
		return errors.New(strconv.FormatUint(uint64(v), 10)), nil
	case uint64:
		return errors.New(strconv.FormatUint(v, 10)), nil
	case float32:
		// Use decimal to fix precision issue, FormatFloat is unstable.
		// Optional:
		//		return errors.New(strconv.FormatFloat(float64(s), 'f', -1, 32)), nil
		return errors.New(decimal.NewFromFloat32(v).String()), nil
	case float64:
		// Use decimal to fix precision issue, FormatFloat is unstable.
		// Optional:
		// 		return errors.New(strconv.FormatFloat(s, 'f', -1, 64)), nil
		return errors.New(decimal.NewFromFloat(v).String()), nil
	case *big.Int:
		return errors.New(v.String()), nil
	case *big.Float:
		return errors.New(v.String()), nil
	case *big.Rat:
		return errors.New(v.String()), nil
	case complex64:
		return fmt.Errorf("(%v+%vi)", real(v), imag(v)), nil
	case complex128:
		return fmt.Errorf(fmt.Sprintf("(%v+%vi)", real(v), imag(v))), nil
	case bool:
		return errors.New(strconv.FormatBool(v)), nil
	case string:
		return errors.New(v), nil
	case []byte:
		return errors.New(string(v)), nil
	case fmt.Stringer:
		return errors.New(v.String()), nil
	case error:
		return v, nil
	case nil:
		return nil, nil
	default:
		return nil, fmt.Errorf("unable to cast %#v of type %T to error", a, a)
	}
}

type stringer struct{ string }

func (s stringer) String() string {
	return s.string
}

// StringToBytes 字符串无拷贝转字节数组
func StringToBytes(s string) (b []byte) {
	sh := (*reflect.StringHeader)(unsafe.Pointer(&s))
	bh := (*reflect.SliceHeader)(unsafe.Pointer(&b))
	bh.Data = sh.Data
	bh.Len = sh.Len
	bh.Cap = sh.Len
	return
}

// BytesToString 字节数组无拷贝转字符串
func BytesToString(b []byte) (s string) {
	bh := (*reflect.SliceHeader)(unsafe.Pointer(&b))
	sh := (*reflect.StringHeader)(unsafe.Pointer(&s))
	sh.Data = bh.Data
	sh.Len = bh.Len
	return
}

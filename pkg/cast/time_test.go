package cast_test

import "testing"

func TestToDurationE(t *testing.T) {
	t.Error("Not implemented")
}

// func TestToDurationE(t *testing.T) {
// 	c := New(t)

// 	var td time.Duration = 5
// 	var jn json.Number
// 	_ = json.Unmarshal([]byte("5"), &jn)

// 	tests := []struct {
// 		input  any
// 		expect time.Duration
// 		iserr  bool
// 	}{
// 		{time.Duration(5), td, false},
// 		{int(5), td, false},
// 		{int64(5), td, false},
// 		{int32(5), td, false},
// 		{int16(5), td, false},
// 		{int8(5), td, false},
// 		{uint(5), td, false},
// 		{uint64(5), td, false},
// 		{uint32(5), td, false},
// 		{uint16(5), td, false},
// 		{uint8(5), td, false},
// 		{float64(5), td, false},
// 		{float32(5), td, false},
// 		{jn, td, false},
// 		{string("5"), td, false},
// 		{string("5ns"), td, false},
// 		{string("5us"), time.Microsecond * td, false},
// 		{string("5µs"), time.Microsecond * td, false},
// 		{string("5ms"), time.Millisecond * td, false},
// 		{string("5s"), time.Second * td, false},
// 		{string("5m"), time.Minute * td, false},
// 		{string("5h"), time.Hour * td, false},
// 		// errors
// 		{"test", 0, true},
// 		{testing.T{}, 0, true},
// 	}

// 	for i, test := range tests {
// 		errmsg := Commentf("i = %d", i) // assert helper message

// 		v, err := ToDurationE(test.input)
// 		if test.iserr {
// 			c.Assert(err, IsNotNil)
// 			continue
// 		}

// 		c.Assert(err, IsNil)
// 		c.Assert(v, Equals, test.expect, errmsg)

// 		// Non-E test
// 		v = ToDuration(test.input)
// 		c.Assert(v, Equals, test.expect, errmsg)
// 	}
// }

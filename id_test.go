package xid

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

const strInvalidID = "xid: invalid ID"

type IDParts struct {
	id        ID
	timestamp int64
	machine   []byte
	pid       uint16
	counter   int32
}

var IDs = []IDParts{
	IDParts{
		ID{0x4d, 0x88, 0xe1, 0x5b, 0x60, 0xf4, 0x86, 0xe4, 0x28, 0x41, 0x2d, 0xc9},
		1300816219,
		[]byte{0x60, 0xf4, 0x86},
		0xe428,
		4271561,
	},
	IDParts{
		ID{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00},
		0,
		[]byte{0x00, 0x00, 0x00},
		0x0000,
		0,
	},
	IDParts{
		ID{0x00, 0x00, 0x00, 0x00, 0xaa, 0xbb, 0xcc, 0xdd, 0xee, 0x00, 0x00, 0x01},
		0,
		[]byte{0xaa, 0xbb, 0xcc},
		0xddee,
		1,
	},
}

func TestIDPartsExtraction(t *testing.T) {
	for i, v := range IDs {
		assert.Equal(t, v.id.Time(), time.Unix(v.timestamp, 0), "#%d timestamp", i)
		assert.Equal(t, v.id.Machine(), v.machine, "#%d machine", i)
		assert.Equal(t, v.id.Pid(), v.pid, "#%d pid", i)
		assert.Equal(t, v.id.Counter(), v.counter, "#%d counter", i)
	}
}

func TestNew(t *testing.T) {
	// Generate 10 ids
	ids := make([]ID, 10)
	for i := 0; i < 10; i++ {
		ids[i] = New()
	}
	for i := 1; i < 10; i++ {
		prevID := ids[i-1]
		id := ids[i]
		// Test for uniqueness among all other 9 generated ids
		for j, tid := range ids {
			if j != i {
				assert.NotEqual(t, id, tid, "Generated ID is not unique")
			}
		}
		// Check that timestamp was incremented and is within 30 seconds of the previous one
		secs := id.Time().Sub(prevID.Time()).Seconds()
		assert.Equal(t, (secs >= 0 && secs <= 30), true, "Wrong timestamp in generated ID")
		// Check that machine ids are the same
		assert.Equal(t, id.Machine(), prevID.Machine())
		// Check that pids are the same
		assert.Equal(t, id.Pid(), prevID.Pid())
		// Test for proper increment
		delta := int(id.Counter() - prevID.Counter())
		assert.Equal(t, delta, 1, "Wrong increment in generated ID")
	}
}

func TestIDString(t *testing.T) {
	id := ID{0x4d, 0x88, 0xe1, 0x5b, 0x60, 0xf4, 0x86, 0xe4, 0x28, 0x41, 0x2d, 0xc9}
	assert.Equal(t, "9m4e2mr0ui3e8a215n4g", id.String())
}

func TestFromString(t *testing.T) {
	id, err := FromString("9m4e2mr0ui3e8a215n4g")
	assert.NoError(t, err)
	assert.Equal(t, ID{0x4d, 0x88, 0xe1, 0x5b, 0x60, 0xf4, 0x86, 0xe4, 0x28, 0x41, 0x2d, 0xc9}, id)
}

func TestFromStringInvalid(t *testing.T) {
	id, err := FromString("invalid")
	assert.EqualError(t, err, strInvalidID)
	assert.Equal(t, ID{}, id)
}

type jsonType struct {
	ID  *ID
	Str string
}

func TestIDJSONMarshaling(t *testing.T) {
	id := ID{0x4d, 0x88, 0xe1, 0x5b, 0x60, 0xf4, 0x86, 0xe4, 0x28, 0x41, 0x2d, 0xc9}
	v := jsonType{ID: &id, Str: "test"}
	data, err := json.Marshal(&v)
	assert.NoError(t, err)
	assert.Equal(t, `{"ID":"9m4e2mr0ui3e8a215n4g","Str":"test"}`, string(data))
}

func TestIDJSONUnmarshaling(t *testing.T) {
	data := []byte(`{"ID":"9m4e2mr0ui3e8a215n4g","Str":"test"}`)
	v := jsonType{}
	err := json.Unmarshal(data, &v)
	assert.NoError(t, err)
	assert.Equal(t, ID{0x4d, 0x88, 0xe1, 0x5b, 0x60, 0xf4, 0x86, 0xe4, 0x28, 0x41, 0x2d, 0xc9}, *v.ID)
}

func TestIDJSONUnmarshalingError(t *testing.T) {
	v := jsonType{}
	err := json.Unmarshal([]byte(`{"ID":"9M4E2MR0UI3E8A215N4G"}`), &v)
	assert.EqualError(t, err, strInvalidID)
	err = json.Unmarshal([]byte(`{"ID":"TYjhW2D0huQoQS"}`), &v)
	assert.EqualError(t, err, strInvalidID)
	err = json.Unmarshal([]byte(`{"ID":"TYjhW2D0huQoQS3kdk"}`), &v)
	assert.EqualError(t, err, strInvalidID)
}

func TestIDDriverValue(t *testing.T) {
	id := ID{0x4d, 0x88, 0xe1, 0x5b, 0x60, 0xf4, 0x86, 0xe4, 0x28, 0x41, 0x2d, 0xc9}
	data, err := id.Value()
	assert.NoError(t, err)
	assert.Equal(t, "9m4e2mr0ui3e8a215n4g", data)
}

func TestIDDriverScan(t *testing.T) {
	id := ID{}
	err := id.Scan("9m4e2mr0ui3e8a215n4g")
	assert.NoError(t, err)
	assert.Equal(t, ID{0x4d, 0x88, 0xe1, 0x5b, 0x60, 0xf4, 0x86, 0xe4, 0x28, 0x41, 0x2d, 0xc9}, id)
}

func TestIDDriverScanError(t *testing.T) {
	id := ID{}
	err := id.Scan(0)
	assert.EqualError(t, err, "xid: scanning unsupported type: int")
	err = id.Scan("0")
	assert.EqualError(t, err, strInvalidID)
}

func TestIDDriverScanByteFromDatabase(t *testing.T) {
	id := ID{}
	bs := []byte("9m4e2mr0ui3e8a215n4g")
	err := id.Scan(bs)
	assert.NoError(t, err)
	assert.Equal(t, ID{0x4d, 0x88, 0xe1, 0x5b, 0x60, 0xf4, 0x86, 0xe4, 0x28, 0x41, 0x2d, 0xc9}, id)
}

func BenchmarkNew(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = New()
		}
	})
}

func BenchmarkNewString(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = New().String()
		}
	})
}

func BenchmarkFromString(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_, _ = FromString("9m4e2mr0ui3e8a215n4g")
		}
	})
}

// func BenchmarkUUIDv1(b *testing.B) {
// 	b.RunParallel(func(pb *testing.PB) {
// 		for pb.Next() {
// 			_ = uuid.NewV1().String()
// 		}
// 	})
// }

// func BenchmarkUUIDv4(b *testing.B) {
// 	b.RunParallel(func(pb *testing.PB) {
// 		for pb.Next() {
// 			_ = uuid.NewV4().String()
// 		}
// 	})
// }

func TestID_IsNil(t *testing.T) {
	tests := []struct {
		name string
		id   ID
		want bool
	}{
		{
			name: "ID not nil",
			id:   New(),
			want: false,
		},
		{
			name: "Nil ID",
			id:   ID{},
			want: true,
		},
	}
	for _, tt := range tests {
		assert.Equal(t, tt.id.IsNil(), tt.want)
	}
}

func TestNilID(t *testing.T) {
	var id ID
	nilid := NilID()
	assert.Equal(t, id, nilid)
}

func TestNilID_IsNil(t *testing.T) {
	assert.True(t, NilID().IsNil())
}

func TestID_Bytes(t *testing.T) {
	id := New()
	underlying := [rawLen]byte(id)
	b := id.Bytes()
	for i := range underlying {
		assert.Equal(t, underlying[i], b[i])
	}
}

func TestFromBytes_Invariant(t *testing.T) {
	id := New()
	b, err := FromBytes(id.Bytes())
	assert.NoError(t, err)
	assert.Equal(t, b, id)
}

func TestFromBytes_InvalidBytes(t *testing.T) {
	cases := []struct{
		length int; shouldFail bool} {
		{11, true},
		{12, false},
		{13, true},
	}
	for _, c := range cases {
		b := make([]byte, c.length, c.length)
		_, err := FromBytes(b)
		if c.shouldFail {
			assert.Error(t, err, "Length %d should fail.", c.length)
		} else {
			assert.NoError(t, err, "Length %d should not fail.", c.length)
		}
	}
}

func TestID_Compare(t *testing.T) {
	pairs := []struct{
		left ID
		right ID
		expected int
	} {
		{IDs[1].id, IDs[0].id, -1},
		{ID{0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}, IDs[2].id, -1},
		{IDs[0].id, IDs[0].id, 0},
	}
	for _, p := range pairs {
		assert.Equal(
			t, p.expected, p.left.Compare(p.right),
			"%s Compare to %s should return %d", p.left, p.right, p.expected,
		)
		assert.Equal(
			t, -1 * p.expected, p.right.Compare(p.left),
			"%s Compare to %s should return %d", p.right, p.left, - 1 * p.expected,
		)
	}
}

var IDList = []ID{IDs[0].id, IDs[1].id, IDs[2].id}


func TestSorter_Len(t *testing.T) {
	assert.Equal(t, 0, sorter([]ID{}).Len())
	assert.Equal(t, 3, sorter(IDList).Len())
}


func TestSorter_Less(t *testing.T) {
	sorter := sorter(IDList)
	assert.True(t, sorter.Less(1, 0))
	assert.False(t, sorter.Less(2, 1))
	assert.False(t, sorter.Less(0, 0))
}

func TestSorter_Swap(t *testing.T) {
	ids := make([]ID, 0)
	for _, id := range IDList {
		ids = append(ids, id)
	}
	sorter := sorter(ids)
	sorter.Swap(0, 1)
	assert.Equal(t, ids[0], IDList[1])
	assert.Equal(t, ids[1], IDList[0])
	sorter.Swap(2, 2)
	assert.Equal(t, ids[2], IDList[2])
}

func TestSort(t *testing.T) {
	ids := make([]ID, 0)
	for _, id := range IDList {
		ids = append(ids, id)
	}
	Sort(ids)
	assert.Equal(t, ids, []ID{IDList[1], IDList[2], IDList[0]})
}

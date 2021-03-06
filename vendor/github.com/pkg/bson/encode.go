package bson

import (
	"errors"
	"math"
	"reflect"
	"strconv"
)

// encode encodes v according to the rules of Marshal into a BSON document.
func encode(v interface{}) ([]byte, error) {
	rv := reflect.ValueOf(v)
	if rv.IsNil() {
		return nil, errors.New("bson: error calling MarshalJSON for type " + rv.Type().String() + ": was nil")
	}
	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}
	var w writer
	switch rv.Kind() {
	case reflect.Map:
		_, err := w.writeMap(rv)
		return w.bson, err
	}
	return nil, errors.New("bson: error calling MarshalJSON for type " + rv.Type().String() + ": unsupported")
}

// writer writes formatted BSON objects.
type writer struct {
	bson []byte
}

// writeMap encodes the contents of a map[string]interface{} as a BSON
// document.
func (w *writer) writeMap(v reflect.Value) (int, error) {
	off := len(w.bson)                  // the location of our header
	w.bson = append(w.bson, 0, 0, 0, 0) // document header
	count := sizeofInt32 + 1            // header plus trailing 0x0
	keys := v.MapKeys()
	for _, k := range keys {
		v := v.MapIndex(k)
		n, err := w.writeValue(k.String(), v)
		if err != nil {
			return 0, err
		}
		count += n
	}
	w.bson = append(w.bson, 0) // document trailer
	// update document header
	w.bson[off] = byte(count)
	w.bson[off+1] = byte(count >> 8)
	w.bson[off+2] = byte(count >> 16)
	w.bson[off+3] = byte(count >> 24)
	return count, nil
}

func (w *writer) writeValue(ename string, v reflect.Value) (int, error) {
	var count int
	if v.IsNil() {
		count += w.writeType(0x0a)
		count += w.writeCstring(ename)
		return count, nil
	}
	switch vv := v.Interface().(type) {
	case ObjectId:
		count += w.writeType(0x07)
		count += w.writeCstring(ename)
		count += w.writeBytes(vv[:])
	case Datetime:
		count += w.writeType(0x09)
		count += w.writeCstring(ename)
		count += w.writeInt64(int64(vv))
	case Timestamp:
		count += w.writeType(0x11)
		count += w.writeCstring(ename)
		count += w.writeInt64(int64(vv))
	default:
		switch v := v.Elem(); v.Kind() {
		case reflect.Float64:
			count += w.writeType(0x01)
			count += w.writeCstring(ename)
			count += w.writeFloat64(v.Float())
		case reflect.String:
			count += w.writeType(0x02)
			count += w.writeCstring(ename)
			s := v.String()
			sz := len(s) + 1
			count += w.writeInt32(int32(sz))
			w.bson = append(w.bson, s...)
			w.bson = append(w.bson, 0)
			count += sz
		case reflect.Bool:
			count += w.writeType(0x08)
			count += w.writeCstring(ename)
			count += w.writeBool(v.Bool())
		case reflect.Int32:
			count += w.writeType(0x10)
			count += w.writeCstring(ename)
			count += w.writeInt32(int32(v.Int()))
		case reflect.Int64:
			count += w.writeType(0x12)
			count += w.writeCstring(ename)
			count += w.writeInt64(int64(v.Int()))
		case reflect.Slice:
			// slices encoded as arrays
			count += w.writeType(0x04)
			count += w.writeCstring(ename)
			n, err := w.writeSlice(v)
			if err != nil {
				return 0, err
			}
			count += n
		case reflect.Map:
			// maps encoded as documents
			count += w.writeType(0x03)
			count += w.writeCstring(ename)
			n, err := w.writeMap(v)
			if err != nil {
				return 0, err
			}
			count += n
		default:
			return 0, errors.New("bson: unsupported type: " + v.Type().String())
		}
	}
	return count, nil
}

func (w *writer) writeSlice(v reflect.Value) (int, error) {
	off := len(w.bson)                  // the location of our header
	w.bson = append(w.bson, 0, 0, 0, 0) // document header
	count := sizeofInt32 + 1            // header plus trailing 0x0
	for i, n := 0, v.Len(); i < n; i++ {
		v := v.Index(i)
		n, err := w.writeValue(strconv.Itoa(i), v)
		if err != nil {
			return 0, err
		}
		count += n
	}
	w.bson = append(w.bson, 0) // document trailer
	// update document header
	w.bson[off] = byte(count)
	w.bson[off+1] = byte(count >> 8)
	w.bson[off+2] = byte(count >> 16)
	w.bson[off+3] = byte(count >> 24)
	return count, nil
}

func (w *writer) writeType(typ byte) int {
	w.bson = append(w.bson, typ)
	return 1
}

func (w *writer) writeBool(b bool) int {
	if b {
		w.bson = append(w.bson, 0x01)
	} else {
		w.bson = append(w.bson, 0x00)
	}
	return 1
}

func (w *writer) writeBytes(b []byte) int {
	w.bson = append(w.bson, b...)
	return len(b)
}

func (w *writer) writeCstring(s string) int {
	w.bson = append(w.bson, s...)
	w.bson = append(w.bson, 0)
	return len(s) + 1
}

func (w *writer) writeInt32(v int32) int {
	w.bson = append(w.bson, byte(v), byte(v>>8), byte(v>>16), byte(v>>24))
	return sizeofInt32
}

func (w *writer) writeInt64(v int64) int {
	w.bson = append(w.bson, byte(v), byte(v>>8), byte(v>>16), byte(v>>24),
		byte(v>>32), byte(v>>40), byte(v>>48), byte(v>>56))
	return sizeofInt64
}

func (w *writer) writeFloat64(f float64) int {
	v := math.Float64bits(f)
	w.bson = append(w.bson, byte(v), byte(v>>8), byte(v>>16), byte(v>>24),
		byte(v>>32), byte(v>>40), byte(v>>48), byte(v>>56))
	return sizeofInt64
}

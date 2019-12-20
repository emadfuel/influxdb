// Generated by tmpl
// https://github.com/benbjohnson/tmpl
//
// DO NOT EDIT!
// Source: arrays.gen.go.tmpl

package gen

import (
	"github.com/influxdata/influxdb/v2/tsdb"
	"github.com/influxdata/influxdb/v2/tsdb/tsm1"
)

type FloatValues interface {
	Copy(*tsdb.FloatArray)
}

type floatArray struct {
	tsdb.FloatArray
}

func newFloatArrayLen(sz int) *floatArray {
	return &floatArray{
		FloatArray: tsdb.FloatArray{
			Timestamps: make([]int64, sz),
			Values:     make([]float64, sz),
		},
	}
}

func (a *floatArray) Encode(b []byte) ([]byte, error) {
	return tsm1.EncodeFloatArrayBlock(&a.FloatArray, b)
}

func (a *floatArray) Copy(dst *tsdb.FloatArray) {
	dst.Timestamps = append(dst.Timestamps[:0], a.Timestamps...)
	dst.Values = append(dst.Values[:0], a.Values...)
}

type IntegerValues interface {
	Copy(*tsdb.IntegerArray)
}

type integerArray struct {
	tsdb.IntegerArray
}

func newIntegerArrayLen(sz int) *integerArray {
	return &integerArray{
		IntegerArray: tsdb.IntegerArray{
			Timestamps: make([]int64, sz),
			Values:     make([]int64, sz),
		},
	}
}

func (a *integerArray) Encode(b []byte) ([]byte, error) {
	return tsm1.EncodeIntegerArrayBlock(&a.IntegerArray, b)
}

func (a *integerArray) Copy(dst *tsdb.IntegerArray) {
	dst.Timestamps = append(dst.Timestamps[:0], a.Timestamps...)
	dst.Values = append(dst.Values[:0], a.Values...)
}

type UnsignedValues interface {
	Copy(*tsdb.UnsignedArray)
}

type unsignedArray struct {
	tsdb.UnsignedArray
}

func newUnsignedArrayLen(sz int) *unsignedArray {
	return &unsignedArray{
		UnsignedArray: tsdb.UnsignedArray{
			Timestamps: make([]int64, sz),
			Values:     make([]uint64, sz),
		},
	}
}

func (a *unsignedArray) Encode(b []byte) ([]byte, error) {
	return tsm1.EncodeUnsignedArrayBlock(&a.UnsignedArray, b)
}

func (a *unsignedArray) Copy(dst *tsdb.UnsignedArray) {
	dst.Timestamps = append(dst.Timestamps[:0], a.Timestamps...)
	dst.Values = append(dst.Values[:0], a.Values...)
}

type StringValues interface {
	Copy(*tsdb.StringArray)
}

type stringArray struct {
	tsdb.StringArray
}

func newStringArrayLen(sz int) *stringArray {
	return &stringArray{
		StringArray: tsdb.StringArray{
			Timestamps: make([]int64, sz),
			Values:     make([]string, sz),
		},
	}
}

func (a *stringArray) Encode(b []byte) ([]byte, error) {
	return tsm1.EncodeStringArrayBlock(&a.StringArray, b)
}

func (a *stringArray) Copy(dst *tsdb.StringArray) {
	dst.Timestamps = append(dst.Timestamps[:0], a.Timestamps...)
	dst.Values = append(dst.Values[:0], a.Values...)
}

type BooleanValues interface {
	Copy(*tsdb.BooleanArray)
}

type booleanArray struct {
	tsdb.BooleanArray
}

func newBooleanArrayLen(sz int) *booleanArray {
	return &booleanArray{
		BooleanArray: tsdb.BooleanArray{
			Timestamps: make([]int64, sz),
			Values:     make([]bool, sz),
		},
	}
}

func (a *booleanArray) Encode(b []byte) ([]byte, error) {
	return tsm1.EncodeBooleanArrayBlock(&a.BooleanArray, b)
}

func (a *booleanArray) Copy(dst *tsdb.BooleanArray) {
	dst.Timestamps = append(dst.Timestamps[:0], a.Timestamps...)
	dst.Values = append(dst.Values[:0], a.Values...)
}

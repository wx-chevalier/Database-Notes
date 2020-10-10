package getters

import (
	"fmt"
	"math/rand"
)

type RandomSample struct {
	name      string
	samples   []interface{}
	allowNull bool
}

func (r *RandomSample) Value() interface{} {
	if r.allowNull && rand.Int63n(100) < nilFrequency {
		return nil
	}
	pos := rand.Int63n(int64(len(r.samples)))
	return r.samples[pos]
}

func (r *RandomSample) String() string {
	v := r.Value()
	if v == nil {
		return NULL
	}
	return fmt.Sprintf("%v", v)
}

func (r *RandomSample) Quote() string {
	v := r.Value()
	if v == nil {
		return NULL
	}
	switch v.(type) {
	case string:
		return fmt.Sprintf("%q", v)
	default:
		return fmt.Sprintf("%v", v)
	}
}

func NewRandomSample(name string, samples []interface{}, allowNull bool) *RandomSample {
	r := &RandomSample{name, samples, allowNull}
	return r
}

package getters

import (
	"fmt"
	"math/rand"
)

type RandomInt struct {
	name      string
	mask      int64
	allowNull bool
}

func (r *RandomInt) Value() interface{} {
	return rand.Int63n(r.mask)
}

func (r *RandomInt) String() string {
	return fmt.Sprintf("%d", r.Value())
}

func (r *RandomInt) Quote() string {
	return r.String()
}

func NewRandomInt(name string, mask int64, allowNull bool) *RandomInt {
	return &RandomInt{name, mask, allowNull}
}

type RandomIntRange struct {
	name      string
	min       int64
	max       int64
	allowNull bool
}

func (r *RandomIntRange) Value() interface{} {
	limit := r.max - r.min + 1
	return r.min + rand.Int63n(limit)
}

func (r *RandomIntRange) String() string {
	return fmt.Sprintf("%d", r.Value())
}

func (r *RandomIntRange) Quote() string {
	return r.String()
}

func NewRandomIntRange(name string, min, max int64, allowNull bool) *RandomIntRange {
	return &RandomIntRange{name, min, max, allowNull}
}

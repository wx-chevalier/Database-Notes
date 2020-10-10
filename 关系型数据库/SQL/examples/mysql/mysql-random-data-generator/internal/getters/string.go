package getters

import (
	"fmt"
	"math/rand"

	"github.com/icrowley/fake"
)

// RandomString getter
type RandomString struct {
	name      string
	maxSize   int64
	allowNull bool
}

func (r *RandomString) Value() interface{} {
	if r.allowNull && rand.Int63n(100) < nilFrequency {
		return nil
	}
	var s string
	maxSize := uint64(r.maxSize)
	if maxSize == 0 {
		maxSize = uint64(rand.Int63n(100))
	}

	if maxSize <= 10 {
		s = fake.FirstName()
	} else if maxSize < 30 {
		s = fake.FullName()
	} else {
		s = fake.Sentence()
	}
	if len(s) > int(maxSize) {
		s = s[:int(maxSize)]
	}
	return s
}

func (r *RandomString) String() string {
	v := r.Value()
	if v == nil {
		return NULL
	}
	return v.(string)
}

func (r *RandomString) Quote() string {
	v := r.Value()
	if v == nil {
		return NULL
	}
	return fmt.Sprintf("%q", v)
}

func NewRandomString(name string, maxSize int64, allowNull bool) *RandomString {
	return &RandomString{name, maxSize, allowNull}
}

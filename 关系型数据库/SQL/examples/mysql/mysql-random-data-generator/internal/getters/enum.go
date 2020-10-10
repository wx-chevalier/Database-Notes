package getters

import (
	"fmt"
	"math/rand"
)

// RandomEnum Getter
type RandomEnum struct {
	allowedValues []string
	allowNull     bool
}

func (r *RandomEnum) Value() interface{} {
	if r.allowNull && rand.Int63n(100) < nilFrequency {
		return nil
	}
	i := rand.Int63n(int64(len(r.allowedValues)))
	return r.allowedValues[i]
}

func (r *RandomEnum) String() string {
	if v := r.Value(); v != nil {
		return v.(string)
	}
	return "NULL"
}

func (r *RandomEnum) Quote() string {
	if v := r.Value(); v != nil {
		return fmt.Sprintf("%q", v)
	}
	return "NULL"
}

func NewRandomEnum(allowedValues []string, allowNull bool) *RandomEnum {
	return &RandomEnum{allowedValues, allowNull}
}

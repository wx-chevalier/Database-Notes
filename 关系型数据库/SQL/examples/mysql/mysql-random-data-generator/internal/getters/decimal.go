package getters

import (
	"fmt"
	"math"
	"math/rand"
)

// RandomDecimal holds unexported data for decimal values
type RandomDecimal struct {
	name      string
	size      int64
	allowNull bool
}

func (r *RandomDecimal) Value() interface{} {
	f := rand.Float64() * float64(rand.Int63n(int64(math.Pow10(int(r.size)))))
	return f
}

func (r *RandomDecimal) String() string {
	return fmt.Sprintf("%0f", r.Value())
}

func (r *RandomDecimal) Quote() string {
	return r.String()
}

func NewRandomDecimal(name string, size int64, allowNull bool) *RandomDecimal {
	return &RandomDecimal{name, size, allowNull}
}

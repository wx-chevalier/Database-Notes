package getters

func NewRandomYear(name string, format int, allowNull bool) *RandomIntRange {
	if format == 2 {
		return &RandomIntRange{name, 01, 99, allowNull}
	}
	return &RandomIntRange{name, 1901, 2155, allowNull}
}

func NewRandomYearRange(name string, min, max int64, allowNull bool) *RandomIntRange {
	return &RandomIntRange{name, min, max, allowNull}
}

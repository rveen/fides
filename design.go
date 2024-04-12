package fides

type Design struct {
	Components []*Component
	Mission    *Mission
}

func NewDesign() *Design {
	return &Design{}
}

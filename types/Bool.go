package types

type Bool struct {
	ScaleDecoder
}

func (b *Bool) Process() {
	b.Value = b.getNextBool()
}

func (b *Bool) Encode(value interface{}) string {
	v, ok := value.(bool)
	if !ok {
		panic("invalid bool input")
	}
	if v {
		return "01"
	}
	return "00"
}

func (b *Bool) TypeStructString() string {
	return "Bool"
}

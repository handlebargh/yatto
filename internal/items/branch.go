package items

type Branch struct {
	Name     string
	Upstream string
}

func (b Branch) Title() string       { return b.Name }
func (b Branch) Description() string { return b.Upstream }
func (b Branch) FilterValue() string { return b.Name }

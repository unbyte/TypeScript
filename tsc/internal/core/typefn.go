package core

func (o *CompilerOptions) GetTypefnVersion() int {
	if o == nil || o.Typefn == nil {
		return 0
	}
	return o.Typefn.Version
}

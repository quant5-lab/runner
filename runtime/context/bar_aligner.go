package context

type BarAligner interface {
	AlignToParent(childBarIdx int) int
	AlignToChild(parentBarIdx int) int
}

type IdentityAligner struct{}

func NewIdentityAligner() *IdentityAligner {
	return &IdentityAligner{}
}

func (a *IdentityAligner) AlignToParent(childBarIdx int) int {
	return childBarIdx
}

func (a *IdentityAligner) AlignToChild(parentBarIdx int) int {
	return parentBarIdx
}

type MappedAligner struct {
	childToParentMap map[int]int
	parentToChildMap map[int]int
}

func NewMappedAligner() *MappedAligner {
	return &MappedAligner{
		childToParentMap: make(map[int]int),
		parentToChildMap: make(map[int]int),
	}
}

func (a *MappedAligner) SetMapping(childIdx, parentIdx int) {
	a.childToParentMap[childIdx] = parentIdx
	a.parentToChildMap[parentIdx] = childIdx
}

func (a *MappedAligner) AlignToParent(childBarIdx int) int {
	if parentIdx, found := a.childToParentMap[childBarIdx]; found {
		return parentIdx
	}
	return -1
}

func (a *MappedAligner) AlignToChild(parentBarIdx int) int {
	if childIdx, found := a.parentToChildMap[parentBarIdx]; found {
		return childIdx
	}
	return -1
}

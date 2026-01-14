package source_identity

type SourceIdentifier struct {
	hash string
}

func NewSourceIdentifier(hash string) SourceIdentifier {
	return SourceIdentifier{hash: hash}
}

func (s SourceIdentifier) Hash() string {
	return s.hash
}

func (s SourceIdentifier) IsEmpty() bool {
	return s.hash == ""
}

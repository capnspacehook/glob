package match

type NonEmpty struct {
	Matcher
}

func NewNonEmpty(m Matcher) NonEmpty {
	return NonEmpty{m}
}

func (n NonEmpty) Match(s string) bool {
	if s == "" {
		return false
	}
	return n.Matcher.Match(s)
}

func (n NonEmpty) Index(s string) (int, []int) {
	return n.Matcher.Index(s)
}

func (n NonEmpty) Len() int {
	return lenZero
}

func (n NonEmpty) String() string {
	return "<non_empty>"
}

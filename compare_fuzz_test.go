package glob

import (
	"strings"
	"testing"

	"github.com/bmatcuk/doublestar/v4"
	"pgregory.net/rapid"
)

// sep is the separator shared by both libraries. doublestar.Match always uses
// '/', so gobwas is compiled with the same separator for a fair comparison.
const sep = '/'

var (
	// classMembers never include the separator (see the header note on classes).
	classMembers = []byte{'a', 'b', 'c', 'd'}

	tokens []byte = []byte{'*', '?', '[', ']', '{', '}', '-', '\\'}
	// alphabet is used for literal pattern characters and target characters. It
	// is intentionally small (so matches actually happen) and includes '/'.
	alphabet = append(classMembers, '/')

	literalGen = rapid.SampledFrom(alphabet)
	tokenGen   = rapid.SampledFrom(tokens)
	memberGen  = rapid.SampledFrom(classMembers)

	// targetGen draws a 0..11 char string over the alphabet, separator included.
	targetGen = rapid.StringOfN(rapid.SampledFrom([]rune{'a', 'b', 'c', 'd', '/'}), 1, 16, -1)

	// patternGen draws a pattern in the gobwas/doublestar common subset.
	patternGen = rapid.Custom(genPattern)
)

func FuzzGlobVsDoublestar(f *testing.F) {
	f.Fuzz(rapid.MakeFuzz(compareGlobs))
}

func TestGlobVsDoublestar(t *testing.T) {
	rapid.Check(t, compareGlobs)
}

// genPattern builds a pattern from literals, '*' (never '**'), '?', positive
// '[abc]' classes, and '{a,bc}' brace alternations. The lastStar guard keeps
// two '*' tokens from ever landing adjacent.
func genPattern(t *rapid.T) string {
	patternLen := rapid.IntRange(1, 16).Draw(t, "patternLen")
	var sb strings.Builder
	var lastStar bool

	var writePattern func(n int, inAnyOf bool, anyOfLen int)
	writePattern = func(n int, inAnyOf bool, anyOfLen int) {
		anyOfN := anyOfLen

		// prevent a star from appearing in braces following another star,
		// a star apprearing in braces followed by another star or any
		// combination thereof; ex '*{*}', '*{?,*}, '{*,?}*'. doublestar
		// doesn't handle this correctly.
		notLastStar := func() bool {
			if inAnyOf {
				return lastStar
			}
			return false
		}

		for i := range n {
			if inAnyOf {
				if anyOfN > 0 {
					if anyOfLen != anyOfN {
						sb.WriteByte(',')
					}
					anyOfN--
				} else {
					inAnyOf = false
					sb.WriteByte('}')
				}
			}

			maxChoice := 8
			if inAnyOf {
				maxChoice = 7
			}

			switch rapid.IntRange(0, maxChoice).Draw(t, "choice") {
			case 0, 1, 2, 3: // literal (weighted so matches actually happen)
				sb.WriteByte(literalGen.Draw(t, "literal"))
				lastStar = notLastStar()
			case 4:
				sb.WriteByte('\\')
				sb.WriteByte(tokenGen.Draw(t, "escapedLiteral"))
			case 5: // single '*', never adjacent to another '*'
				if lastStar {
					sb.WriteByte(literalGen.Draw(t, "literal"))
					lastStar = notLastStar()
				} else {
					sb.WriteByte('*')
					lastStar = true
				}
			case 6: // '?'
				sb.WriteByte('?')
				lastStar = notLastStar()
			case 7: // list '[ ]'
				listLen := rapid.IntRange(0, 4).Draw(t, "listLen")
				sb.WriteByte('[')
				// negate 50% of the time
				// var negated bool
				// TODO: fix or comment out
				if rapid.Bool().Draw(t, "negated") {
					// negated = true
					// sb.WriteByte('!')
				}

				for range listLen {
					// range 33% of the time
					if rapid.IntRange(0, 2).Draw(t, "isRange") == 0 {
						// don't use an escaped token as the low end of the
						// range as it could cause '/' to be included in the
						// range which doublestar doesn't always handle correctly
						sb.WriteString(genChar(t, false))
						sb.WriteByte('-')
						sb.WriteString(genChar(t, true))
					} else {
						sb.WriteString(genChar(t, true))
					}
				}

				sb.WriteByte(']')
				lastStar = notLastStar()
			case 8: // any of '{ }'
				// don't generate '{}' as doublestar doesn't always handle
				// it correctly
				if i == n-1 {
					continue
				}

				sb.WriteByte('{')
				writePattern(n-i-1, true, rapid.IntRange(1, 3).Draw(t, "anyOfLen"))
				// don't set lastStar to false; if there was a star before,
				// then doublestar will incorrectly handle patterns like '*{*}'
			}
		}

		if inAnyOf {
			sb.WriteByte('}')
		}
	}

	writePattern(patternLen, false, 0)

	return sb.String()
}

func genChar(t *rapid.T, tokenPossible bool) string {
	// 33% of the time pick an escaped token
	if tokenPossible && rapid.IntRange(0, 2).Draw(t, "escapeToken") == 0 {
		return "\\" + string(tokenGen.Draw(t, "token"))
	}

	literal := string(memberGen.Draw(t, "literal"))
	if rapid.IntRange(0, 2).Draw(t, "escapeLiteral") == 0 {
		return "\\" + literal
	}
	return literal
}

func compareGlobs(t *rapid.T) {
	pattern := patternGen.Draw(t, "pattern")
	target := targetGen.Draw(t, "target")

	g, err := Compile(pattern, sep)
	if err != nil {
		return
	}

	gm := g.Match(target)
	dm, err := doublestar.Match(pattern, target)
	if err != nil {
		return
	}

	if gm != dm {
		t.Fatalf(`
match disagreement:
  pattern    = %s
  target     = %s
  gobwas     = %v
  doublestar = %v`,
			pattern, target, gm, dm)
	}
}

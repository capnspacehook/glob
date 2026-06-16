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
	targetGen = rapid.StringOfN(rapid.SampledFrom([]rune{'a', 'b', 'c', 'd', '/'}), 0, 16, -1)

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
	n := rapid.IntRange(1, 16).Draw(t, "")
	var sb strings.Builder
	var lastStar bool
	var inAnyOf bool
	var anyOfN int

	for range n {
		if inAnyOf {
			anyOfN--
			sb.WriteByte(',')
			if anyOfN == 0 {
				inAnyOf = false
				sb.WriteByte('}')
				lastStar = false
			}
		}

		switch rapid.IntRange(0, 8).Draw(t, "") {
		case 0, 1, 2, 3: // literal (weighted so matches actually happen)
			sb.WriteByte(literalGen.Draw(t, ""))
			lastStar = false
		case 4:
			sb.WriteByte('\\')
			sb.WriteByte(tokenGen.Draw(t, ""))
		case 5: // single '*', never adjacent to another '*'
			if lastStar {
				sb.WriteByte(literalGen.Draw(t, ""))
				lastStar = false
			} else {
				sb.WriteByte('*')
				lastStar = true
			}
		case 6: // '?'
			sb.WriteByte('?')
			lastStar = false
		case 7: // character class '[ ]'
			classLen := rapid.IntRange(0, 4).Draw(t, "")
			sb.WriteByte('[')
			// negate 50% of the time
			// if rapid.Bool().Draw(t, "") {
			// 	sb.WriteByte('!')
			// }

			// range 50% of the time
			if rapid.Bool().Draw(t, "") {
				genChar(t, &sb)
				sb.WriteByte('-')
				genChar(t, &sb)
			} else {
				for range classLen {
					genChar(t, &sb)
				}
			}

			sb.WriteByte(']')
			lastStar = false
		case 8: // brace alternation '{ }'
			inAnyOf = true
			anyOfN = rapid.IntRange(0, 3).Draw(t, "") + 1
			sb.WriteByte('{')
			lastStar = false
		}
	}

	if inAnyOf {
		sb.WriteByte(',')
		sb.WriteByte('}')
	}

	return sb.String()
}

func genChar(t *rapid.T, sb *strings.Builder) {
	// 33% of the time pick an escaped token
	if rapid.IntRange(0, 3).Draw(t, "") == 0 {
		sb.WriteByte('\\')
		sb.WriteByte(tokenGen.Draw(t, ""))
	} else {
		if rapid.IntRange(0, 2).Draw(t, "") == 0 {
			sb.WriteByte('\\')
		}
		sb.WriteByte(memberGen.Draw(t, ""))
	}
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

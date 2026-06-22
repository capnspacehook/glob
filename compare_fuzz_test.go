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
	classMembers = []rune{'a', 'b', 'c', 'd', '\u00e9', '\u672c', '\uff0f', '\uff1f', '\U0001f600'}

	tokens = []rune{'*', '?', '[', ']', '{', '}', '-', '\\'}
	// alphabet is used for literal pattern characters and target characters. It
	// is intentionally small (so matches actually happen) and includes '/'.
	alphabet = append(classMembers, '/')

	literalGen = rapid.SampledFrom(alphabet)
	tokenGen   = rapid.SampledFrom(tokens)
	memberGen  = rapid.SampledFrom(classMembers)

	targetGen      = rapid.StringOfN(rapid.SampledFrom(append(classMembers, '/')), 1, 16, -1)
	targetNoSepGen = rapid.StringOfN(rapid.SampledFrom(classMembers), 1, 16, -1)
)

func getPattern(t *rapid.T) (string, bool) {
	patternLen := rapid.IntRange(1, 16).Draw(t, "patternLen")
	var sb strings.Builder
	var lastStar bool
	var negatedCharClass bool

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

		for step := 0; step < n; step++ {
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
			// don't generate '{}' as doublestar doesn't always handle
			// it correctly
			if step == n-1 {
				maxChoice = 7
			}

			switch rapid.IntRange(0, maxChoice).Draw(t, "choice") {
			case 0, 1, 2, 3: // literal (weighted so matches actually happen)
				sb.WriteRune(literalGen.Draw(t, "literal"))
				lastStar = notLastStar()
			case 4:
				sb.WriteByte('\\')
				sb.WriteRune(tokenGen.Draw(t, "escapedLiteral"))
			case 5: // single '*', never adjacent to another '*'
				if lastStar {
					sb.WriteRune(literalGen.Draw(t, "literal"))
					lastStar = notLastStar()
				} else {
					sb.WriteByte('*')
					lastStar = true
				}
			case 6: // '?'
				sb.WriteByte('?')
				lastStar = notLastStar()
			case 7: // list '[ ]'
				listLen := rapid.IntRange(1, 4).Draw(t, "listLen")
				sb.WriteByte('[')
				// negate 50% of the time
				if !lastStar && rapid.Bool().Draw(t, "negated") {
					negatedCharClass = true
					sb.WriteByte('!')
				}

				for range listLen {
					// range 33% of the time
					if rapid.IntRange(0, 2).Draw(t, "isRange") == 0 {
						low := memberGen.Filter(func(r rune) bool {
							return r != classMembers[len(classMembers)-1]
						}).Draw(t, "low")
						sb.WriteRune(low)
						sb.WriteByte('-')
						high := memberGen.Filter(func(r rune) bool {
							return r > low
						}).Draw(t, "high")
						sb.WriteRune(high)
					} else {
						sb.WriteString(genChar(t, true))
					}
				}

				sb.WriteByte(']')
				lastStar = notLastStar()
			case 8: // any of '{ }'
				sb.WriteByte('{')

				al := rapid.IntRange(1, 3).Draw(t, "anyOfLen")
				stepsLeft := n - step - 1
				al = min(al, stepsLeft)

				writePattern(al, true, al)
				step += al
				// don't set lastStar to false; if there was a star before,
				// then doublestar will incorrectly handle patterns like '*{*}'
			}
		}

		if inAnyOf {
			sb.WriteByte('}')
		}
	}

	writePattern(patternLen, false, 0)

	return sb.String(), negatedCharClass
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

func TestGlobVsDoublestar(t *testing.T) {
	rapid.Check(t, compareGlobs)
}

func FuzzGlobVsDoublestar(f *testing.F) {
	f.Fuzz(rapid.MakeFuzz(compareGlobs))
}

func compareGlobs(t *rapid.T) {
	pattern, negatedCharClass := getPattern(t)

	// avoid a separator in the target if the pattern has a negated character class
	// doublestar doesn't handle this correctly
	var target string
	if negatedCharClass {
		target = targetNoSepGen.Draw(t, "target")
	} else {
		target = targetGen.Draw(t, "target")
	}

	g, err := Compile(pattern, sep)
	if err != nil {
		t.Fatalf("compile error: %v", err)
	}

	gm := g.Match(target)
	dm, err := doublestar.Match(pattern, target)
	if err != nil {
		t.Fatalf("glob accepted what doublestar did not: %v", err)
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

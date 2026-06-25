package glob

import (
	"slices"
	"strings"
	"testing"
	"unicode/utf8"

	"pgregory.net/rapid"

	"github.com/capnspacehook/glob/compiler"
	"github.com/capnspacehook/glob/syntax"
	"github.com/capnspacehook/glob/syntax/ast"
)

func TestSimpleMatchSanity(t *testing.T) {
	for _, tt := range tests {
		t.Run(tt.pattern, func(t *testing.T) {
			tree, err := syntax.Parse(tt.pattern)
			if err != nil {
				t.Fatalf("parse %q: %v", tt.pattern, err)
			}
			if got := simpleMatch(tree, tt.match, tt.delimiters); got != tt.should {
				t.Errorf(
					"pattern %q matching %q should be %v but got %v",
					tt.pattern, tt.match, tt.should, got,
				)
			}
		})
	}
}

// simpleMatch reports whether the pattern (as a parsed AST) matches s.
func simpleMatch(tree *ast.Node, s string, seps []rune) bool {
	return step(tree, s, seps, map[int]bool{0: true})[len(s)]
}

func isSep(r rune, seps []rune) bool {
	return slices.Contains(seps, r)
}

// step maps the set of start offsets to the set of end offsets reachable after
// consuming node. Every offset is a position in the original s, so each set has
// at most len(s)+1 elements regardless of how many wildcards appear.
func step(node *ast.Node, s string, seps []rune, starts map[int]bool) map[int]bool {
	switch node.Kind {
	case ast.KindPattern:
		cur := starts
		for _, c := range node.Children {
			cur = step(c, s, seps, cur)
			if len(cur) == 0 {
				return cur
			}
		}
		return cur

	case ast.KindAnyOf:
		out := map[int]bool{}
		for _, c := range node.Children {
			for p := range step(c, s, seps, starts) {
				out[p] = true
			}
		}
		return out

	case ast.KindNothing:
		return starts

	case ast.KindText:
		txt := node.Value.(ast.Text).Text
		out := map[int]bool{}
		for p := range starts {
			if strings.HasPrefix(s[p:], txt) {
				out[p+len(txt)] = true
			}
		}
		return out

	case ast.KindSingle:
		out := map[int]bool{}
		for p := range starts {
			if p < len(s) {
				if r, w := utf8.DecodeRuneInString(s[p:]); !isSep(r, seps) {
					out[p+w] = true
				}
			}
		}
		return out

	case ast.KindCharClass:
		pred := charClassPred(node, seps)
		out := map[int]bool{}
		for p := range starts {
			if p < len(s) {
				if r, w := utf8.DecodeRuneInString(s[p:]); pred(r) {
					out[p+w] = true
				}
			}
		}
		return out

	case ast.KindAny:
		out := map[int]bool{}
		for p := range starts {
			i := p
			out[i] = true // consume nothing
			for i < len(s) {
				r, w := utf8.DecodeRuneInString(s[i:])
				if isSep(r, seps) {
					break
				}
				i += w
				out[i] = true
			}
		}
		return out

	case ast.KindSuper:
		out := map[int]bool{}
		for p := range starts {
			for i := p; ; {
				out[i] = true
				if i >= len(s) {
					break
				}
				_, w := utf8.DecodeRuneInString(s[i:])
				i += w
			}
		}
		return out
	}

	panic("simpleMatch: unhandled kind " + node.Kind.String())
}

// charClassPred builds the membership predicate for a KindCharClass node from
// its KindList and KindRange children, honoring negation.
func charClassPred(node *ast.Node, seps []rune) func(rune) bool {
	not := node.Value.(ast.CharClass).Not
	var list []rune
	var ranges []ast.Range
	for _, c := range node.Children {
		switch c.Kind {
		case ast.KindList:
			list = append(list, []rune(c.Value.(ast.List).Chars)...)
		case ast.KindRange:
			ranges = append(ranges, c.Value.(ast.Range))
		}
	}
	return func(r rune) bool {
		if slices.Contains(seps, r) {
			return false
		}

		in := slices.Contains(list, r)
		for _, rg := range ranges {
			in = in || (r >= rg.Low && r <= rg.High)
		}
		return in != not
	}
}

const patternDepth = 2

var (
	classMemberRunes = append(append([]rune{}, classMembers...), sep)

	patternGen    = rapid.Custom(genPattern)
	fullTargetGen = rapid.Custom(genTarget)
)

func genPattern(t *rapid.T) string {
	var sb strings.Builder
	writePattern(t, &sb, patternDepth)
	return sb.String()
}

func writePattern(t *rapid.T, sb *strings.Builder, depth int) {
	maxTerms := 16 - ((patternDepth - depth) * 4)
	for range rapid.IntRange(1, maxTerms).Draw(t, "nTerms") {
		genTerm(t, sb, depth)
	}
}

func genTerm(t *rapid.T, sb *strings.Builder, depth int) {
	choices := []string{"lit", "escToken", "single", "any", "super", "class"}
	if depth > 0 {
		choices = append(choices, "anyOf")
	}

	switch rapid.SampledFrom(choices).Draw(t, "term") {
	case "lit":
		r := rapid.SampledFrom(classMemberRunes).Draw(t, "lit")
		sb.WriteRune(r)
	case "escToken":
		sb.WriteByte('\\')
		sb.WriteRune(rapid.SampledFrom(tokens).Draw(t, "token"))
	case "single":
		sb.WriteByte('?')
	case "any":
		sb.WriteByte('*')
	case "super":
		sb.WriteString("**")
	case "class":
		sb.WriteByte('[')
		if rapid.Bool().Draw(t, "neg") {
			sb.WriteByte('!')
		}

		for range rapid.IntRange(1, 3).Draw(t, "members") {
			writeClassMember(t, sb)
		}
		sb.WriteByte(']')
	case "anyOf":
		sb.WriteByte('{')
		for i := range rapid.IntRange(1, 3).Draw(t, "alts") {
			if i > 0 {
				sb.WriteByte(',')
			}
			// allow an empty alternative, ex '{a,}'
			if rapid.IntRange(0, 3).Draw(t, "emptyAlt") == 0 {
				continue
			}
			writePattern(t, sb, depth-1)
		}
		sb.WriteByte('}')
	default:
		panic("unhandled term")
	}
}

func writeClassMember(t *rapid.T, sb *strings.Builder) {
	if rapid.Bool().Draw(t, "isRange") {
		var low, high rune
		var lowEscToken, highEscToken bool
		if rapid.Bool().Draw(t, "isLowEscToken") {
			lowEscToken = true
			low = rapid.SampledFrom(tokens).Draw(t, "rangeLow")
		} else {
			low = rapid.SampledFrom(classMemberRunes).Draw(t, "rangeLow")
		}
		if rapid.Bool().Draw(t, "isHighEscToken") {
			highEscToken = true
			high = rapid.SampledFrom(tokens).Draw(t, "rangeHigh")
		} else {
			high = rapid.SampledFrom(classMemberRunes).Draw(t, "rangeHigh")
		}

		if low >= high {
			low = high - 1
			lowEscToken = highEscToken // just to be safe
		}

		if lowEscToken {
			sb.WriteByte('\\')
		}
		sb.WriteRune(low)
		sb.WriteByte('-')
		if highEscToken {
			sb.WriteByte('\\')
		}
		sb.WriteRune(high)
		return
	}

	if rapid.Bool().Draw(t, "isEscToken") {
		sb.WriteByte('\\')
		sb.WriteRune(rapid.SampledFrom(tokens).Draw(t, "token"))
		return
	}

	sb.WriteRune(rapid.SampledFrom(classMembers).Draw(t, "member"))
}

// invalidBytes are lone bytes that never form valid UTF-8 on their own.
var invalidBytes = []byte{0x80, 0xbf, 0xc0, 0xc1, 0xfe, 0xff}

var targetPartGen = rapid.Custom(func(t *rapid.T) string {
	if rapid.IntRange(0, 4).Draw(t, "invalid") == 0 {
		return string([]byte{rapid.SampledFrom(invalidBytes).Draw(t, "badByte")})
	}
	return string(rapid.SampledFrom(classMemberRunes).Draw(t, "validRune"))
})

func genTarget(t *rapid.T) string {
	parts := rapid.SliceOfN(targetPartGen, 0, 16).Draw(t, "targetParts")
	return strings.Join(parts, "")
}

func TestCompileVsSimple(t *testing.T) {
	rapid.Check(t, testCompileVsSimple)
}

func FuzzCompileVsSimple(f *testing.F) {
	f.Fuzz(rapid.MakeFuzz(testCompileVsSimple))
}

// test that optimized compiled matchers have the same behavior as a
// simple AST matcher.
func testCompileVsSimple(t *rapid.T) {
	pattern := patternGen.Draw(t, "pattern")
	target := fullTargetGen.Draw(t, "target")

	tree, err := syntax.Parse(pattern)
	if err != nil {
		t.Fatalf("parse error for %q: %v", pattern, err)
	}
	m, err := compiler.Compile(tree, []rune{sep})
	if err != nil {
		t.Fatalf("compile error for %q: %v", pattern, err)
	}

	got := m.Match(target)
	want := simpleMatch(tree, target, []rune{sep})
	if got != want {
		t.Fatalf(`
match disagreement:
  pattern = %q
  target  = %q
  glob    = %v
  simple   = %v`,
			pattern, target, got, want)
	}
}

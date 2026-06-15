package glob

import (
	"regexp"
	"testing"
)

const (
	pattern_all          = "[a-z][!a-x]*cat*[h][!b]*eyes*"
	regexp_all           = `^[a-z][^a-x].*cat.*[h][^b].*eyes.*$`
	fixture_all_match    = "my cat has very bright eyes"
	fixture_all_mismatch = "my dog has very bright eyes"

	pattern_plain          = "google.com"
	regexp_plain           = `^google\.com$`
	fixture_plain_match    = "google.com"
	fixture_plain_mismatch = "gobwas.com"

	pattern_multiple          = "https://*.google.*"
	regexp_multiple           = `^https:\/\/.*\.google\..*$`
	fixture_multiple_match    = "https://account.google.com"
	fixture_multiple_mismatch = "https://google.com"

	pattern_alternatives          = "{https://*.google.*,*yandex.*,*yahoo.*,*mail.ru}"
	regexp_alternatives           = `^(https:\/\/.*\.google\..*|.*yandex\..*|.*yahoo\..*|.*mail\.ru)$`
	fixture_alternatives_match    = "http://yahoo.com"
	fixture_alternatives_mismatch = "http://google.com"

	pattern_alternatives_suffix                = "{https://*gobwas.com,http://exclude.gobwas.com}"
	regexp_alternatives_suffix                 = `^(https:\/\/.*gobwas\.com|http://exclude.gobwas.com)$`
	fixture_alternatives_suffix_first_match    = "https://safe.gobwas.com"
	fixture_alternatives_suffix_first_mismatch = "http://safe.gobwas.com"
	fixture_alternatives_suffix_second         = "http://exclude.gobwas.com"

	pattern_prefix                 = "abc*"
	regexp_prefix                  = `^abc.*$`
	pattern_suffix                 = "*def"
	regexp_suffix                  = `^.*def$`
	pattern_prefix_suffix          = "ab*ef"
	regexp_prefix_suffix           = `^ab.*ef$`
	fixture_prefix_suffix_match    = "abcdef"
	fixture_prefix_suffix_mismatch = "af"

	pattern_alternatives_combine_lite = "{abc*def,abc?def,abc[zte]def}"
	regexp_alternatives_combine_lite  = `^(abc.*def|abc.def|abc[zte]def)$`
	fixture_alternatives_combine_lite = "abczdef"

	pattern_alternatives_combine_hard = "{abc*[a-c]def,abc?[d-g]def,abc[zte]?def}"
	regexp_alternatives_combine_hard  = `^(abc.*[a-c]def|abc.[d-g]def|abc[zte].def)$`
	fixture_alternatives_combine_hard = "abczqdef"
)

type test struct {
	pattern, match string
	should         bool
	delimiters     []rune
}

var tests = []test{
	{should: true, pattern: "* ?at * eyes", match: "my cat has very bright eyes"},

	{should: true, pattern: "", match: ""},
	{should: false, pattern: "", match: "b"},

	{should: true, pattern: "*ä", match: "åä"},
	{should: true, pattern: "abc", match: "abc"},
	{should: true, pattern: "a*c", match: "abc"},
	{should: true, pattern: "a*c", match: "a12345c"},
	{should: true, pattern: "a?c", match: "a1c"},
	{should: true, pattern: "a.b", match: "a.b", delimiters: []rune{'.'}},
	{should: true, pattern: "a.*", match: "a.b", delimiters: []rune{'.'}},
	{should: true, pattern: "a.**", match: "a.b.c", delimiters: []rune{'.'}},
	{should: true, pattern: "a.?.c", match: "a.b.c", delimiters: []rune{'.'}},
	{should: true, pattern: "a.?.?", match: "a.b.c", delimiters: []rune{'.'}},
	{should: true, pattern: "?at", match: "cat"},
	{should: true, pattern: "?at", match: "fat"},
	{should: true, pattern: "*", match: "abc"},
	{should: true, pattern: `\*`, match: "*"},
	{should: true, pattern: "**", match: "a.b.c", delimiters: []rune{'.'}},

	{should: false, pattern: "?at", match: "at"},
	{should: false, pattern: "?at", match: "fat", delimiters: []rune{'f'}},
	{should: false, pattern: "a.*", match: "a.b.c", delimiters: []rune{'.'}},
	{should: false, pattern: "a.?.c", match: "a.bb.c", delimiters: []rune{'.'}},
	{should: false, pattern: "*", match: "a.b.c", delimiters: []rune{'.'}},

	{should: true, pattern: "*test", match: "this is a test"},
	{should: true, pattern: "this*", match: "this is a test"},
	{should: true, pattern: "*is *", match: "this is a test"},
	{should: true, pattern: "*is*a*", match: "this is a test"},
	{should: true, pattern: "**test**", match: "this is a test"},
	{should: true, pattern: "**is**a***test*", match: "this is a test"},

	{should: false, pattern: "*is", match: "this is a test"},
	{should: false, pattern: "*no*", match: "this is a test"},
	{should: true, pattern: "[!a]*", match: "this is a test3"},

	{should: true, pattern: "*abc", match: "abcabc"},
	{should: true, pattern: "**abc", match: "abcabc"},
	{should: true, pattern: "???", match: "abc"},
	{should: true, pattern: "?*?", match: "abc"},
	{should: true, pattern: "?*?", match: "ac"},
	{should: false, pattern: "sta", match: "stagnation"},
	{should: true, pattern: "sta*", match: "stagnation"},
	{should: false, pattern: "sta?", match: "stagnation"},
	{should: false, pattern: "sta?n", match: "stagnation"},

	{should: true, pattern: "{abc,def}ghi", match: "defghi"},
	{should: true, pattern: "{abc,abcd}a", match: "abcda"},
	{should: true, pattern: "{a,ab}{bc,f}", match: "abc"},
	{should: true, pattern: "{*,**}{a,b}", match: "ab"},
	{should: false, pattern: "{*,**}{a,b}", match: "ac"},

	{should: true, pattern: "/{rate,[a-z][a-z][a-z]}*", match: "/rate"},
	{should: true, pattern: "/{rate,[0-9][0-9][0-9]}*", match: "/rate"},
	{should: true, pattern: "/{rate,[a-z][a-z][a-z]}*", match: "/usd"},

	{should: true, pattern: "{*.google.*,*.yandex.*}", match: "www.google.com", delimiters: []rune{'.'}},
	{should: true, pattern: "{*.google.*,*.yandex.*}", match: "www.yandex.com", delimiters: []rune{'.'}},
	{should: false, pattern: "{*.google.*,*.yandex.*}", match: "yandex.com", delimiters: []rune{'.'}},
	{should: false, pattern: "{*.google.*,*.yandex.*}", match: "google.com", delimiters: []rune{'.'}},

	{should: true, pattern: "{*.google.*,yandex.*}", match: "www.google.com", delimiters: []rune{'.'}},
	{should: true, pattern: "{*.google.*,yandex.*}", match: "yandex.com", delimiters: []rune{'.'}},
	{should: false, pattern: "{*.google.*,yandex.*}", match: "www.yandex.com", delimiters: []rune{'.'}},
	{should: false, pattern: "{*.google.*,yandex.*}", match: "google.com", delimiters: []rune{'.'}},

	{should: true, pattern: "*//{,*.}example.com", match: "https://www.example.com"},
	{should: true, pattern: "*//{,*.}example.com", match: "http://example.com"},
	{should: false, pattern: "*//{,*.}example.com", match: "http://example.com.net"},

	{should: true, pattern: pattern_all, match: fixture_all_match},
	{should: false, pattern: pattern_all, match: fixture_all_mismatch},

	{should: true, pattern: pattern_plain, match: fixture_plain_match},
	{should: false, pattern: pattern_plain, match: fixture_plain_mismatch},

	{should: true, pattern: pattern_multiple, match: fixture_multiple_match},
	{should: false, pattern: pattern_multiple, match: fixture_multiple_mismatch},

	{should: true, pattern: pattern_alternatives, match: fixture_alternatives_match},
	{should: false, pattern: pattern_alternatives, match: fixture_alternatives_mismatch},

	{should: true, pattern: pattern_alternatives_suffix, match: fixture_alternatives_suffix_first_match},
	{should: false, pattern: pattern_alternatives_suffix, match: fixture_alternatives_suffix_first_mismatch},
	{should: true, pattern: pattern_alternatives_suffix, match: fixture_alternatives_suffix_second},

	{should: true, pattern: pattern_alternatives_combine_hard, match: fixture_alternatives_combine_hard},

	{should: true, pattern: pattern_alternatives_combine_lite, match: fixture_alternatives_combine_lite},

	{should: true, pattern: pattern_prefix, match: fixture_prefix_suffix_match},
	{should: false, pattern: pattern_prefix, match: fixture_prefix_suffix_mismatch},

	{should: true, pattern: pattern_suffix, match: fixture_prefix_suffix_match},
	{should: false, pattern: pattern_suffix, match: fixture_prefix_suffix_mismatch},

	{should: true, pattern: pattern_prefix_suffix, match: fixture_prefix_suffix_match},
	{should: false, pattern: pattern_prefix_suffix, match: fixture_prefix_suffix_mismatch},
}

func TestGlob(t *testing.T) {
	for _, test := range tests {
		t.Run("", func(t *testing.T) {
			g := MustCompile(test.pattern, test.delimiters...)
			result := g.Match(test.match)
			if result != test.should {
				t.Errorf("pattern %q matching %q should be %v but got %v\n%s",
					test.pattern, test.match, test.should, result, g,
				)
			}
		})
	}
}

func TestQuoteMeta(t *testing.T) {
	for id, test := range []struct {
		in, out string
	}{
		{
			in:  `[foo*]`,
			out: `\[foo\*\]`,
		},
		{
			in:  `{foo*}`,
			out: `\{foo\*\}`,
		},
		{
			in:  `*?\[]{}`,
			out: `\*\?\\\[\]\{\}`,
		},
		{
			in:  `some text and *?\[]{}`,
			out: `some text and \*\?\\\[\]\{\}`,
		},
	} {
		act := QuoteMeta(test.in)
		if act != test.out {
			t.Errorf("#%d QuoteMeta(%q) = %q; want %q", id, test.in, act, test.out)
		}
		if _, err := Compile(act); err != nil {
			t.Errorf("#%d _, err := Compile(QuoteMeta(%q) = %q); err = %q", id, test.in, act, err)
		}
	}
}

func BenchmarkParseGlob(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Compile(pattern_all)
	}
}

func BenchmarkParseRegexp(b *testing.B) {
	for i := 0; i < b.N; i++ {
		regexp.MustCompile(regexp_all)
	}
}

func BenchmarkAllGlobMatch(b *testing.B) {
	m, _ := Compile(pattern_all)

	for i := 0; i < b.N; i++ {
		_ = m.Match(fixture_all_match)
	}
}

func BenchmarkAllGlobMatchParallel(b *testing.B) {
	m, _ := Compile(pattern_all)

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = m.Match(fixture_all_match)
		}
	})
}

func BenchmarkAllRegexpMatch(b *testing.B) {
	m := regexp.MustCompile(regexp_all)
	f := []byte(fixture_all_match)

	for i := 0; i < b.N; i++ {
		_ = m.Match(f)
	}
}

func BenchmarkAllGlobMismatch(b *testing.B) {
	m, _ := Compile(pattern_all)

	for i := 0; i < b.N; i++ {
		_ = m.Match(fixture_all_mismatch)
	}
}

func BenchmarkAllGlobMismatchParallel(b *testing.B) {
	m, _ := Compile(pattern_all)

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			_ = m.Match(fixture_all_mismatch)
		}
	})
}

func BenchmarkAllRegexpMismatch(b *testing.B) {
	m := regexp.MustCompile(regexp_all)
	f := []byte(fixture_all_mismatch)

	for i := 0; i < b.N; i++ {
		_ = m.Match(f)
	}
}

func BenchmarkMultipleGlobMatch(b *testing.B) {
	m, _ := Compile(pattern_multiple)

	for i := 0; i < b.N; i++ {
		_ = m.Match(fixture_multiple_match)
	}
}

func BenchmarkMultipleRegexpMatch(b *testing.B) {
	m := regexp.MustCompile(regexp_multiple)
	f := []byte(fixture_multiple_match)

	for i := 0; i < b.N; i++ {
		_ = m.Match(f)
	}
}

func BenchmarkMultipleGlobMismatch(b *testing.B) {
	m, _ := Compile(pattern_multiple)

	for i := 0; i < b.N; i++ {
		_ = m.Match(fixture_multiple_mismatch)
	}
}

func BenchmarkMultipleRegexpMismatch(b *testing.B) {
	m := regexp.MustCompile(regexp_multiple)
	f := []byte(fixture_multiple_mismatch)

	for i := 0; i < b.N; i++ {
		_ = m.Match(f)
	}
}

func BenchmarkAlternativesGlobMatch(b *testing.B) {
	m, _ := Compile(pattern_alternatives)

	for i := 0; i < b.N; i++ {
		_ = m.Match(fixture_alternatives_match)
	}
}

func BenchmarkAlternativesGlobMismatch(b *testing.B) {
	m, _ := Compile(pattern_alternatives)

	for i := 0; i < b.N; i++ {
		_ = m.Match(fixture_alternatives_mismatch)
	}
}

func BenchmarkAlternativesRegexpMatch(b *testing.B) {
	m := regexp.MustCompile(regexp_alternatives)
	f := []byte(fixture_alternatives_match)

	for i := 0; i < b.N; i++ {
		_ = m.Match(f)
	}
}

func BenchmarkAlternativesRegexpMismatch(b *testing.B) {
	m := regexp.MustCompile(regexp_alternatives)
	f := []byte(fixture_alternatives_mismatch)

	for i := 0; i < b.N; i++ {
		_ = m.Match(f)
	}
}

func BenchmarkAlternativesSuffixFirstGlobMatch(b *testing.B) {
	m, _ := Compile(pattern_alternatives_suffix)

	for i := 0; i < b.N; i++ {
		_ = m.Match(fixture_alternatives_suffix_first_match)
	}
}

func BenchmarkAlternativesSuffixFirstGlobMismatch(b *testing.B) {
	m, _ := Compile(pattern_alternatives_suffix)

	for i := 0; i < b.N; i++ {
		_ = m.Match(fixture_alternatives_suffix_first_mismatch)
	}
}

func BenchmarkAlternativesSuffixSecondGlobMatch(b *testing.B) {
	m, _ := Compile(pattern_alternatives_suffix)

	for i := 0; i < b.N; i++ {
		_ = m.Match(fixture_alternatives_suffix_second)
	}
}

func BenchmarkAlternativesCombineLiteGlobMatch(b *testing.B) {
	m, _ := Compile(pattern_alternatives_combine_lite)

	for i := 0; i < b.N; i++ {
		_ = m.Match(fixture_alternatives_combine_lite)
	}
}

func BenchmarkAlternativesCombineHardGlobMatch(b *testing.B) {
	m, _ := Compile(pattern_alternatives_combine_hard)

	for i := 0; i < b.N; i++ {
		_ = m.Match(fixture_alternatives_combine_hard)
	}
}

func BenchmarkAlternativesSuffixFirstRegexpMatch(b *testing.B) {
	m := regexp.MustCompile(regexp_alternatives_suffix)
	f := []byte(fixture_alternatives_suffix_first_match)

	for i := 0; i < b.N; i++ {
		_ = m.Match(f)
	}
}

func BenchmarkAlternativesSuffixFirstRegexpMismatch(b *testing.B) {
	m := regexp.MustCompile(regexp_alternatives_suffix)
	f := []byte(fixture_alternatives_suffix_first_mismatch)

	for i := 0; i < b.N; i++ {
		_ = m.Match(f)
	}
}

func BenchmarkAlternativesSuffixSecondRegexpMatch(b *testing.B) {
	m := regexp.MustCompile(regexp_alternatives_suffix)
	f := []byte(fixture_alternatives_suffix_second)

	for i := 0; i < b.N; i++ {
		_ = m.Match(f)
	}
}

func BenchmarkAlternativesCombineLiteRegexpMatch(b *testing.B) {
	m := regexp.MustCompile(regexp_alternatives_combine_lite)
	f := []byte(fixture_alternatives_combine_lite)

	for i := 0; i < b.N; i++ {
		_ = m.Match(f)
	}
}

func BenchmarkAlternativesCombineHardRegexpMatch(b *testing.B) {
	m := regexp.MustCompile(regexp_alternatives_combine_hard)
	f := []byte(fixture_alternatives_combine_hard)

	for i := 0; i < b.N; i++ {
		_ = m.Match(f)
	}
}

func BenchmarkPlainGlobMatch(b *testing.B) {
	m, _ := Compile(pattern_plain)

	for i := 0; i < b.N; i++ {
		_ = m.Match(fixture_plain_match)
	}
}

func BenchmarkPlainRegexpMatch(b *testing.B) {
	m := regexp.MustCompile(regexp_plain)
	f := []byte(fixture_plain_match)

	for i := 0; i < b.N; i++ {
		_ = m.Match(f)
	}
}

func BenchmarkPlainGlobMismatch(b *testing.B) {
	m, _ := Compile(pattern_plain)

	for i := 0; i < b.N; i++ {
		_ = m.Match(fixture_plain_mismatch)
	}
}

func BenchmarkPlainRegexpMismatch(b *testing.B) {
	m := regexp.MustCompile(regexp_plain)
	f := []byte(fixture_plain_mismatch)

	for i := 0; i < b.N; i++ {
		_ = m.Match(f)
	}
}

func BenchmarkPrefixGlobMatch(b *testing.B) {
	m, _ := Compile(pattern_prefix)

	for i := 0; i < b.N; i++ {
		_ = m.Match(fixture_prefix_suffix_match)
	}
}

func BenchmarkPrefixRegexpMatch(b *testing.B) {
	m := regexp.MustCompile(regexp_prefix)
	f := []byte(fixture_prefix_suffix_match)

	for i := 0; i < b.N; i++ {
		_ = m.Match(f)
	}
}

func BenchmarkPrefixGlobMismatch(b *testing.B) {
	m, _ := Compile(pattern_prefix)

	for i := 0; i < b.N; i++ {
		_ = m.Match(fixture_prefix_suffix_mismatch)
	}
}

func BenchmarkPrefixRegexpMismatch(b *testing.B) {
	m := regexp.MustCompile(regexp_prefix)
	f := []byte(fixture_prefix_suffix_mismatch)

	for i := 0; i < b.N; i++ {
		_ = m.Match(f)
	}
}

func BenchmarkSuffixGlobMatch(b *testing.B) {
	m, _ := Compile(pattern_suffix)

	for i := 0; i < b.N; i++ {
		_ = m.Match(fixture_prefix_suffix_match)
	}
}

func BenchmarkSuffixRegexpMatch(b *testing.B) {
	m := regexp.MustCompile(regexp_suffix)
	f := []byte(fixture_prefix_suffix_match)

	for i := 0; i < b.N; i++ {
		_ = m.Match(f)
	}
}

func BenchmarkSuffixGlobMismatch(b *testing.B) {
	m, _ := Compile(pattern_suffix)

	for i := 0; i < b.N; i++ {
		_ = m.Match(fixture_prefix_suffix_mismatch)
	}
}

func BenchmarkSuffixRegexpMismatch(b *testing.B) {
	m := regexp.MustCompile(regexp_suffix)
	f := []byte(fixture_prefix_suffix_mismatch)

	for i := 0; i < b.N; i++ {
		_ = m.Match(f)
	}
}

func BenchmarkPrefixSuffixGlobMatch(b *testing.B) {
	m, _ := Compile(pattern_prefix_suffix)

	for i := 0; i < b.N; i++ {
		_ = m.Match(fixture_prefix_suffix_match)
	}
}

func BenchmarkPrefixSuffixRegexpMatch(b *testing.B) {
	m := regexp.MustCompile(regexp_prefix_suffix)
	f := []byte(fixture_prefix_suffix_match)

	for i := 0; i < b.N; i++ {
		_ = m.Match(f)
	}
}

func BenchmarkPrefixSuffixGlobMismatch(b *testing.B) {
	m, _ := Compile(pattern_prefix_suffix)

	for i := 0; i < b.N; i++ {
		_ = m.Match(fixture_prefix_suffix_mismatch)
	}
}

func BenchmarkPrefixSuffixRegexpMismatch(b *testing.B) {
	m := regexp.MustCompile(regexp_prefix_suffix)
	f := []byte(fixture_prefix_suffix_mismatch)

	for i := 0; i < b.N; i++ {
		_ = m.Match(f)
	}
}

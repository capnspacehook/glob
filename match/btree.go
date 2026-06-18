package match

import (
	"fmt"
	"unicode/utf8"
)

type BTree struct {
	Value Matcher
	Left  Matcher
	Right Matcher

	ValueLengthRunes int
	LeftLengthRunes  int
	RightLengthRunes int
	LengthRunes      int
}

func NewBTree(Value, Left, Right Matcher) (tree BTree) {
	tree.Value = Value
	tree.Left = Left
	tree.Right = Right

	lenOk := true
	if tree.ValueLengthRunes = Value.Len(); tree.ValueLengthRunes == -1 {
		lenOk = false
	}

	if Left != nil {
		if tree.LeftLengthRunes = Left.Len(); tree.LeftLengthRunes == -1 {
			lenOk = false
		}
	}

	if Right != nil {
		if tree.RightLengthRunes = Right.Len(); tree.RightLengthRunes == -1 {
			lenOk = false
		}
	}

	if lenOk {
		tree.LengthRunes = tree.LeftLengthRunes + tree.ValueLengthRunes + tree.RightLengthRunes
	} else {
		tree.LengthRunes = -1
	}

	return tree
}

func (b BTree) Len() int {
	return b.LengthRunes
}

func (b BTree) Index(s string) (int, []int) {
	for i := 0; i <= len(s); {
		var segments []int
		var length int

		for {
			if b.Match(s[i : i+length]) {
				segments = append(segments, length)
			}
			if i+length >= len(s) {
				break
			}
			_, w := utf8.DecodeRuneInString(s[i+length:])
			length += w
		}

		if len(segments) > 0 {
			return i, segments
		}
		if i >= len(s) {
			break
		}

		_, w := utf8.DecodeRuneInString(s[i:])
		i += w
	}

	return -1, nil
}

func (b BTree) Match(s string) bool {
	inputLen := len(s)
	// try to cut unnecessary parts
	// by knowledge of length of right and left part
	offset, limit := b.offsetLimit(inputLen)

	// offset may equal limit if s is empty
	for offset <= limit {
		// search for matching part in substring
		index, segments := b.Value.Index(s[offset:limit])
		if index == -1 {
			releaseSegments(segments)
			return false
		}

		l := s[:offset+index]
		var left bool
		if b.Left != nil {
			left = b.Left.Match(l)
		} else {
			left = l == ""
		}

		if left {
			for i := len(segments) - 1; i >= 0; i-- {
				length := segments[i]

				var right bool
				var r string
				// if there is no string for the right branch
				if inputLen <= offset+index+length {
					r = ""
				} else {
					r = s[offset+index+length:]
				}

				if b.Right != nil {
					right = b.Right.Match(r)
				} else {
					right = r == ""
				}

				if right {
					releaseSegments(segments)
					return true
				}
			}
		}

		_, step := utf8.DecodeRuneInString(s[offset+index:])
		offset += index + step

		releaseSegments(segments)

		// avoid an infinite loop if there are no more runes in the string
		if step == 0 {
			break
		}
	}

	return false
}

func (b BTree) offsetLimit(inputLen int) (offset int, limit int) {
	// self.Length, self.RLen and self.LLen are values meaning the length of runes for each part
	// here we manipulating byte length for better optimizations
	// but these checks still works, cause minLen of 1-rune string is 1 byte.
	if b.LengthRunes != -1 && b.LengthRunes > inputLen {
		return 0, 0
	}
	if b.LeftLengthRunes >= 0 {
		offset = b.LeftLengthRunes
	}
	if b.RightLengthRunes >= 0 {
		limit = inputLen - b.RightLengthRunes
	} else {
		limit = inputLen
	}
	return offset, limit
}

func (b BTree) String() string {
	const n string = "<nil>"
	var l, r string
	if b.Left == nil {
		l = n
	} else {
		l = b.Left.String()
	}
	if b.Right == nil {
		r = n
	} else {
		r = b.Right.String()
	}

	return fmt.Sprintf("<btree:[%s<-%s->%s]>", l, b.Value, r)
}

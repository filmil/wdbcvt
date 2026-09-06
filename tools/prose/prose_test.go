// SPDX-License-Identifier: Apache-2.0

package prose

import "testing"

func TestCheck(t *testing.T) {
	cases := []struct {
		name string
		text string
		want int
	}{
		{"data as the actor", "each object takes its value size\n", 1},
		{"the file as the actor", "the file reserves 8 bytes for it\n", 0},
		{"a verb for texture", "the skill carries the gotchas\n", 1},
		{"an invented compound", "the question is case-movable\n", 1},
		{"an em dash", "a value — an aside — and on\n", 1},
		{"a fenced block", "```\neach object takes 8 bytes\n```\n", 0},
		{"an indented command", "    grep -n 'each value takes' file\n", 0},
		{"the allow marker", "<!-- prose-lint: allow -->\neach object takes 8 bytes\nand each value costs one\n", 0},
		{"the allow marker ends at a blank line", "<!-- prose-lint: allow -->\neach object takes 8 bytes\n\neach value costs one\n", 1},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := len(Check("t.md", c.text)); got != c.want {
				t.Errorf("Check found %d hits, want %d", got, c.want)
			}
		})
	}
}

// SPDX-License-Identifier: Apache-2.0

// Package prose holds the checks that keep the documents of this
// repository to the writing rules in AGENTS.md. The rules that can be
// checked mechanically are checked here, so that breaking one fails a
// test rather than waiting for a reader to notice.
package prose

import (
	"fmt"
	"regexp"
	"strings"
)

// Rule is one thing the prose must not do.
type Rule struct {
	// Name appears in the failure.
	Name string
	// Pat matches the offending text.
	Pat *regexp.Regexp
	// Why tells the writer what to do instead.
	Why string
}

// Rules are the checks, in the order they are reported.
var Rules = []Rule{
	{
		Name: "data as the actor",
		Pat: regexp.MustCompile(`(?i)\b(object|value|type|field|record|entry|signal|package|scope|generic|parameter|variable|constant|array|string|handle|chunk|arena|page|case)s?\s+(takes?|costs?|wants?|knows|owns?|decides?)\b`),
		Why: "name what acts: the file, the writer, the reader, the simulator, or the frame. " +
			"Write \"the handle space grows by 8 bytes for one integer\", not \"each object takes 8 bytes\".",
	},
	{
		Name: "a verb chosen for texture",
		Pat:  regexp.MustCompile(`(?i)\bcarr(y|ies|ied|ying)\b`),
		Why:  "say what it does: holds, lists, states, sets, needs, provides, names.",
	},
	{
		Name: "an invented compound",
		Pat:  regexp.MustCompile(`(?i)\b(case-movable|corpus-movable|self-evidencing|finding-shaped|tier-shaped)\b`),
		Why:  "a term this repository coined and no reader has. Say the thing in words.",
	},
	{
		Name: "a word that adds nothing",
		Pat:  regexp.MustCompile(`(?i)\b(in order to|it should be noted that|essentially|simply put|needless to say|of course,)`),
		Why:  "cut it, or say the thing it stands in front of.",
	},
	{
		Name: "an em dash or an en dash",
		Pat:  regexp.MustCompile(`[\x{2014}\x{2013}]`),
		Why:  "start a new sentence, use parentheses, or delete the aside.",
	},
}

// Hit is one match of one rule.
type Hit struct {
	File string
	Line int
	Rule Rule
	Text string
}

func (h Hit) String() string {
	return fmt.Sprintf("%s:%d: %s: %q\n    %s", h.File, h.Line, h.Rule.Name, strings.TrimSpace(h.Text), h.Rule.Why)
}

// allow marks a passage the checks skip, for a document that quotes a
// violation in order to correct it. It holds from the marker to the
// next blank line.
const allow = "prose-lint: allow"

// Check reports every rule a document breaks. Fenced code blocks and
// indented command lines are not prose and are skipped, and so is any
// line that carries the allow marker or follows one.
func Check(file, text string) []Hit {
	var hits []Hit
	fenced := false
	allowed := false
	for i, line := range strings.Split(text, "\n") {
		if strings.HasPrefix(strings.TrimSpace(line), "```") {
			fenced = !fenced
			continue
		}
		if strings.TrimSpace(line) == "" {
			allowed = false
			continue
		}
		if strings.Contains(line, allow) {
			allowed = true
			continue
		}
		if fenced || allowed {
			continue
		}
		// An indented block is a command or a listing, not prose.
		if strings.HasPrefix(line, "    ") || strings.HasPrefix(line, "\t") {
			continue
		}
		for _, r := range Rules {
			if r.Pat.MatchString(line) {
				hits = append(hits, Hit{File: file, Line: i + 1, Rule: r, Text: line})
			}
		}
	}
	return hits
}

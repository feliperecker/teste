package helpers

import (
	"errors"
	"html"
	"html/template"
	"regexp"
	"unicode"
	"unicode/utf8"

	"github.com/spf13/cast"
)

var (
	tagRE        = regexp.MustCompile(`^<(/)?([^ ]+?)(?:(\s*/)| .*?)?>`)
	htmlSinglets = map[string]bool{
		"br": true, "col": true, "link": true,
		"base": true, "img": true, "param": true,
		"area": true, "hr": true, "input": true,
	}
)

type htmlTag struct {
	name    string
	pos     int
	openTag bool
}

// Truncate truncates a given string to the specified length.
func Truncate(textParam interface{}, length int, ellipsis string) (template.HTML, error) {
	text, err := cast.ToStringE(textParam)
	if err != nil {
		return "", errors.New("text must be a string")
	}

	_, isHTML := textParam.(template.HTML)

	if utf8.RuneCountInString(text) <= length {
		if isHTML {
			return template.HTML(text), nil
		}
		return template.HTML(html.EscapeString(text)), nil
	}

	tags := []htmlTag{}
	var lastWordIndex, lastNonSpace, currentLen, endTextPos, nextTag int

	for i, r := range text {
		if i < nextTag {
			continue
		}

		if isHTML {
			// Make sure we keep tag of HTML tags
			slice := text[i:]
			m := tagRE.FindStringSubmatchIndex(slice)
			if len(m) > 0 && m[0] == 0 {
				nextTag = i + m[1]
				tagname := slice[m[4]:m[5]]
				lastWordIndex = lastNonSpace
				_, singlet := htmlSinglets[tagname]
				if !singlet && m[6] == -1 {
					tags = append(tags, htmlTag{name: tagname, pos: i, openTag: m[2] == -1})
				}

				continue
			}
		}

		currentLen++
		if unicode.IsSpace(r) {
			lastWordIndex = lastNonSpace
		} else if unicode.In(r, unicode.Han, unicode.Hangul, unicode.Hiragana, unicode.Katakana) {
			lastWordIndex = i
		} else {
			lastNonSpace = i + utf8.RuneLen(r)
		}

		if currentLen > length {
			if lastWordIndex == 0 {
				endTextPos = i
			} else {
				endTextPos = lastWordIndex
			}
			out := text[0:endTextPos]
			if isHTML {
				out += ellipsis
				// Close out any open HTML tags
				var currentTag *htmlTag
				for i := len(tags) - 1; i >= 0; i-- {
					tag := tags[i]
					if tag.pos >= endTextPos || currentTag != nil {
						if currentTag != nil && currentTag.name == tag.name {
							currentTag = nil
						}
						continue
					}

					if tag.openTag {
						out += ("</" + tag.name + ">")
					} else {
						currentTag = &tag
					}
				}

				return template.HTML(out), nil
			}
			return template.HTML(html.EscapeString(out) + ellipsis), nil
		}
	}

	if isHTML {
		return template.HTML(text), nil
	}
	return template.HTML(html.EscapeString(text)), nil
}

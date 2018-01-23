package regexp

import (
	"fmt"
	"regexp"

	"qiniupkg.com/http/httputil.v2"
)

// ---------------------------------------------------------------------------

type match struct {
	Regexp *regexp.Regexp
	Prompt string
}

func Match(prompt string, regex *regexp.Regexp) match {
	return match{regex, prompt}
}

func (p match) Validate(k, v string) error {
	if p.Regexp.MatchString(v) {
		return nil
	}
	return httputil.NewError(400, fmt.Sprintf("%s is invalid: %s", k, p.Prompt))
}

// ---------------------------------------------------------------------------

type email struct {
	match
}

var regexpEmail = regexp.MustCompile(
	"^[\\w!#$%&'*+/=?^_`{|}~-]+(?:\\.[\\w!#$%&'*+/=?^_`{|}~-]+)*@(?:[\\w](?:[\\w-]*[\\w])?\\.)+[a-zA-Z0-9](?:[\\w-]*[\\w])?$")

func Email() match {
	return Match("must be a valid email address", regexpEmail)
}

// ---------------------------------------------------------------------------


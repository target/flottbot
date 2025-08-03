// SPDX-License-Identifier: Apache-2.0

package text

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/rs/zerolog/log"
)

// Match checks given value against given pattern.
func Match(pattern, value string, trimValue bool) (string, bool) {
	// set the default regex pattern; assumes given pattern is not regex already
	regxPattern := fmt.Sprintf(`(?i)^(%s$|%s[^\S])`, pattern, pattern)
	// check if we're dealing with regex
	isRegex := strings.HasPrefix(pattern, "/") && strings.HasSuffix(pattern, "/")

	// check if the given pattern was a regex expression
	if isRegex {
		regxPattern = fmt.Sprintf(`(?i)%s`, pattern[1:len(pattern)-1])
	}

	// try to compile the regex,
	regx, err := regexp.Compile(regxPattern)
	if err != nil {
		log.Error().Msgf("unsupported regex: %s", pattern)

		return "", false
	}

	// check whether given value matches regex
	matchFound := regx.MatchString(value)
	if !matchFound {
		return "", false
	}

	// remove the regex match from the given value and trim the space
	if trimValue {
		value = strings.Trim(strings.Replace(value, regx.FindString(value), "", 1), " ")
	}

	return value, matchFound
}

// Substitute checks given value for variables and looks them up
// to determine whether we have a matching replacement available,
// either in the supplied map, or from environment variables.
func Substitute(value string, tokens map[string]string) (string, error) {
	var errs []string

	if match, hits := findVars(value); match {
		for _, hit := range hits {
			tok := strip(hit)
			// Check if token was already stored as a token
			if _, ok := tokens[tok]; ok {
				// TODO: check on this
				envTok := os.Getenv(tok)
				if envTok != "" {
					log.Warn().Msgf("you are using %s as %#q but it is also an environment variable. consider renaming.", tok, tok)
				}

				value = strings.ReplaceAll(value, hit, orDefault(tokens[tok], ""))

				continue
			}
			// Check if token is an environment variable
			envTok := os.Getenv(tok)

			if envTok != "" {
				value = strings.ReplaceAll(value, hit, os.Getenv(tok))
			} else {
				err := fmt.Sprintf("Variable %#q has not been defined.", tok)
				errs = append(errs, err)
			}
		}
	}
	// Concat any caught errors into one error message and return it with unsubstituted value
	if len(errs) > 0 {
		errMsg := strings.Join(errs, " ")
		return value, errors.New(errMsg)
	}

	return value, nil
}

// RuleArgTokenizer goes through a string and tokenizes as parameters for use when identifying rules to be triggered (ignoring empty arguments).
func RuleArgTokenizer(stripped string) []string {
	re := regexp.MustCompile(`["“]([^"“”]+)["”]|([^"“”\s]+)`)
	argmatch := re.FindAllString(stripped, -1)

	for i, arg := range argmatch {
		argmatch[i] = strings.Trim(arg, `"“”`)
	}

	return argmatch
}

// ExecArgTokenizer goes through a string and tokenizes as parameters for use when executing a script (respecting empty arguments).
func ExecArgTokenizer(stripped string) []string {
	re := regexp.MustCompile(`('[^']*')|("[^"]*")|“([^”]*”)|([^'"“”\s]+)`)
	argmatch := re.FindAllString(stripped, -1)

	for i, arg := range argmatch {
		argmatch[i] = strings.Trim(arg, `'"“”`)
	}

	return argmatch
}

// find variables within strings with pattern ${var}.
func findVars(value string) (match bool, tokens []string) {
	match = false
	re := regexp.MustCompile(`\${([A-Za-z0-9:*_\|\-\.\?]+)}`)
	tokens = re.FindAllString(strings.ReplaceAll(value, "$${", "X{"), -1)

	if len(tokens) > 0 {
		match = true
	}

	return match, tokens
}

// helper to provide default value.
func orDefault(value, def string) string {
	if strings.TrimSpace(value) == "" {
		return def
	}

	return value
}

// strip variable demarcations.
func strip(value string) (stripped string) {
	stripped = strings.ReplaceAll(value, "${", "")
	stripped = strings.ReplaceAll(stripped, "}", "")

	return stripped
}

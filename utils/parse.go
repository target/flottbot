package utils

import (
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"
)

// Match checks given value against given pattern
func Match(pattern, value string, trimInput bool) (string, bool) {
	var regx *regexp.Regexp
	re := strings.HasPrefix(pattern, "/") && strings.HasSuffix(pattern, "/")

	if re {
		pattern = strings.Replace(pattern, "/", "", -1)
		regx = regexp.MustCompile("(?i)" + pattern)
	} else {
		regx = regexp.MustCompile(fmt.Sprintf(`(?i)^(%s$|%s[^\S])`, pattern, pattern))
	}

	input := value
	if trimInput {
		input = strings.Replace(value, regx.FindString(value), "", 1)
	}

	matchFound := regx.MatchString(value)
	if !matchFound {
		return "", false
	}

	return strings.Trim(input, " "), regx.MatchString(value)
}

// Substitute checks given value for variables and looks them up to determine whether we
// have a matching replacement available.
// If tokens is not supplied, it wil
func Substitute(value string, tokens map[string]string) (string, error) {
	var errs []string
	if match, hits := findVars(value); match {
		for _, hit := range hits {
			tok := strip(hit)
			// Check if token was already stored as a token
			if _, ok := tokens[tok]; ok {
				envTok := os.Getenv(tok)
				if envTok != "" {
					log.Printf("Warning: you are using %s as '%s' but it is also an environment variable. Consider renaming.", tok, tok)
				}
				value = strings.Replace(value, hit, orDefault(tokens[tok], ""), -1)
				continue
			}
			// Check if token is an environment variable
			envTok := os.Getenv(tok)
			if envTok != "" {
				value = strings.Replace(value, hit, os.Getenv(tok), -1)
			} else {
				err := fmt.Sprintf("Variable '%s' has not been defined.", tok)
				errs = append(errs, err)
			}
		}
	}
	// Concat any caught errors into one error message and return it with unsubstituted value
	if len(errs) > 0 {
		errMsg := strings.Join(errs, " ")
		return value, fmt.Errorf(errMsg)
	}
	return value, nil
}

// RuleArgTokenizer goes through a string and tokenizes as parameters for use when identifying rules to be triggered (ignoring empty arguments)
func RuleArgTokenizer(stripped string) []string {
	re := regexp.MustCompile(`["“]([^"“”]+)["”]|([^"“”\s]+)`)
	argmatch := re.FindAllString(stripped, -1)

	for i, arg := range argmatch {
		argmatch[i] = strings.Trim(arg, `"“”`)
	}

	return argmatch
}

// ExecArgTokenizer goes through a string and tokenizes as parameters for use when executing a script (respecting empty arguments)
func ExecArgTokenizer(stripped string) []string {
	re := regexp.MustCompile(`('[^']*')|("[^"]*")|“([^”]*”)|([^'"“”\s]+)`)
	argmatch := re.FindAllString(stripped, -1)

	for i, arg := range argmatch {
		argmatch[i] = strings.Trim(arg, `'"“”`)
	}

	return argmatch
}

// find variables within strings with pattern ${var}
func findVars(value string) (match bool, tokens []string) {
	match = false
	re := regexp.MustCompile(`\${([A-Za-z0-9:*_\|\-\.\?]+)}`)
	tokens = re.FindAllString(strings.Replace(value, "$${", "X{", -1), -1)
	if len(tokens) > 0 {
		match = true
	}
	return match, tokens
}

// helper to provide default value
func orDefault(value, def string) string {
	if strings.TrimSpace(value) == "" {
		return def
	}
	return value
}

// strip variable demarcations
func strip(value string) (stripped string) {
	stripped = strings.Replace(value, "${", "", -1)
	stripped = strings.Replace(stripped, "}", "", -1)
	return stripped
}

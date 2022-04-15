package utils

import (
	"regexp"
	"strings"
)

// ToAddress is a utility function to trim the GRPC protocol prefix as it is not needed by GO
// if the GRPC protocol is not found, the url is returned unchanged
func ToAddress(url string) string {
	if strings.HasPrefix(url, "grpc://") {
		return strings.TrimPrefix(url, "grpc://")
	}
	if strings.HasPrefix(url, "grpcs://") {
		return strings.TrimPrefix(url, "grpcs://")
	}
	return url
}

//AttemptSecured is a utility function which verifies URL and returns if secured connections needs to established
// for protocol 'grpcs' in URL returns true
// for protocol 'grpc' in URL returns false
// for no protocol mentioned, returns !allowInSecure
func AttemptSecured(url string, allowInSecure bool) bool {
	ok, err := regexp.MatchString(".*(?i)s://", url)
	if ok && err == nil {
		return true
	} else if strings.Contains(url, "://") {
		return false
	} else {
		return !allowInSecure
	}
}

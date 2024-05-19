package main

import (
	"fmt"
	"regexp"
)

type request string

func (r request) getVerb() (verb, error) {
	reqString := string(r)
	regex := regexp.MustCompile(`^(GET|POST) .* HTTP/1\.1`)
	match := regex.FindStringSubmatch(reqString)

	if match == nil {
		return "", fmt.Errorf("verb not found")
	}
	switch match[1] {
	case "GET":
		return get, nil
	case "POST":
		return post, nil
	default:
		return "", fmt.Errorf("verb %s not recognized", match[1])
	}
}

func (r request) getTarget() (string, error) {
	reqString := string(r)
	regex := regexp.MustCompile(`^(GET|POST) (.*) HTTP/1\.1`)
	match := regex.FindStringSubmatch(reqString)

	if match == nil {
		return "", fmt.Errorf("target not found")
	}
	return match[2], nil
}

func (r request) getHeader(header string) (string, error) {
	reqString := string(r)
	regex, regexError := regexp.Compile(fmt.Sprintf(`(?i)%s: (\S+)`, header))

	if regexError != nil {
		panic(fmt.Sprintf("something is wrong with header: %s", regexError))
	}

	match := regex.FindStringSubmatch(reqString)

	if match == nil {
		return "", fmt.Errorf("header not found")
	}

	return match[1], nil
}

func (r request) getContentEncoding() string {
	acceptedFormats := []string{
		"gzip",
		"Brotli",
	}
	encoding, _ := r.getHeader("Accept-Encoding")

	for _, v := range acceptedFormats {
		if v == encoding {
			return encoding
		}
	}

	return ""
}

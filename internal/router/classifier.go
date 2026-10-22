package router

import (
	"regexp"
	"strings"
)

type ContentType string

const (
	TypeCensored   ContentType = "censored"
	TypeAmateur    ContentType = "amateur"
	TypeUncensored ContentType = "uncensored"
	TypeFC2        ContentType = "fc2"
	TypeOther      ContentType = "other"
)

var (
	fc2Regex      = regexp.MustCompile(`(?i)^(?:FC2[-_]?(?:PPV[-_]?)?|FC-)\d+$`)
	amateurPrefix = regexp.MustCompile(`(?i)^(?:SIRO|STP|STN|MIUM|LUXU|GANA|ARA)[-_]?\d+$|^\d{3}[A-Z]+[-_]?\d+$`)
	uncensoredReg = regexp.MustCompile(`(?i)^\d{6}[-_]\d{3}$|^HEYZO[-_]?\d+$|^n\d{4}$`)
	standardCode  = regexp.MustCompile(`(?i)^[A-Z]{2,6}[-_]?\d{2,6}$`)
)

// ClassifyNumber determines the content category based on the release code pattern.
func ClassifyNumber(code string) ContentType {
	clean := strings.TrimSpace(code)
	if fc2Regex.MatchString(clean) {
		return TypeFC2
	}
	if amateurPrefix.MatchString(clean) {
		return TypeAmateur
	}
	if uncensoredReg.MatchString(clean) {
		return TypeUncensored
	}
	if standardCode.MatchString(clean) {
		return TypeCensored
	}
	return TypeCensored
}

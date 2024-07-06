package storage

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// ContextKey creates an org:${orgID}:context:${contextID} key
func ContextKey(orgID string, contextID string) string {
	return fmt.Sprintf("org:%s:context:%s", orgID, contextID)
}

// RuleKey creates an org:${orgID}:context:${contextID}:rule:${ruleID} key
func RuleKey(orgID, contextID, ruleID string) string {
	return fmt.Sprintf("org:%s:context:%s:rule:%s", orgID, contextID, ruleID)
}

// VersionKey creates an org:${orgID}:context:${contextID}:rule:${ruleID}:v${version} key
func VersionKey(orgID, contextID, ruleID string, version uint64) string {
	return fmt.Sprintf("org:%s:context:%s:rule:%s:v%d", orgID, contextID, ruleID, version)
}

var vSuffix = regexp.MustCompile(".*:v\\d+$")
var numbers = regexp.MustCompile("[0-9]+")

func IsVersionKey(key string) bool {
	return strings.HasPrefix(key, "org:") && vSuffix.MatchString(key)
}

func VersionFromVersionKey(key string) uint64 {
	vSfx := vSuffix.FindString(key)
	vStr := numbers.FindString(vSfx)
	if s, err := strconv.ParseUint(vStr, 10, 64); err == nil {
		return s
	}
	return 0
}

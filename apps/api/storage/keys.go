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

func SplitContextKey(key string) (orgID, contextID string) {
	parts := strings.Split(key, ":")
	return parts[2], parts[3]
}

// ContextSchemaKey creates an org:${orgID}:context:${contextID}:schema key
func ContextSchemaKey(orgID string, contextID string) string {
	return fmt.Sprintf("org:%s:context:%s:schema", orgID, contextID)
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

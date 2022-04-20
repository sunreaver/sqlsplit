package sqlsplit

import "strings"

func SQLType(raw string) string {
	raw = strings.ToUpper(raw)
	if strings.HasPrefix(raw, "AUDIT") ||
		strings.HasPrefix(raw, "COMMENT") ||
		strings.HasPrefix(raw, "CONNECT") ||
		strings.HasPrefix(raw, "DISCONNECT") ||
		strings.HasPrefix(raw, "EXIT") ||
		strings.HasPrefix(raw, "GRANT") ||
		strings.HasPrefix(raw, "NOAUDIT") ||
		strings.HasPrefix(raw, "QUIT") ||
		strings.HasPrefix(raw, "REVOKE") ||
		strings.HasPrefix(raw, "RENAME") {
		return "DCL"
	} else if strings.HasPrefix(raw, "DROP") ||
		strings.HasPrefix(raw, "ALTER") ||
		strings.HasPrefix(raw, "TRUNCATE") ||
		strings.HasPrefix(raw, "CREATE") {
		return "DDL"
	} else if strings.HasPrefix(raw, "INSERT") ||
		strings.HasPrefix(raw, "UPDATE") ||
		strings.HasPrefix(raw, "DELETE") {
		return "DML"
	}
	return "DQL"
}

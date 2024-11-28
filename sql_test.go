package sqlsplit

import (
	"testing"
)

func TestSQLType(t *testing.T) {
	testCases := []struct {
		input string
		want  SQLTYPE
	}{
		{"CREATE   USER abc", DCL},
		{"DROP   USER", DCL},
		{"ALTER    USER", DCL},
		{"ALTER  abc  USER", DDL},
		{"CREATE    TABLE", DDL},
		{"DROP TABLE", DDL},
		{"ALTER TABLE", DDL},
		{"INSERT INTO", DML},
		{"UPDATE TABLE", DML},
		{"DELETE FROM", DML},
		{"COMMIT", TTL},
		{"ROLLBACK", TTL},
		{"SAVEPOINT", TTL},
		{"SELECT * FROM", DQL},
		{"SET FOREIGN_KEY_CHECKS=0", DCL},
		{"RENAME TABLE", DCL},
		{"CREATE INDEX", DDL},
		{"DROP INDEX", DDL},
		{"ALTER INDEX", DDL},
		{"CREATE ROLE", DCL},
		{"DROP ROLE", DCL},
		{"ALTER ROLE", DCL},
		{"CREATE USER", DCL},
		{"DROP USER", DCL},
		{"ALTER USER", DCL},
		{"CREATE EXTENSION", DDL},
		{"DROP EXTENSION", DDL},
		{"ALTER EXTENSION", DDL},
		{"savepoint abc", TTL},
		{"reindex abc", DDL},
		{"close   abc", DDL},
		{"shutdown c", DCL},
		{"comment on table", DDL},
		{"comment table", DDL},
		{"comment on user", DCL},
		{"comment on policy", DCL},
		{"create policy", DCL},
		{"create database", DCL},
		{"create table", DDL},
	}

	for _, v := range testCases {
		tp := SQLType(v.input)
		if tp != v.want {
			t.Errorf("SQLType(%s) = %v, want %v", v.input, tp, v.want)
		}
	}
}

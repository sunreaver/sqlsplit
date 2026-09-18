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
		// 带前导注释的测试
		{"/* block comment */ CREATE TABLE t1 (id int)", DDL},
		{"-- line comment\nSELECT 1", DQL},
		{"# mysql comment\nUPDATE t1 SET col = 1", DML},
		// 复核 SQL
		{"SELECT * FROM accounts WHERE id = 1 FOR UPDATE", DQL},
		{"SELECT * FROM accounts WHERE id = 1 LOCK IN SHARE MODE", DQL},
		{"INSERT INTO vip_users SELECT * FROM users WHERE score > 90", DML},
		// 视图与存储过程
		{"CREATE VIEW v_users AS SELECT * FROM users", DDL},
		{"CREATE OR REPLACE FORCE VIEW v_f AS SELECT 1 FROM dual", DDL},
		{"CREATE PROCEDURE p1() BEGIN SELECT 1; END", DDL},
		// CTE
		{"WITH cte AS (SELECT 1 AS val) SELECT * FROM cte", DQL},
		{"WITH cte AS (SELECT 1 AS val) INSERT INTO t SELECT * FROM cte", DML},
		{"WITH cte AS (SELECT 1 AS val) UPDATE t SET val = 1", DML},
		{"WITH cte AS (SELECT 1 AS val) DELETE FROM t WHERE val = 1", DML},
		// MERGE
		{"MERGE INTO t USING s ON (t.id = s.id) WHEN MATCHED THEN UPDATE SET t.v = s.v", DML},
	}

	for _, v := range testCases {
		tp := SQLType(v.input)
		if tp != v.want {
			t.Errorf("SQLType(%s) = %v, want %v", v.input, tp, v.want)
		}
	}
}

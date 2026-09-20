-- ============================================================================
-- SQL 切分示例文件（包含常见、复核、复杂过程、视图及多方言特性）
-- ============================================================================

-- 前置多余分号与空分号压力测试（将被解析引擎自动忽略，不生成空条目）
;
;;

-- ----------------------------------------------------------------------------
-- 1. 基础 DML / DQL（通用增删改查）
-- ----------------------------------------------------------------------------

-- 查询活跃用户
SELECT id, username, email, status
FROM users
WHERE status = 'ACTIVE' AND deleted = 0;

-- 批量插入用户数据
INSERT INTO users (id, username, email, score, created_at)
VALUES 
  (1, 'alice', 'alice@example.com', 100, NOW()),
  (2, 'bob', 'bob@example.com', 85, NOW()),
  (3, 'charlie', 'charlie@example.com', 95, NOW());

-- 更新用户积分与状态
UPDATE users
SET score = score + 10, updated_at = NOW()
WHERE score < 90 AND status = 'ACTIVE';

-- 删除无用或失效会话
DELETE FROM user_sessions
WHERE expire_time < NOW() OR is_revoked = 1;

-- ----------------------------------------------------------------------------
-- 2. 复核与审核 SQL（Review & Audit SQL）
-- ----------------------------------------------------------------------------

-- 行级排他锁查询（SELECT FOR UPDATE 复核场景）
SELECT account_id, balance, frozen_amount
FROM bank_accounts
WHERE account_id = 'ACC_8888001'
FOR UPDATE;

-- 带超时的排他锁查询（WAIT / NOWAIT 场景）
SELECT order_id, status, pay_amount
FROM orders
WHERE order_id = 'ORD_20260918_001'
FOR UPDATE NOWAIT;

-- 共享锁查询（LOCK IN SHARE MODE）
SELECT product_id, stock_count
FROM inventory
WHERE product_id = 10086
LOCK IN SHARE MODE;

-- INSERT (SELECT) 复核迁移场景
INSERT INTO vip_customer_archive (customer_id, customer_name, total_asset, archive_date)
SELECT c.id, c.name, a.total_asset, CURRENT_DATE
FROM customers c
INNER JOIN assets a ON c.id = a.customer_id
WHERE a.total_asset >= 1000000;

-- ----------------------------------------------------------------------------
-- 3. 视图（VIEW）与物化视图（DDL 场景）
-- ----------------------------------------------------------------------------

-- 创建基础普通视图
CREATE VIEW v_active_customer_overview AS
SELECT c.id, c.name, c.email, a.total_asset
FROM customers c
JOIN assets a ON c.id = a.customer_id
WHERE c.status = 'ACTIVE';

-- 创建/替换带 WITH CHECK OPTION 的视图
CREATE OR REPLACE VIEW v_high_value_orders AS
SELECT order_id, customer_id, total_amount, created_at
FROM orders
WHERE total_amount > 5000
WITH CHECK OPTION;

-- Oracle 风格 FORCE 视图
CREATE OR REPLACE FORCE VIEW v_pending_transactions AS
SELECT tx_id, tx_type, amount, tx_status
FROM tx_log
WHERE tx_status = 'PENDING';

-- ----------------------------------------------------------------------------
-- 4. 复杂查询与 CTE（Common Table Expressions）
-- ----------------------------------------------------------------------------

-- CTE 递归查询（层级部门统计）
WITH RECURSIVE dept_tree AS (
  SELECT id, name, parent_id, 1 AS level
  FROM departments
  WHERE parent_id IS NULL
  UNION ALL
  SELECT d.id, d.name, d.parent_id, dt.level + 1
  FROM departments d
  JOIN dept_tree dt ON d.parent_id = dt.id
)
SELECT id, name, level FROM dept_tree ORDER BY level, id;

-- CTE 配合 INSERT 场景
WITH new_ranks AS (
  SELECT user_id, RANK() OVER (ORDER BY score DESC) as rank_val
  FROM users
)
INSERT INTO user_rank_snapshot (user_id, rank_val, snapshot_time)
SELECT user_id, rank_val, NOW() FROM new_ranks;

-- MERGE INTO 场景
MERGE INTO customer_targets t
USING (SELECT customer_id, target_level FROM customer_updates) s
ON (t.customer_id = s.customer_id)
WHEN MATCHED THEN
  UPDATE SET t.target_level = s.target_level, t.updated_at = SYSDATE
WHEN NOT MATCHED THEN
  INSERT (customer_id, target_level, created_at)
  VALUES (s.customer_id, s.target_level, SYSDATE);

-- ----------------------------------------------------------------------------
-- 5. Oracle PL/SQL 复杂过程与函数（生产级带声明区、游标、循环与异常处理）
-- ----------------------------------------------------------------------------

create or replace procedure SP_FTP_ACCT_DATA(START_DATE in varchar, --批次日期 yyyy-mm-dd
                                     END_DATE IN VARCHAR,
                                    --i_data in varchar2,
                                    o_sql_state out varchar) as
  /************************************************************************
  脚本名称 ： 存款每日数据导入
  目的     : 将月底的数据改为每日数据插入进表
  作者     ：王洪
  创建日期 ：2021/6/2
  源数据表 ：存储过程每步操作
  目标表   ：etl_log_tbl

  (i_data_date        varchar2, --批次日期 yyyy-mm-dd
                                       i_data     varchar2
                                       ) is
  ************************************************************************/
  -- i_data_date varchar2(20):='111';
  -- i_data  varchar2(20):='222';
  i_data_date varchar2(20);
  --cur_bal NUMBER(24,6);
  i_up_date date;
  i_down_date date;
  i_cur_bal number(24,6);
begin
  execute immediate 'truncate table mspub_model_depositsource';
  commit;

  i_data_date := START_DATE;
  --cur_bal := 150000;

  while i_data_date <= END_DATE
  loop
    i_down_date := trunc(to_date(i_data_date,'yyyy-mm-dd') + 1,'MM') - 1;
    i_up_date := add_months(i_down_date + 1,1) - 1;
    select b.cur_bal + (a.cur_bal-b.cur_bal) / (i_down_date - i_up_date)*(to_date(i_data_date,'yyyy-mm-dd') - i_up_date)  into i_cur_bal
      from test_hqcdl a
      left join test_hqcdl b
        on b.data_dt = to_char(i_up_date,'yyyymmdd')
     where a.data_dt = to_char(i_down_date,'yyyymmdd');

    insert into mspub_model_depositsource
      (CUR_BAL,
       DEP_TYPE,
       DATA_DATE,
       TS,
       DR,
       ACCOUNT_TYPE,
       BRAN_CD,
       CCY_CD
      )
    select
      i_cur_bal,
      '1',
      i_data_date,
      '',
      '0',
      '1',
      '0512001',
      'CNY'
      from DUAL;
    commit;

    i_data_date := TO_CHAR(TO_DATE(i_data_date,'YYYY-MM-DD') + 1,'YYYY-MM-DD');
    --cur_bal := cur_bal - 100;
  end loop;

end;


-- 这是一条没有归属的备注

/***
这是一段没有归属的备注
*****/

create PROCEDURE SP_FTP_DEL_DATA(
                                            P_AS_OF_DATE VARCHAR2
                                           ,RET_MSG      OUT VARCHAR2
                                           ,RET_FLG      OUT VARCHAR2) -- 数据日期 只处理在此数据日期之前的数据
 IS
   /**************************************************************************
    -- 功能描述  ： 根据清理机制配置表T_DEL_EXPIRED_DATA_CFG中定义的删除规则 删除过期数据
    -- 参数描述  ： P_AS_OF_DATE  数据日期,RET_FLG  批次标识,RET_MSG 错误信息
    -- 目标表    ：
    -- 作    者  ： 王洪
    -- 创建日期  ： 2021-06-03
   **************************************************************************/
    V_SQL VARCHAR2(800);
    V_WHERE VARCHAR2(400);
    V_PARTITIONED VARCHAR2(10);
    V_DATE DATE := TO_DATE(P_AS_OF_DATE,'YYYY-MM-DD');
BEGIN

  EXECUTE IMMEDIATE 'TRUNCATE TABLE DEL_PART_TAB';

  INSERT INTO DEL_PART_TAB
  SELECT TABLE_OWNER,TABLE_NAME,PARTITION_NAME
    FROM ALL_TAB_PARTITIONS
   WHERE TABLE_NAME IN (
         SELECT TABLE_NAME FROM T_DEL_EXPIRED_DATA_CFG
         )
     AND TABLE_OWNER IN (
         SELECT DISTINCT OWNER FROM T_DEL_EXPIRED_DATA_CFG
         );

  COMMIT;

  FOR X IN ( SELECT * FROM T_DEL_EXPIRED_DATA_CFG )
  --读入删除表名称
  LOOP
    DBMS_OUTPUT.PUT_LINE('--'||X.OWNER ||'.'||X.TABLE_NAME);
    SELECT PARTITIONED INTO V_PARTITIONED FROM ALL_TABLES WHERE TABLE_NAME = X.TABLE_NAME AND OWNER = X.OWNER;
    --如果需要删除数据的表为分区表，使用ALTER TABLE ..TRUNCATE PARTITION 来删除数据
    IF(V_PARTITIONED ='YES') THEN    --是分区表
        IF X.END_YEAR_DATA IS NULL THEN --年底数据保留时间没有定义
            IF X.END_MONTH_DATA IS NULL THEN    --月底保留时间没有定义
                FOR Y IN (SELECT TABLE_NAME,PARTITION_NAME
                            FROM DEL_PART_TAB
                           WHERE TABLE_OWNER=X.OWNER AND TABLE_NAME = X.TABLE_NAME
                             AND TO_DATE(SUBSTR(PARTITION_NAME,2,8),'YYYYMMDD') < CASE WHEN X.DAY_DATA_MULT='M' THEN ADD_MONTHS(V_DATE,-1*X.DAY_DATA)+1
                                                                                       WHEN X.DAY_DATA_MULT='D' THEN V_DATE - X.DAY_DATA+1 END )
                --查询保留时间前的分区
                LOOP
                    --使用 ALTER TABLE TRUNCATE 删除符合条件的数据数据

                    EXECUTE IMMEDIATE ' ALTER TABLE ' ||X.OWNER || '.' || Y.TABLE_NAME || ' TRUNCATE PARTITION '|| Y.PARTITION_NAME ;

                END LOOP;
            ELSE --定义了月底保留时间
                FOR Y IN (SELECT TABLE_NAME,PARTITION_NAME FROM DEL_PART_TAB
                           WHERE TABLE_OWNER=X.OWNER AND TABLE_NAME = X.TABLE_NAME
                             AND TO_DATE(SUBSTR(PARTITION_NAME,2,8),'YYYYMMDD') < ADD_MONTHS(V_DATE,-1*X.END_MONTH_DATA)+1
                             AND TO_CHAR(TO_DATE(SUBSTR(PARTITION_NAME,2,8),'YYYYMMDD')+1,'DD')='01')
                --查询月底保留时间前的所有月底数据
                LOOP
                    --使用 ALTER TABLE TRUNCATE 删除符合条件的数据数据

                    EXECUTE IMMEDIATE ' ALTER TABLE ' ||X.OWNER || '.' || Y.TABLE_NAME || ' TRUNCATE PARTITION '|| Y.PARTITION_NAME;
                END LOOP;
                FOR Y IN (SELECT TABLE_NAME,PARTITION_NAME FROM DEL_PART_TAB
                           WHERE TABLE_OWNER=X.OWNER AND TABLE_NAME = X.TABLE_NAME
                             AND TO_DATE(SUBSTR(PARTITION_NAME,2,8),'YYYYMMDD') < CASE WHEN X.DAY_DATA_MULT='M' THEN ADD_MONTHS(V_DATE,-1*X.DAY_DATA)+1
                                                                                       WHEN X.DAY_DATA_MULT='D' THEN V_DATE-X.DAY_DATA+1 END
                             AND TO_CHAR(TO_DATE(SUBSTR(PARTITION_NAME,2,8),'YYYYMMDD')+1,'DD')<>'01')
                --查询保留日前的非月底数据
                LOOP
                    --使用 ALTER TABLE TRUNCATE 删除符合条件的数据数据

                    EXECUTE IMMEDIATE ' ALTER TABLE ' ||X.OWNER || '.' || Y.TABLE_NAME || ' TRUNCATE PARTITION '|| Y.PARTITION_NAME;
                END LOOP;
            END IF;
        ELSE --定义了年底数据的保留时间
             --删除年底数据
            FOR Y IN (SELECT TABLE_NAME,PARTITION_NAME FROM DEL_PART_TAB
                       WHERE TABLE_OWNER=X.OWNER AND TABLE_NAME = X.TABLE_NAME
                         AND TO_DATE(SUBSTR(PARTITION_NAME,2,8),'YYYYMMDD') < ADD_MONTHS(V_DATE,-1*X.END_YEAR_DATA)+1
                         AND TO_CHAR(TO_DATE(SUBSTR(PARTITION_NAME,2,8),'YYYYMMDD')+1,'MMDD')='0101')
            --查询您保留年底数据自谦的年底数据
            LOOP
                --使用 ALTER TABLE TRUNCATE 删除符合条件的数据数据

                EXECUTE IMMEDIATE ' ALTER TABLE ' ||X.OWNER || '.' || Y.TABLE_NAME || ' TRUNCATE PARTITION '|| Y.PARTITION_NAME;
            END LOOP;

            IF X.END_MONTH_DATA IS NULL THEN -- 没有定义了保留月末数据时间
                FOR Y IN (SELECT TABLE_NAME,PARTITION_NAME FROM DEL_PART_TAB
                           WHERE TABLE_OWNER=X.OWNER AND TABLE_NAME = X.TABLE_NAME
                             AND TO_DATE(SUBSTR(PARTITION_NAME,2,8),'YYYYMMDD') < CASE WHEN X.DAY_DATA_MULT='M' THEN ADD_MONTHS(V_DATE,-1*X.DAY_DATA)+1
                                                                                       WHEN X.DAY_DATA_MULT='D' THEN V_DATE-X.DAY_DATA+1 END
                             AND TO_CHAR(TO_DATE(SUBSTR(PARTITION_NAME,2,8),'YYYYMMDD')+1,'MMDD')<>'0101')
                --查询日保留数据前的非年底数据
                LOOP
                    --使用 ALTER TABLE TRUNCATE 删除符合条件的数据数据

                   EXECUTE IMMEDIATE ' ALTER TABLE ' ||X.OWNER || '.' || Y.TABLE_NAME || ' TRUNCATE PARTITION '|| Y.PARTITION_NAME;
                END LOOP;
            ELSE --定义了保留月末数据期
                --查询月底保留数据前的非年底数据
                FOR Y IN (SELECT TABLE_NAME,PARTITION_NAME FROM DEL_PART_TAB
                           WHERE TABLE_OWNER=X.OWNER AND TABLE_NAME = X.TABLE_NAME
                             AND TO_DATE(SUBSTR(PARTITION_NAME,2,8),'YYYYMMDD') < ADD_MONTHS(V_DATE,-1*X.END_MONTH_DATA)+1
                             AND TO_CHAR(TO_DATE(SUBSTR(PARTITION_NAME,2,8),'YYYYMMDD')+1,'DD')='01'
                             AND TO_CHAR(TO_DATE(SUBSTR(PARTITION_NAME,2,8),'YYYYMMDD')+1,'MMDD')<>'0101')
                LOOP
                    --使用 ALTER TABLE TRUNCATE 删除符合条件的数据数据

                   EXECUTE IMMEDIATE ' ALTER TABLE ' ||X.OWNER || '.' || Y.TABLE_NAME || ' TRUNCATE PARTITION '|| Y.PARTITION_NAME;
                END LOOP;

                FOR Y IN (SELECT TABLE_NAME,PARTITION_NAME FROM DEL_PART_TAB
                           WHERE TABLE_OWNER=X.OWNER AND TABLE_NAME = X.TABLE_NAME
                             AND TO_DATE(SUBSTR(PARTITION_NAME,2,8),'YYYYMMDD') < CASE WHEN X.DAY_DATA_MULT='M' THEN ADD_MONTHS(V_DATE,-1*X.DAY_DATA)+1
                                                                                       WHEN X.DAY_DATA_MULT='D' THEN V_DATE-X.DAY_DATA+1 END
                             AND TO_CHAR(TO_DATE(SUBSTR(PARTITION_NAME,2,8),'YYYYMMDD')+1,'DD')<>'01')
                -- 查询日保留日期前的非月末数据
                LOOP
                    --使用 ALTER TABLE TRUNCATE 删除符合条件的数据数据

                    EXECUTE IMMEDIATE ' ALTER TABLE ' ||X.OWNER || '.' || Y.TABLE_NAME || ' TRUNCATE PARTITION '|| Y.PARTITION_NAME;
                END LOOP;
            END IF;
        END IF;
    ELSIF V_PARTITIONED='NO' THEN   --非分区表
       --如果需要删除数据的表不为分区表，使用DELETE FROM 来删除数据
       V_SQL :=' DELETE FROM ' ||X.OWNER||'.'||X.TABLE_NAME ||' WHERE ';
       IF X.DAY_DATA_MULT ='M' THEN --保留日单位为月
           --过滤数据日期 小于 保留日前
           CASE
                WHEN X.COMMENTS = '1' THEN
                  V_WHERE := '( DATA_DATE < ADD_MONTHS('||'TO_DATE('||P_AS_OF_DATE||',''YYYY-MM-DD'''||')'||',-'||X.DAY_DATA||')+1 ';
                WHEN X.COMMENTS = '2' THEN
                  V_WHERE := '( DATA_DT < ADD_MONTHS('||'TO_DATE('||P_AS_OF_DATE||',''YYYY-MM-DD'''||')'||',-'||X.DAY_DATA||')+1 ';
                ELSE
                 V_WHERE := '( AS_OF_DATE < ADD_MONTHS('||'TO_DATE('||P_AS_OF_DATE||',''YYYYMMDD'''||')'||',-'||X.DAY_DATA||')+1 ';
           END  CASE;
       ELSIF X.DAY_DATA_MULT = 'D' THEN --保留日单位为日
           --过滤数据日期 小于 保留日前
           V_WHERE := '( DATA_DT < '||'TO_DATE('||P_AS_OF_DATE||',''YYYY-MM-DD'''||')'||'-'||X.DAY_DATA||'+1 ';
       END IF;

       IF X.END_MONTH_DATA IS NOT NULL THEN --定义了保留月末
            --保留数据日期为月底的数据，过滤保留月前的月数据
            V_WHERE := V_WHERE ||' AND TO_CHAR(DATA_DT+1,''DD'')<>''01'') OR ( DATA_DT < ADD_MONTHS('||'TO_DATE('
                               ||P_AS_OF_DATE||',''YYYY-MM-DD'''||')'||',-'||X.END_MONTH_DATA||')+1';
       ELSIF X.END_YEAR_DATA IS NULL THEN --没有定义年数据
            --结束
            V_WHERE := V_WHERE ||')';
       END IF;

       IF X.END_YEAR_DATA IS NOT NULL THEN --定义了年底保留时间
            #保留数据日期为年底的数据，过滤保留年前的月数据
            V_WHERE := V_WHERE ||' AND TO_CHAR(DATA_DT+1,''MMDD'')<>''0101'') OR  (DATA_DT < ADD_MONTHS('||'TO_DATE('
                               ||P_AS_OF_DATE||',''YYYY-MM-DD'''||')'||',-'||X.END_YEAR_DATA||')+1)';
       ELSIF X.END_MONTH_DATA IS NOT NULL THEN --定义了月数据
            --结束
            V_WHERE := V_WHERE ||')';
       END IF;
       V_SQL := V_SQL || V_WHERE;
       DBMS_OUTPUT.PUT_LINE(REPLACE(V_SQL,'DELETE','SELECT * ')||';');

       EXECUTE IMMEDIATE V_SQL;
       COMMIT;
    END IF;
  END LOOP;

  RET_FLG := '0';
  RET_MSG := 'SUCCESSFUL';

EXCEPTION
  # 异常
  WHEN OTHERS THEN

    RET_FLG := '1';
    RET_MSG := SQLERRM;

END;

-- Oracle 独立行 / 结束符匿名块测试
DECLARE
  v_test_cnt NUMBER := 0;
BEGIN
  SELECT COUNT(*) INTO v_test_cnt FROM users;
  DBMS_OUTPUT.PUT_LINE('User count is: ' || v_test_cnt);
END;
/

-- Oracle 行级触发器
CREATE OR REPLACE TRIGGER trg_user_audit
AFTER INSERT OR UPDATE ON users
FOR EACH ROW
BEGIN
  INSERT INTO user_audit_log (user_id, op_time)
  VALUES (:NEW.id, SYSDATE);
END;
/

-- ----------------------------------------------------------------------------
-- 6. MySQL 方言特性（反引号包含分号、DELIMITER 自定义分隔符、# 注释、\ 转义）
-- ----------------------------------------------------------------------------

-- 反引号包含分号和特殊字符的字段/表名
SELECT `user;id`, `user;name`, `role;code`
FROM `corp;db`.`emp;table`
WHERE `status;flag` = 1;

-- DELIMITER 自定义分隔符存储过程
DELIMITER //
CREATE PROCEDURE sp_mysql_calc_salary(IN p_emp_id INT, OUT p_bonus DECIMAL(10,2))
BEGIN
  DECLARE v_base DECIMAL(10,2) DEFAULT 0.00;
  SELECT salary INTO v_base FROM employees WHERE id = p_emp_id;
  IF v_base > 10000 THEN
    SET p_bonus = v_base * 0.20;
  ELSE
    SET p_bonus = v_base * 0.10;
  END IF;
END //
DELIMITER ;

-- MySQL # 注释与字符串内分号
# 这是MySQL特定行注释
SELECT * FROM users WHERE note = 'hello;world#test';

-- 字符串转义字符测试
select * from a where a.c = 'abc\'bcd';
select * from a where a.c = 'abc\ \'bcd';

-- ----------------------------------------------------------------------------
-- 7. PostgreSQL 方言特性（$$ 与 $tag$ 代码块引用）
-- ----------------------------------------------------------------------------

-- PostgreSQL $$ 引用 PL/pgSQL 函数
CREATE OR REPLACE FUNCTION get_user_full_name(first_name varchar, last_name varchar)
RETURNS varchar AS $$
BEGIN
  IF first_name IS NULL THEN
    RETURN last_name;
  END IF;
  RETURN first_name || ' ' || last_name;
END;
$$ LANGUAGE plpgsql;

-- PostgreSQL $tag$ 命名 Dollar-quote 引用函数
CREATE OR REPLACE FUNCTION calculate_tax(amount numeric)
RETURNS numeric AS $tax_calc$
DECLARE
  rate numeric := 0.06;
BEGIN
  RETURN amount * rate;
END;
$tax_calc$ LANGUAGE plpgsql;

-- ----------------------------------------------------------------------------
-- 8. 事务控制（TTL）与权限管理（DCL）
-- ----------------------------------------------------------------------------

COMMIT;
ROLLBACK;
SAVEPOINT point_insert_ok;

GRANT SELECT, INSERT, UPDATE ON users TO app_rw_user;
REVOKE DELETE ON users FROM app_rw_user;
CREATE USER reporter_user IDENTIFIED BY 'Pass#2026';
ALTER USER reporter_user ACCOUNT LOCK;

-- ----------------------------------------------------------------------------
-- 9. 连续空分号与边界压力测试
-- ----------------------------------------------------------------------------

SELECT * from users;;;
;SELECT * from users;;;
;;SELECT * from users;;;
;;;SELECT * from users;; ;
; ;SELECT * from users; ; ;
; ;SELECT * from users where a = 'cbd';
; ;SELECT * from users where a = ';;';  -- abc
;SELECT * from users where a = ';;';;;
;SELECT * from users where a = 'aa;;';
SELECT * from users where a = ';;bb张三' or b = '李四;;';
TRUNCATE TABLE users;
TRUNCATE TABLE target_users;
-- ----------------------------------------------------------------------------
-- 10. 高级边界与高危语法场景（P0/P1 测试用例及业务复杂 PL/SQL 块）
-- ----------------------------------------------------------------------------

-- 惠交贷添加默认菜单（复杂匿名块包含循环与批量插入）
BEGIN
    -- 第二步：为每个符合条件的用户插入权限目标和配置
    FOR user_rec IN (
        SELECT DISTINCT a.userseq, a.cifseq
          FROM euser a
         WHERE a.userstate <> '1' 
           AND EXISTS (SELECT 1
                         FROM euserproduct b
                        WHERE b.userseq = a.userseq
                          AND b.prdid = 'ELoanToPay')
           AND NOT EXISTS (SELECT 1
                             FROM euserproduct c
                            WHERE c.userseq = a.userseq
                              AND c.prdid = 'GeneralFinancingCredit')
    ) LOOP
        -- 插入权限目标表
        INSERT INTO eauthtarget (AUTHTARGETSEQ, CIFSEQ, CURRENCY, ACSEQ, PRDID, MINAMOUNT, MAXAMOUNT)
        VALUES (eAuthTargetSeq.nextval, user_rec.cifseq, 'CNY', null, 'GeneralFinancingCredit', 0.00, 9999999999.00);
        
        -- 插入权限配置表（使用currval获取刚生成的序列值）
        INSERT INTO eauthtargetauthcfg (AUTHTARGETSEQ, AUTHSTEP, USERGRPID, AUTHUSERCOUNT, USERSEQ)
        VALUES (eAuthTargetSeq.currval, 1, '1', 1, null);
    END LOOP;
    
    -- 第一步：插入euserproduct数据
    INSERT INTO euserproduct (userseq, bankseq, prdgrpid, prdid)
    SELECT a.userseq,
           1 AS bankseq,
           'equery' AS prdgrpid,
           'GeneralFinancingCredit' AS prdid
      FROM euser a
     WHERE a.userstate <> '1' 
       AND EXISTS (SELECT 1
                     FROM euserproduct b
                    WHERE b.userseq = a.userseq
                      AND b.prdid = 'ELoanToPay')
       AND NOT EXISTS (SELECT 1
                         FROM euserproduct c
                        WHERE c.userseq = a.userseq
                          AND c.prdid = 'GeneralFinancingCredit');
    COMMIT;
    DBMS_OUTPUT.PUT_LINE('Batch insert completed successfully');
END;

-- SEC-01: PostgreSQL / ANSI 单引号函数体定义
CREATE FUNCTION add_num(a INT, b INT) RETURNS INT AS 'SELECT a + b;' LANGUAGE sql;

-- SEC-02: Oracle PACKAGE BODY 复杂包体（包含内部子过程）
CREATE OR REPLACE PACKAGE BODY emp_mgmt AS
  PROCEDURE hire_emp(id NUMBER) IS
  BEGIN
    INSERT INTO emp VALUES(id);
  END hire_emp;

  PROCEDURE fire_emp(id NUMBER) IS
  BEGIN
    DELETE FROM emp WHERE emp_id = id;
  END fire_emp;
END emp_mgmt;

-- SEC-03: 独立行除号算术表达式
SELECT
  amount
  /
  quantity
FROM orders;

-- SEC-04: PostgreSQL 自定义枚举类型声明
CREATE TYPE mood AS ENUM ('sad', 'ok', 'happy');

-- SEC-05: 反斜杠路径字符串
SELECT 'C:\dir\' FROM dual;

-- 文件末尾孤立注释（不应生成空假 SQL）
-- 脚本执行结束
/* 审计记录留存 */

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

create     procedure SP_FTP_ACCT_DATA(START_DATE in varchar, --批次日期 yyyy-mm-dd
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

select * from users where a='';

select * ---abc
from a
/***
select * from a where abc=1;
*/
where abc=1;


create or    replace procedure SP_FTP_ACCT_DATA(START_DATE in varchar, --批次日期 yyyy-mm-dd
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

--end;
end;

/***asb***/
--- end

'abg' select * from add;

select * from abc;


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
   /*第1次修改记录
    -- 修改人    ：
    -- 修改目的  ：
    -- 修改内容  ：
    -- 修改时间  ：
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
    IF V_PARTITIONED ='YES' THEN    --是分区表
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
            --保留数据日期为年底的数据，过滤保留年前的月数据
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
  --异常
  WHEN OTHERS THEN

    RET_FLG := '1';
    RET_MSG := SQLERRM;

END;

SELECT * from uses;
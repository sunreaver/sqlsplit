# 切分sql文件中的sql语句

1. 需保证各条sql语法正确.

2. 切分后，会返回各条独立的sql语句.

## 使用约束

- **编码要求**：输入必须为 **UTF-8** 编码文本，不支持 GBK/GB18030 等其他编码。
- **大小限制**：输入大小不应超过 **100MB**（`MaxInputSize`），超出将返回 nil。

## 多方言支持

- **Oracle**：PL/SQL 过程、声明区分号保护、`/` 独立行结束符
- **MySQL**：反引号、`#` 注释、`DELIMITER` 切换
- **PostgreSQL**：`$$` 引用块（Dollar-quote）

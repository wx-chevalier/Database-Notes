# SQL 解析引擎

![uber queryparser 流程](https://s1.ax1x.com/2020/03/28/Gk0luT.md.png)

- Phase 1: Parse. Transforms the query from a raw string of characters into an abstract syntax tree (AST) representation.

- Phase 2: Resolve. Scans the raw AST and applies scoping rules. Transforms plain column names by adding the table name, and transforms plain table names by adding the schema name. Requires as input the full list of columns in every table and the full list of tables in every schema, otherwise known as “catalog information.”

- Phase 3: Analyze. Scans the resolved AST, looking for columns which are compared for equality.

# TBD

- https://marianogappa.github.io/software/2019/06/05/lets-build-a-sql-parser-in-go/

---
title: "db2i-execute-sql"
type: docs
weight: 1
description: > 
  A "db2i-execute-sql" tool executes a SQL statement against a Db2 for i
  database.
aliases:
- /resources/tools/db2i-execute-sql
---

## About

A `db2i-execute-sql` tool executes a SQL statement against a Db2 for i
database. It's compatible with [db2i](../sources/db2i.md) sources.

`db2i-execute-sql` takes one input parameter `sql` and runs the SQL
statement against the `source`.

> **Note:** This tool is intended for developer assistant workflows with
> human-in-the-loop and shouldn't be used for production agents.

## Example

```yaml
tools:
 execute_sql_tool:
    kind: db2i-execute-sql
    source: my-db2i-instance
    description: Use this tool to execute SQL statements against Db2 for i.
```

## Reference

| **field**   |                  **type**                  | **required** | **description**                                                                                  |
|-------------|:------------------------------------------:|:------------:|--------------------------------------------------------------------------------------------------|
| kind        |                   string                   |     true     | Must be "db2i-execute-sql".                                                                     |
| source      |                   string                   |     true     | Name of the source the SQL should execute on.                                                    |
| description |                   string                   |     true     | Description of the tool that is passed to the LLM.                                               |
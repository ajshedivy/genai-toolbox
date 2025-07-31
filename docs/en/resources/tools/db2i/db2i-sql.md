---
title: "db2i-sql"
type: docs
weight: 1
description: > 
  A "db2i-sql" tool executes a pre-defined SQL statement against a Db2 for i
  database.
aliases:
- /resources/tools/db2i-sql
---

## About

A `db2i-sql` tool executes a pre-defined SQL statement against a Db2 for i
database. It's compatible with [db2i](../sources/db2i.md) sources.

The specified SQL statement is executed with parameter binding support,
and specified parameters will be inserted according to their position. If template parameters are included, they will be resolved before execution
of the statement.

## Example

> **Note:** This tool uses parameterized queries to prevent SQL injections.
> Query parameters can be used as substitutes for arbitrary expressions.
> Parameters cannot be used as substitutes for identifiers, column names, table
> names, or other parts of the query.

```yaml
tools:
 search_jobs_by_subsystem:
    kind: db2i-sql
    source: my-db2i-instance
    statement: |
      SELECT JOB_NAME, SUBSYSTEM, CPU_TIME, MEMORY_POOL
      FROM TABLE(QSYS2.ACTIVE_JOB_INFO()) 
      WHERE SUBSYSTEM = ?
      AND JOB_STATUS = ?
      ORDER BY CPU_TIME DESC
      FETCH FIRST 10 ROWS ONLY
    description: |
      Use this tool to find active jobs in a specific subsystem with a given status.
      Takes a subsystem name and job status and returns job information.
      Example subsystem names: QINTER, QBATCH, QUSRWRK
      Example job statuses: *ACTIVE, *JOBQ, *OUTQ
      Example:
      {{
          "subsystem": "QINTER",
          "job_status": "*ACTIVE",
      }}
    parameters:
      - name: subsystem
        type: string
        description: IBM i subsystem name (e.g., QINTER, QBATCH, QUSRWRK)
      - name: job_status
        type: string
        description: Job status filter (e.g., *ACTIVE, *JOBQ, *OUTQ)
```

### Example with Template Parameters

> **Note:** This tool allows direct modifications to the SQL statement,
> including identifiers, column names, and table names. **This makes it more
> vulnerable to SQL injections**. Using basic parameters only (see above) is
> recommended for performance and safety reasons. For more details, please check
> [templateParameters](_index#template-parameters).

```yaml
tools:
 list_table:
    kind: db2i-sql
    source: my-db2i-instance
    statement: |
      SELECT * FROM {{.schema}}.{{.tableName}}
      FETCH FIRST 100 ROWS ONLY
    description: |
      Use this tool to list all information from a specific table in a schema.
      Example:
      {{
          "schema": "MYLIB",
          "tableName": "CUSTOMERS",
      }}
    templateParameters:
      - name: schema
        type: string
        description: IBM i library/schema name
      - name: tableName
        type: string
        description: Table name to select from
```

## Reference

| **field**           |                  **type**                                 | **required** | **description**                                                                                                                            |
|---------------------|:---------------------------------------------------------:|:------------:|--------------------------------------------------------------------------------------------------------------------------------------------|
| kind                |                   string                                  |     true     | Must be "db2i-sql".                                                                                                                       |
| source              |                   string                                  |     true     | Name of the source the SQL should execute on.                                                                                              |
| description         |                   string                                  |     true     | Description of the tool that is passed to the LLM.                                                                                         |
| statement           |                   string                                  |     true     | SQL statement to execute on.                                                                                                               |
| parameters          | [parameters](_index#specifying-parameters)                |    false     | List of [parameters](_index#specifying-parameters) that will be inserted into the SQL statement.                                           |
| templateParameters  |  [templateParameters](_index#template-parameters)         |    false     | List of [templateParameters](_index#template-parameters) that will be inserted into the SQL statement before executing prepared statement. |
---
title: "Db2 for i"
type: docs
weight: 1
description: >
  Db2 for i is IBM's relational database for the IBM i platform.
---

## About

Db2 for i is IBM's relational database for the [IBM i platform](https://www.ibm.com/products/ibm-i) (formerly AS/400). It provides high performance, reliability, and security for business-critical applications. The Db2i source allows you to connect to IBM i systems using the Mapepire database access layer.

## Available Tools

- [`db2i-sql`](../tools/db2i/db2i-sql.md): Executes pre-defined SQL statements against a Db2 for i database.
- [`db2i-execute-sql`](../tools/db2i/db2i-execute-sql.md): Executes SQL statements against a Db2 for i database.

## Requirements

### IBM i System

This source requires an IBM i system with the Mapepire daemon running. You will need:

- IBM i server with Mapepire daemon installed and running, see [Mapepire documentation](https://mapepire-ibmi.github.io/guides/sysadmin/)
- Network connectivity to the IBM i server on the configured port
- Valid user credentials for database access

## Example

```yaml
sources:
    my-db2i-source:
        kind: db2i
        host: ${DB2I_HOST}
        port: 8076
        database: db2i
        user: ${DB2I_USER}
        password: ${DB2I_PASSWORD}
```

{{< notice tip >}}
Use environment variable replacement with the format ${ENV_NAME}
instead of hardcoding your secrets into the configuration file.
{{< /notice >}}

## Reference

| **field** | **type** | **required** | **description**                                                        |
|-----------|:--------:|:------------:|------------------------------------------------------------------------|
| kind      |  string  |     true     | Must be "db2i".                                                        |
| host      |  string  |     true     | IP address or hostname of IBM i server (e.g. "192.168.1.100")         |
| port      |  string  |     true     | Port for Mapepire daemon (e.g. "8076")                               |
| database  |  string  |     true     | Database name on IBM i (typically "db2i")                            |
| user      |  string  |     true     | IBM i user name for authentication                                    |
| password  |  string  |     true     | Password for the IBM i user                                           |
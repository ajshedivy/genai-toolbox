// Copyright 2025 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package db2ilisttables

import (
	"context"
	"fmt"
	"strings"

	mapepire "github.com/deady54/mapepire-go"
	"github.com/goccy/go-yaml"
	"github.com/googleapis/genai-toolbox/internal/sources"
	"github.com/googleapis/genai-toolbox/internal/sources/db2i"
	"github.com/googleapis/genai-toolbox/internal/tools"
)

const kind string = "db2i-list-tables"

const listTablesQuery = `
SELECT TABLE_NAME, TABLE_TYPE, TABLE_TEXT, NUMBER_ROWS, LAST_ALTERED_TIMESTAMP
FROM QSYS2.SYSTABLES
WHERE TABLE_SCHEMA = ?
ORDER BY TABLE_NAME
`

const listColumnsQuery = `
SELECT TABLE_NAME, COLUMN_NAME, DATA_TYPE, LENGTH, NUMERIC_SCALE,
       IS_NULLABLE, COLUMN_DEFAULT, HAS_DEFAULT, COLUMN_TEXT,
       ORDINAL_POSITION
FROM QSYS2.SYSCOLUMNS
WHERE TABLE_SCHEMA = ?
ORDER BY TABLE_NAME, ORDINAL_POSITION
`

func init() {
	if !tools.Register(kind, newConfig) {
		panic(fmt.Sprintf("tool kind %q already registered", kind))
	}
}

func newConfig(ctx context.Context, name string, decoder *yaml.Decoder) (tools.ToolConfig, error) {
	actual := Config{Name: name}
	if err := decoder.DecodeContext(ctx, &actual); err != nil {
		return nil, err
	}
	return actual, nil
}

type compatibleSource interface {
	Db2iPool() *mapepire.JobPool
}

var _ compatibleSource = &db2i.Source{}

var compatibleSources = [...]string{db2i.SourceKind}

type Config struct {
	Name         string   `yaml:"name" validate:"required"`
	Kind         string   `yaml:"kind" validate:"required"`
	Source       string   `yaml:"source" validate:"required"`
	Description  string   `yaml:"description" validate:"required"`
	AuthRequired []string `yaml:"authRequired"`
}

// validate interface
var _ tools.ToolConfig = Config{}

// validate interface
var _ tools.Tool = Tool{}

type Tool struct {
	Name         string           `yaml:"name"`
	Kind         string           `yaml:"kind"`
	AuthRequired []string         `yaml:"authRequired"`
	AllParams    tools.Parameters `yaml:"allParams"`

	Pool        *mapepire.JobPool
	manifest    tools.Manifest
	mcpManifest tools.McpManifest
}

func (cfg Config) ToolConfigKind() string {
	return kind
}

func (cfg Config) Initialize(srcs map[string]sources.Source) (tools.Tool, error) {
	rawS, ok := srcs[cfg.Source]
	if !ok {
		return nil, fmt.Errorf("no source named %q configured", cfg.Source)
	}

	s, ok := rawS.(compatibleSource)
	if !ok {
		return nil, fmt.Errorf("invalid source for %q tool: source kind must be one of %q", kind, compatibleSources)
	}

	allParameters := tools.Parameters{
		tools.NewStringParameter("schema", "The schema (library) name to list tables from (e.g. QIWS, MYLIB)."),
		tools.NewStringParameterWithDefault("table_names", "", "Optional: a comma-separated list of table names to filter. If empty, all tables in the schema are listed."),
	}
	mcpManifest := tools.GetMcpManifest(cfg.Name, cfg.Description, cfg.AuthRequired, allParameters)

	t := Tool{
		Name:         cfg.Name,
		Kind:         kind,
		AuthRequired: cfg.AuthRequired,
		AllParams:    allParameters,
		Pool:         s.Db2iPool(),
		manifest:     tools.Manifest{Description: cfg.Description, Parameters: allParameters.Manifest(), AuthRequired: cfg.AuthRequired},
		mcpManifest:  mcpManifest,
	}
	return t, nil
}

func (t Tool) Invoke(ctx context.Context, params tools.ParamValues, accessToken tools.AccessToken) (any, error) {
	paramsMap := params.AsMap()

	schema, ok := paramsMap["schema"].(string)
	if !ok {
		return nil, fmt.Errorf("invalid 'schema' parameter; expected a string")
	}

	tableNames, _ := paramsMap["table_names"].(string)

	// Build optional table name filter set
	filterSet := make(map[string]bool)
	if tableNames != "" {
		for _, name := range strings.Split(tableNames, ",") {
			name = strings.TrimSpace(name)
			if name != "" {
				filterSet[strings.ToUpper(name)] = true
			}
		}
	}

	// Query tables
	tableOpts := mapepire.QueryOptions{
		Parameters: [][]any{{schema}},
	}
	tableResults, err := t.Pool.ExecuteSQLWithOptions(listTablesQuery, tableOpts)
	if err != nil {
		return nil, fmt.Errorf("unable to execute table query: %w", err)
	}
	if !tableResults.Success {
		return nil, fmt.Errorf("table query failed: %s", tableResults.Error)
	}

	// Query columns
	colOpts := mapepire.QueryOptions{
		Parameters: [][]any{{schema}},
	}
	colResults, err := t.Pool.ExecuteSQLWithOptions(listColumnsQuery, colOpts)
	if err != nil {
		return nil, fmt.Errorf("unable to execute column query: %w", err)
	}
	if !colResults.Success {
		return nil, fmt.Errorf("column query failed: %s", colResults.Error)
	}

	// Group columns by table name
	columnsByTable := make(map[string][]map[string]any)
	for _, row := range colResults.Data {
		tableName, _ := row["TABLE_NAME"].(string)
		if tableName == "" {
			continue
		}
		columnsByTable[tableName] = append(columnsByTable[tableName], row)
	}

	// Build output, applying optional table name filter
	var out []map[string]any
	for _, row := range tableResults.Data {
		tableName, _ := row["TABLE_NAME"].(string)
		if len(filterSet) > 0 && !filterSet[strings.ToUpper(tableName)] {
			continue
		}
		entry := map[string]any{
			"table_name":   row["TABLE_NAME"],
			"table_type":   row["TABLE_TYPE"],
			"table_text":   row["TABLE_TEXT"],
			"number_rows":  row["NUMBER_ROWS"],
			"last_altered": row["LAST_ALTERED_TIMESTAMP"],
			"columns":      columnsByTable[tableName],
		}
		out = append(out, entry)
	}

	return out, nil
}

func (t Tool) ParseParams(data map[string]any, claims map[string]map[string]any) (tools.ParamValues, error) {
	return tools.ParseParams(t.AllParams, data, claims)
}

func (t Tool) Manifest() tools.Manifest {
	return t.manifest
}

func (t Tool) McpManifest() tools.McpManifest {
	return t.mcpManifest
}

func (t Tool) Authorized(verifiedAuthServices []string) bool {
	return tools.IsAuthorized(t.AuthRequired, verifiedAuthServices)
}

func (t Tool) RequiresClientAuthorization() bool {
	return false
}

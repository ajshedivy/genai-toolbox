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

package db2isql_test

import (
	"testing"

	yaml "github.com/goccy/go-yaml"
	"github.com/google/go-cmp/cmp"
	"github.com/googleapis/genai-toolbox/internal/server"
	"github.com/googleapis/genai-toolbox/internal/testutils"
	"github.com/googleapis/genai-toolbox/internal/tools"
	"github.com/googleapis/genai-toolbox/internal/tools/db2i/db2isql"
)

func TestParseFromYamlDb2iSql(t *testing.T) {
	ctx, err := testutils.ContextWithNewLogger()
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	tcs := []struct {
		desc string
		in   string
		want server.ToolConfigs
	}{
		{
			desc: "basic example",
			in: `
			tools:
				example_tool:
					kind: db2i-sql
					source: my-db2i-instance
					description: some description
					statement: |
						SELECT * FROM QIWS.QCUSTCDT;
					authRequired:
						- my-auth-service
					parameters:
						- name: limit
						  type: integer
						  description: max rows to return
			`,
			want: server.ToolConfigs{
				"example_tool": db2isql.Config{
					Name:         "example_tool",
					Kind:         "db2i-sql",
					Source:       "my-db2i-instance",
					Description:  "some description",
					Statement:    "SELECT * FROM QIWS.QCUSTCDT;\n",
					AuthRequired: []string{"my-auth-service"},
					Parameters: tools.Parameters{
						tools.NewIntParameter("limit", "max rows to return"),
					},
				},
			},
		},
	}
	for _, tc := range tcs {
		t.Run(tc.desc, func(t *testing.T) {
			got := struct {
				Tools server.ToolConfigs `yaml:"tools"`
			}{}
			err := yaml.UnmarshalContext(ctx, testutils.FormatYaml(tc.in), &got)
			if err != nil {
				t.Fatalf("unable to unmarshal: %s", err)
			}
			if diff := cmp.Diff(tc.want, got.Tools); diff != "" {
				t.Fatalf("incorrect parse: diff %v", diff)
			}
		})
	}
}

func TestParseFromYamlWithTemplateParamsDb2iSql(t *testing.T) {
	ctx, err := testutils.ContextWithNewLogger()
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	tcs := []struct {
		desc string
		in   string
		want server.ToolConfigs
	}{
		{
			desc: "with template parameters",
			in: `
			tools:
				example_tool:
					kind: db2i-sql
					source: my-db2i-instance
					description: some description
					statement: |
						SELECT * FROM {{.schema}}.{{.table}};
					parameters:
						- name: limit
						  type: integer
						  description: max rows
					templateParameters:
						- name: schema
						  type: string
						  description: The schema to query.
						- name: table
						  type: string
						  description: The table to query.
			`,
			want: server.ToolConfigs{
				"example_tool": db2isql.Config{
					Name:         "example_tool",
					Kind:         "db2i-sql",
					Source:       "my-db2i-instance",
					Description:  "some description",
					Statement:    "SELECT * FROM {{.schema}}.{{.table}};\n",
					AuthRequired: []string{},
					Parameters: tools.Parameters{
						tools.NewIntParameter("limit", "max rows"),
					},
					TemplateParameters: tools.Parameters{
						tools.NewStringParameter("schema", "The schema to query."),
						tools.NewStringParameter("table", "The table to query."),
					},
				},
			},
		},
	}
	for _, tc := range tcs {
		t.Run(tc.desc, func(t *testing.T) {
			got := struct {
				Tools server.ToolConfigs `yaml:"tools"`
			}{}
			err := yaml.UnmarshalContext(ctx, testutils.FormatYaml(tc.in), &got)
			if err != nil {
				t.Fatalf("unable to unmarshal: %s", err)
			}
			if diff := cmp.Diff(tc.want, got.Tools); diff != "" {
				t.Fatalf("incorrect parse: diff %v", diff)
			}
		})
	}
}

/*
Copyright 2024 The Kubernetes Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package types

type CukeComment struct {
	Value string `json:"value"`
	Line  int    `json:"line"`
}

type CukeDocstring struct {
	Value       string `json:"value"`
	ContentType string `json:"content_type"`
	Line        int    `json:"line"`
}

type CukeTag struct {
	Name string `json:"name"`
	Line int    `json:"line"`
}

type CukeResult struct {
	Status   string `json:"status"`
	Error    string `json:"error_message,omitempty"`
	Duration *int   `json:"duration,omitempty"`
}

type CukeMatch struct {
	Location string `json:"location"`
}

type CukeStep struct {
	Keyword   string              `json:"keyword"`
	Name      string              `json:"name"`
	Line      int                 `json:"line"`
	Docstring *CukeDocstring      `json:"doc_string,omitempty"`
	Match     CukeMatch           `json:"match"`
	Result    CukeResult          `json:"result"`
	DataTable []*CukeDataTableRow `json:"rows,omitempty"`
}

type CukeDataTableRow struct {
	Cells []string `json:"cells"`
}

type CukeElement struct {
	ID          string     `json:"id"`
	Keyword     string     `json:"keyword"`
	Name        string     `json:"name"`
	Description string     `json:"description"`
	Line        int        `json:"line"`
	Type        string     `json:"type"`
	Tags        []CukeTag  `json:"tags,omitempty"`
	Steps       []CukeStep `json:"steps,omitempty"`
}

// CukeFeatureJSON ...
type CukeFeatureJSON struct {
	URI         string        `json:"uri"`
	ID          string        `json:"id"`
	Keyword     string        `json:"keyword"`
	Name        string        `json:"name"`
	Description string        `json:"description"`
	Line        int           `json:"line"`
	Comments    []CukeComment `json:"comments,omitempty"`
	Tags        []CukeTag     `json:"tags,omitempty"`
	Elements    []CukeElement `json:"elements,omitempty"`
}

type Results struct {
	Total            int64
	Passed           int64
	Failed           int64
	Variants         []string
	ResultsByVariant []VariantResults
}

type VariantResults struct {
	Name        string
	Total       int64
	Passed      int64
	Failed      int64
	FailedTests []FailedTest
}

type FailedTest struct {
	Scenario   string
	FailedStep string
	Feature    string
	Source     string
}

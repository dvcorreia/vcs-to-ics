// Copyright (c) Diogo Correia
// SPDX-License-Identifier: MIT

package vcstoics_test

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	vcstoics "github.com/dvcorreia/vcs-to-ics"
)

func TestConvert(t *testing.T) {
	const email = "dv_correia@hotmail.com"

	goldenFilesDir := "testdata"
	vcsDir := filepath.Join(goldenFilesDir, "vcs")
	icsDir := filepath.Join(goldenFilesDir, "ics")

	vcsFiles, err := os.ReadDir(vcsDir)
	if err != nil {
		t.Fatalf("failed to read VCS directory: %v", err)
	}

	var vcsFileNames []string
	for _, file := range vcsFiles {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".vcs") {
			vcsFileNames = append(vcsFileNames, file.Name())
		}
	}

	if len(vcsFileNames) == 0 {
		t.Fatalf("no .vcs files found in %s", vcsDir)
	}

	for _, vcsFileName := range vcsFileNames {
		t.Run(vcsFileName, func(t *testing.T) {
			vcsPath := filepath.Join(vcsDir, vcsFileName)

			icsFileName := strings.TrimSuffix(vcsFileName, ".vcs") + ".ics"
			icsPath := filepath.Join(icsDir, icsFileName)

			vcsFile, err := os.Open(vcsPath)
			if err != nil {
				t.Fatalf("failed to open VCS file %s: %v", vcsPath, err)
			}
			defer vcsFile.Close()

			expectedICS, err := os.ReadFile(icsPath)
			if err != nil {
				t.Fatalf("failed to read expected ICS file %s: %v", icsPath, err)
			}

			var output bytes.Buffer
			err = vcstoics.Convert(vcsFile, &output, email)
			if err != nil {
				t.Fatalf("convert function failed for %s: %v", vcsFileName, err)
			}

			actualICS := output.Bytes()
			if !bytes.Equal(actualICS, expectedICS) {
				t.Errorf("output mismatch for %s\nexpected:\n%s\n\nactual:\n%s",
					vcsFileName, string(expectedICS), string(actualICS))
			}
		})
	}
}

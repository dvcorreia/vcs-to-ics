// Copyright (c) Diogo Correia
// SPDX-License-Identifier: MIT

package main

import (
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	vcstoics "github.com/dvcorreia/vcs-to-ics"
)

func warning(format string, v ...any) {
	format = fmt.Sprintf("warning: %s\n", format)
	fmt.Fprintf(os.Stderr, format, v...)
}

func run() error {
	var (
		email  string
		merge  bool
		output string
	)

	flag.StringVar(&email, "email", "", "recipient email address for the calendar event")
	flag.BoolVar(&merge, "merge", false, "create a single ICS file for all events")
	flag.StringVar(&output, "o", "", "output directory for the .ics files")

	flag.Parse()

	if email == "" {
		return fmt.Errorf("missing email address")
	}

	var in []io.Reader

	if files := flag.Args(); len(files) > 0 {
		for _, name := range files {
			if !strings.HasSuffix(name, ".vcs") {
				warning("%s may not be a vcs file: is missing .vcs file extension", name)
			}

			f, err := os.OpenFile(name, os.O_RDONLY, 0444)
			if err != nil {
				return err
			}
			defer f.Close()

			in = append(in, f)
		}
	} else {
		stat, err := os.Stdin.Stat()
		if err != nil {
			return fmt.Errorf("failed to get stdin status: %w", err)
		}

		if (stat.Mode() & os.ModeCharDevice) == 0 {
			in = append(in, os.Stdin)
		} else {
			return fmt.Errorf("no .vcs files were specified")
		}
	}

	for _, r := range in {
		err := vcstoics.Convert(r, os.Stdout, email)
		if err != nil {
			return err
		}
	}

	return nil
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}
}

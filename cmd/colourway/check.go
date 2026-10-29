package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/janearc/libtheme-css/colourway"
	"github.com/janearc/libtheme-css/readable"
)

// errOutside is what check returns when a source has a colour outside the
// range. The colours were printed; this is the reason for the exit status.
var errOutside = errors.New("a colour is outside the reader's range")

// check holds each source, as its editor scheme, to the reader's measured
// range, and prints every colour that is outside it. The range is the
// reference range, or the file after --policy, which has the same shape.
// It returns errOutside if any source had a colour outside the range.
func check(arguments []string, to streams) error {
	policy, paths, err := rangeAndPaths(arguments)
	if err != nil {
		return err
	}
	outside := false
	for _, path := range paths {
		report, err := reportOn(path, policy)
		if err != nil {
			return err
		}
		for _, line := range report.Lines() {
			fmt.Fprintln(to.out, line)
		}
		outside = outside || !report.OK()
	}
	if outside {
		return errOutside
	}
	return nil
}

// rangeAndPaths splits check's arguments into the range to hold to and the
// sources to hold.
func rangeAndPaths(arguments []string) (policy readable.Policy,
	paths []string, err error) {
	flags := flag.NewFlagSet("check", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	file := flags.String("policy", "", "a measured range, as a file")
	if err = flags.Parse(arguments); err != nil {
		return policy, nil, err
	}
	if paths = flags.Args(); len(paths) == nothing {
		return policy, nil, errors.New("check what? give it sources")
	}
	policy, err = policyFrom(*file)
	return policy, paths, err
}

// policyFrom is the range in a file, or the reference range if the name
// is empty.
func policyFrom(name string) (policy readable.Policy, err error) {
	if name == "" {
		return readable.Reference()
	}
	file, err := os.Open(name)
	if err != nil {
		return policy, err
	}
	defer file.Close()
	return readable.Read(file)
}

// reportOn is a source's scheme held to the policy.
func reportOn(path string, policy readable.Policy) (readable.Report, error) {
	var report readable.Report
	source, err := colourway.Read(path)
	if err != nil {
		return report, err
	}
	scheme, err := source.Scheme()
	if err != nil {
		return report, fmt.Errorf("%s: %w", path, err)
	}
	return readable.Check(scheme, policy), nil
}

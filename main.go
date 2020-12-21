package main

import (
	"flag"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"os"
	ttemplate "text/template"

	"github.com/Masterminds/sprig/v3"
)

func main() {
	f := flag.CommandLine
	f.Usage = func() {
		fmt.Fprintf(f.Output(),
			"Usage:\n  sprig [-format <format>] [-file <templateFile>] <input>\n\nOptions:\n")
		f.PrintDefaults()
	}
	format := f.String("format", "text", "The template format (text|html) to use.")
	tmplFile := f.String("file", "", "A go template file to use as an alternative to stdin.")
	f.Parse(os.Args[1:])

	var tmpl io.Reader = os.Stdin
	if isFlagUsed(f, "file") {
		file, err := os.Open(*tmplFile)
		if err != nil {
			parseFail(f, "%s", err)
		}
		defer file.Close()
		tmpl = file
	}

	if len(f.Args()) > 1 {
		parseFail(f, "expects only one <input> argument")
	}

	if err := run(*format, flag.Arg(0), tmpl, os.Stdout); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(2)
	}
}

func parseFail(f *flag.FlagSet, format string, args ...interface{}) {
	fmt.Fprintf(f.Output(), format+"\n", args...)
	f.Usage()
	os.Exit(2)
}

func isFlagUsed(f *flag.FlagSet, name string) bool {
	found := false
	f.Visit(func(f *flag.Flag) {
		if f.Name == name {
			found = true
		}
	})
	return found
}

func run(format, input string, tmpl io.Reader, output io.Writer) error {
	templateError := func(err error) error {
		return fmt.Errorf("could not apply the go template (%s): %w", format, err)
	}
	b, err := ioutil.ReadAll(tmpl)
	if err != nil {
		return templateError(err)
	}
	switch format {
	case "text":
		fMap := sprig.TxtFuncMap()
		t, err := ttemplate.New("tmpl").Funcs(fMap).Parse(string(b))
		if err != nil {
			return templateError(err)
		}
		err = t.Execute(output, input)
		if err != nil {
			return templateError(err)
		}
	case "html":
		fMap := sprig.HtmlFuncMap()
		t, err := template.New("tmpl").Funcs(fMap).Parse(string(b))
		if err != nil {
			return templateError(err)
		}
		err = t.Execute(output, input)
		if err != nil {
			return templateError(err)
		}
	default:
		return fmt.Errorf("invalid %s format", format)
	}
	return nil
}

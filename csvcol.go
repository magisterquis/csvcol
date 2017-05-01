/*
 * csvcol.go
 * Program to select columns (and rows) from a CSV file.
 * by J. Stuart McMurray
 * Created 20141119
 * Last modified 20170501
 *
 * Copyright (c) 2014-2017 J. Stuart McMurray <kd5pbo@gmail.com>
 *
 * Permission to use, copy, modify, and distribute this software for any
 * purpose with or without fee is hereby granted, provided that the above
 * copyright notice and this permission notice appear in all copies.
 *
 * THE SOFTWARE IS PROVIDED "AS IS" AND THE AUTHOR DISCLAIMS ALL WARRANTIES
 * WITH REGARD TO THIS SOFTWARE INCLUDING ALL IMPLIED WARRANTIES OF
 * MERCHANTABILITY AND FITNESS. IN NO EVENT SHALL THE AUTHOR BE LIABLE FOR
 * ANY SPECIAL, DIRECT, INDIRECT, OR CONSEQUENTIAL DAMAGES OR ANY DAMAGES
 * WHATSOEVER RESULTING FROM LOSS OF USE, DATA OR PROFITS, WHETHER IN AN
 * ACTION OF CONTRACT, NEGLIGENCE OR OTHER TORTIOUS ACTION, ARISING OUT OF
 * OR IN CONNECTION WITH THE USE OR PERFORMANCE OF THIS SOFTWARE.
 */

package main

import (
	"bufio"
	"encoding/csv"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/magisterquis/ranges"
)

/* Global config */
var gc struct {
	csvfile     *string
	rows        *string
	rowfile     *string
	cols        *string
	colfile     *string
	verbose     *bool
	v           *bool
	debug       *bool
	d           *bool
	commentChar *string
}

func main() {
	/* Set flags and parse */
	gc.csvfile = flag.String("csvfile", "", "CSV file to read.  CSV-formatted data will be also be read from the file(s) listed on the command line (in the order listed).  If -csvfile is - or no files are listed on the command line and -csvfile is not specified, CSV-formatted data will be read from standard input (in which case, neither rowfile nor colfile may be -).  If both -csvfile and additional files are given, the file named by -csvfile will be read first (even if it is -).")
	gc.rows = flag.String("rows", "", "The row(-number)s to output.  This is given as a comma-separated list of row numbers or ranges.  Either the starting or ending number may be omitted in a range to indicate the first or last row, respectively.  Example: -3,5-7,9,11-, which outputs rows 1, 2, 3, 5, 6, 7, 9, and all rows from the 11th row to the end of the data (inclusive of the 11th row).  By default, all rows are output if neither -ros nor -rowfile are specified.  The row counter is not reset between each file.  It is as if all the files were concatenated.")
	gc.rowfile = flag.String("rowfile", "", "If specified, 1-indexed row numbers to to indicate rows to output will be read from this file.  The format is the nearly the same as for -rows, but may be given on multiple lines.  May be - to read from the standard input (in which case, neither csvfile nor colfile may be -).  If both this and -rows are specified, rows specified by either this file or -rows will be output.")
	gc.cols = flag.String("cols", "", "The column(-number)s to output.  This is given as a comma-separated list of column numbers or ranges.  Either the starting or ending number may be omitted in a range to indicate the first or last column, respectively.  Example: -3,5-7,9,11-, which outputs columns 1, 2, 3, 5, 6, 7, 9, and all columns from the 11th column to the end of the data (inclusive of the 11th column).  By default, all columns are output if neither -cols nor -colfile are specified.")
	gc.colfile = flag.String("colfile", "", "If specified, 1-indexed column numbers to to indicate columns to output will be read from this file.  The format is the nearly the same as for -columns, but may be given on multiple lines.  May be - to read from the standard input (in which case, neither csvfile nor rowfile may be -).  If both this and -cols are specified, columns specified by either this file or -cols will be output.")
	gc.commentChar = flag.String("commentchar", "#", "Comment character.  If a line starts with this character, it will be ignored.  Set to \"\" to disable ignoring comments.")
	gc.verbose = flag.Bool("verbose", false, "Print informational messages to the standard error stream.")
	gc.v = flag.Bool("v", false, "Same as -verbose")
	gc.debug = flag.Bool("debug", false, "Print debugging messages to the standard error stream.")
	gc.d = flag.Bool("d", false, "Same as debug")
	flag.Parse()

	/* Handle -v and -d */
	*gc.verbose = *gc.verbose || *gc.v
	*gc.debug = *gc.debug || *gc.d

	/* Ensure that only one of the files is stdin */
	s := false /* Using stdin */
	checkStdin(&s, "-" == *gc.rowfile)
	checkStdin(&s, "-" == *gc.colfile)
	checkStdin(&s, ("-" == *gc.csvfile) ||
		("" == *gc.csvfile && 0 == flag.NArg()))

	/* Work out which rows to print */
	rFilter := mkFilter(*gc.rows, *gc.rowfile, "row")
	/* Work out which columns to print */
	cFilter := mkFilter(*gc.cols, *gc.colfile, "column")

	debug("Row Filter: %v", rFilter)
	debug("Colunm Filter: %v", cFilter)

	/* Make an array of filenames to read. */
	csvfile := []string{}
	/* Only stdin */
	if "" == *gc.csvfile && 0 == flag.NArg() {
		csvfile = []string{"-"}
	} else {
		/* First -csvfile */
		if "" != *gc.csvfile {
			csvfile = append(csvfile, *gc.csvfile)
		}
		/* Then command-line files */
		if flag.NArg() > 0 {
			csvfile = append(csvfile, flag.Args()...)
		}
	}

	/* Set up stdout as a CSV writer */
	w := csv.NewWriter(os.Stdout)

	lineNumber := 1    /* Current line number */
	ldone := false     /* Above the filter */
	orsize := 1        /* Size of previous output record */
	comment := rune(0) /* CSV comment character */
	if len(*gc.commentChar) > 0 {
		comment = []rune(*gc.commentChar)[0]
	}

	/* Read data from each file */
	for _, f := range csvfile {
		/* Printable name */
		var fp *os.File
		fname := f
		if "-" == fname {
			fname = "standard input"
			fp = os.Stdin
		} else {
			fpl, err := os.Open(f)
			if err != nil {
				inform("Unable to open %v: %v", f, err)
				os.Exit(-5)
			}
			fp = fpl
		}
		verbose("Parsing %v", fname)
		/* Make a CSV reader */
		r := csv.NewReader(fp)
		/* Reader settings */
		r.Comment = comment
		r.FieldsPerRecord = -1
		r.LazyQuotes = true

		/* Parse lines until the file is done */
		for ; ; lineNumber++ {
			/* Get a line */
			record, e := r.Read()
			if nil != record {
				debug("%v) Got %v fields: %#v", lineNumber,
					len(record), record)
			}
			/* Give up if we have an error */
			if e != nil {
				/* If it's EOF, go to the next file */
				if "EOF" == e.Error() {
					break
				}
				debug("Got error reading %v (%T): %v", fname,
					e, e)
				break
			}
			/* Work out whether to ignore it */
			if !ldone {
				a, y := rFilter.AllowsOut(lineNumber)
				/* Read the next one if not allowed */
				if !a {
					continue
				}
				/* Done checking lines if all are allowed or
				if we're past the upper limit */
				if ranges.AllMatch == y || ranges.Above == y {
					ldone = true
				}
			}
			/* Roll an output slice */
			orec := make([]string, 0, orsize)
			cdone := false /* Done worrying about columns */
			/* Add the right columns */
			for i := 1; i <= len(record); i++ {
				/* Work out whether to add this column */
				if !cdone {
					a, y := cFilter.AllowsOut(i)
					if !a {
						continue
					}
					/* Done checking if upper limit or
					all allowed */
					if ranges.AllMatch == y ||
						ranges.Above == y {
						cdone = true
					}
				}
				orec = append(orec, record[i-1])
			}
			orsize = len(orec)

			/* Actually output line */
			if err := w.Write(orec); err != nil {
				inform("Error writing %v: %v", orec, err)
				os.Exit(-8)
			}
		}
		/* TODO: Finish this */
		/* Flush output after each file */
		w.Flush()
		if err := w.Error(); err != nil {
			inform("Error flushing output: %v", err)
			os.Exit(-6)
		}
	}

}

/* Check if is is true.  If it is and s is true, die with an error.  If is is
true and s isn't, make s true.  If it's false all is good. */
func checkStdin(s *bool, is bool) {
	/* Don't care if is isn't */
	if !is {
		return
	}
	/* Set s if it's not set */
	if !*s {
		*s = true
		return
	}
	/* If both are set, die with an error. */
	inform("Only one of -csvfile, -rowfile, or -colfile may come from " +
		"the standard input.\n")
	os.Exit(-1)
}

/* mkFilter makes a filter from the specified flagfile (i.e. rowfile) and flag
(i.e. rows).  Name is passed in for error reporting. */
func mkFilter(flag, flagfile, name string) ranges.Filter {
	debug("Making %v filter from flag [%v] and file [%v]", name, flag,
		flagfile)
	/* Filter to return */
	f := ranges.New(verbose, debug)
	/* If we have nothing to set, return a permissive filter */
	if "" == flag && "" == flagfile {
		f.All = true
		return f
	}

	/* Process ranges on the command line */
	if "" != flag {
		verbose("Processing %v ranges from the commandline (%v)", name, flag)
		if err := f.Update(flag); err != nil {
			inform("Unable to process %v ranges (%v): %v", name,
				flag, err)
			os.Exit(-3)
		}

	}

	/* Will be what we read from */
	var in *os.File
	/* Populate in with something */
	if "" != flagfile {
		if "-" == flagfile {
			in = os.Stdin
		} else {
			i, err := os.Open(flagfile)
			if err != nil {
				inform("Unable to open %v file %v: %v",
					name, flagfile, err)
				os.Exit(-2)
			}
			in = i
		}
		/* Read lines from the file */
		scanner := bufio.NewScanner(in)
		fname := flagfile
		if "-" == fname {
			fname = "standard input"
		}
		for scanner.Scan() {
			t := scanner.Text()
			verbose("Processing %v from %v", t, fname)
			if err := f.Update(t); err != nil {
				inform("Unable to process %v ranges from "+
					"%v: %v", name, fname, err)
				os.Exit(-7)
			}
		}
		if err := scanner.Err(); err != nil {
			inform("Error reading from %v: %v", fname, err)
			os.Exit(-4)
		}
	}

	return f
}

/* verbose prints a message if -v */
func verbose(f string, a ...interface{}) {
	if *gc.verbose || *gc.debug {
		inform(f, a...)
	}
}

/* debug prints a message if -v or -d */
func debug(f string, a ...interface{}) {
	f = "D: " + f /* Note as a debug string */
	if *gc.debug {
		inform(f, a...)
	}
}

/* inform prints informational mesages to stderr */
func inform(f string, a ...interface{}) {
	if !strings.HasSuffix(f, "\n") {
		f += "\n"
	}
	fmt.Fprintf(os.Stderr, f, a...)
}

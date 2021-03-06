csvcol
======

This is a little program that prints specified rows and columns of CSV files
to the standard output.

Quick Example:

```
./csvcol -cols=-3,5-7,9,11- -rows=1,10- datafile.csv
```

This reads datafile.csv and print to the standard output columns
1, 2, 3, 5, 6, 7, 9, and all columns from the 11th column to the end of the
data (inclusive of the 11th column).  Only the first row and all rows from the
10th onward (inclusive of the 10th row) are printed.

Input
-----
csvcol can read data from the standard input (the default) or from any number
of files.  Files to be read can be specified on the command line after any
flags (like -cols and -rows, above) or can be given with the -csvdata flag.

Row/Column Specification
------------------------
The rows and columns to be printed can be specifed in three ways: on the
command line with -rows/-cols, in a file with -rowfile/-colfile, or on the
standard input, by passing "-" as the argument to either -rowfile or -colfile
(but not both, of course).  -rows/-cols may be mixed with -rowfile/-colfile,
in which case rows or columns matched by either will be printed.

There are five ways to specify a row/column:
n   Specifies a single row or column
m-n Specifies a range of rows or columns
-n  Specifes all rows or columns from the first one (1, not 0) to the nth
n-  Specifies all rows or columns from the nth to the last
-   Specifies all rows or columns

These can be combined with commas, as in the above example.  By default, all
rows and columns are printed.

Examples
--------
Print the first and third column from all rows in data.csv which correspond to
lines that contain "ZZ" in the file file.csv:

grep -n ZZ file.csv | cut -f 1 -d ":" | csvcol -cols 1,3 -rowfile - data.csv

Three ways to print all but the first 10 columns in the file data.csv:

```
csvcol -cols 11- data.csv
cat data.csv | csvcol -cols 11-
csvcol -cols 11- <data.csv
```

Save the first five columns of bigfile.csv to smallfile.csv

```
csvcol -cols -5 bigfile.csv >smallfile.csv
```

Gotchas
-------
This is not well-tested code.  The -verbose and -debug flags (or -v and -d)
may help if things don't go as expected.  -h and -help (as well as any flag
that csvcol doesn't expect) will cause csvcol to print out a list of its flags
and what they do.

Building
--------
The only library not included in the go distribution is
github.com/magisterquis/ranges, which was written specifically for csvcol.  The
easiest way to build (and install) csvcol is with the following commands:

```
go get github.com/magisterquis/csvcol
go install github.com/magisterquis/csvcol
```

This, of course, assumes GOROOT is set up in the typical fashion.

Alternatively, the two .go files from csvcol and rages can be placed in a
directory.  The "package ranges" line at the top of ranges.go should be
changed to "package main" and "github.com/magisterquis/ranges" will need to be
removed from the import section in csvcol.go

Binaries
--------
There are two binaries in the git repo, csvcol.linux.x86 and csvcol.linux.x64.
They are 32-bit an 64-bit Linux binaries, respectively.  As they were compiled
by the go cross-compiler, it's not impossible that there will be hidden bugs,
but given the simplicity of the code, this is unlikely (at least for bugs due
to cross-compiling).

Sample Data
-----------
Included with the source is a sample CSV file containing census information
for Texas.  It should serve as a reasonable test file.

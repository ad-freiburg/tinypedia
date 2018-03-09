# tinypedia
A very bare bones, text only, local Wikipedia server. It works directly on the
`.bz2` compressed dumps without creating any additional files.

## Current State
Retrieving articles by title using the path `#<URL-encoded-article-name>`works
and the text can be viewed as extracted by `wtf_wikipedia.js` + some formatting
for sections. Sadly this fails to extract the text from special markup such as
IPA pronounciations. The raw mediawiki markdown can also be extracted using
`/wiki/<URL-encoded-article-name>`

Since we currently use URL encoding directly this is not compatible with the
title encoding used by Wikipedia (e.g. `Ada%20Lovelace` instead of `Ada_Lovelace`).

## Building and Installing
First make sure you have Go and the `go` command installed and that
`$GOTPATH/bin` is in your path. Then install with a simple `go get`

    go get github.com/ad-freiburg/tinypedia

## Running
Change to the directory containing the dump files

* enwiki-latest-pages-articles-multistream-index.txt.bz2
* enwiki-latest-pages-articles-multistream.xml.bz2

And simply run the tinypedia exectuable

    tinypedia

If you have named the files differenty use the `-i` and `-d` command line
switches to point `tinypedia` to the _index_ and _data_ files respectively.

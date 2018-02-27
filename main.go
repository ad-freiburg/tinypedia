package main

import (
	"bufio"
	"bytes"
	"compress/bzip2"
	"encoding/xml"
	"flag"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
)

var indexFilePath, contentFilePath string

func init() {
	const (
		defaultIndexFile   = "enwiki-latest-pages-articles-multistream-index.txt.bz2"
		defaultContentFile = "enwiki-latest-pages-articles-multistream.xml.bz2"
	)

	flag.StringVar(&indexFilePath, "i", defaultIndexFile, "the index file to use")
	flag.StringVar(&contentFilePath, "d", defaultContentFile, "the content file to use")
}

type OffsetAndId struct {
	Offset int64
	Id     uint64
}

func readBzip2StreamOffsetAndId(indexFile *os.File) (map[string]OffsetAndId, error) {
	indexFile.Seek(0, 0)
	offsetMap := make(map[string]OffsetAndId)
	indexStream := bzip2.NewReader(indexFile)
	indexScanner := bufio.NewScanner(indexStream)
	for indexScanner.Scan() {
		splits := strings.SplitN(indexScanner.Text(), ":", 3)
		offStr, idStr, currTitle := splits[0], splits[1], splits[2]
		offset, err := strconv.ParseInt(offStr, 10, 64)
		if err != nil {
			log.Println(err)
			continue
		}
		id, err := strconv.ParseUint(idStr, 10, 64)
		if err != nil {
			log.Println(err)
			continue
		}
		offsetMap[currTitle] = OffsetAndId{offset, id}
	}
	if err := indexScanner.Err(); err != nil {
		return offsetMap, err
	}

	return offsetMap, nil
}

func extractArticleMediawiki(bz2MultiStreamPath string, offId OffsetAndId) (content string, err error) {
	const (
		OUTSIDE       = iota
		IN_PAGE       = iota
		IN_ID         = iota
		IN_TEXT       = iota
		FOUND_ID      = iota
		IN_MATCH_TEXT = iota
	)
	bz2MultiStream, err := os.Open(bz2MultiStreamPath)
	if err != nil {
		return "", err
	}
	defer bz2MultiStream.Close()
	bz2MultiStream.Seek(offId.Offset, 0)
	contentStream := bzip2.NewReader(bz2MultiStream)
	dexml := xml.NewDecoder(contentStream)

	depth, pageDepth := 0, 0
	var tempData bytes.Buffer
	state := OUTSIDE
	for {
		tok, err := dexml.Token()
		if err != nil && err != io.EOF {
			log.Fatal(err)
		}
		switch tok := tok.(type) {
		case xml.StartElement:
			depth += 1
			switch {
			case tok.Name.Local == "page":
				pageDepth = depth
				state = IN_PAGE
			case tok.Name.Local == "id" && state != FOUND_ID:
				state = IN_ID
			case tok.Name.Local == "text":
				if state == FOUND_ID {
					state = IN_MATCH_TEXT
				} else {
					state = IN_TEXT
				}
			}
		case xml.EndElement:
			depth -= 1
			switch {
			case tok.Name.Local == "page":
				state = OUTSIDE
			case tok.Name.Local == "id" && state != FOUND_ID:
				state = IN_PAGE
				// Does this id belong to the latest page element
				if depth != pageDepth {
					tempData.Reset()
					continue
				}
				currId, err := strconv.ParseUint(tempData.String(), 10, 64)
				if err != nil {
					log.Println(err)
				} else if currId == offId.Id {
					state = FOUND_ID
				}
				tempData.Reset()
			case tok.Name.Local == "text":
				if state == IN_MATCH_TEXT {
					return tempData.String(), nil
				}
				state = IN_PAGE
			}
		case xml.CharData:
			if state == IN_ID || state == IN_MATCH_TEXT {
				tempData.Write(tok)
			}
		}
	}
	return content, err
}

type TinyWikiHandler struct {
	offsetMap       map[string]OffsetAndId
	contentFilePath string
}

func NewTinyWikiHandler(offsetMap map[string]OffsetAndId, contentFilePath string) *TinyWikiHandler {
	return &TinyWikiHandler{offsetMap, contentFilePath}
}

func (h *TinyWikiHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	title := r.URL.Path
	log.Println("Title:", title)
	offsetAndId, ok := h.offsetMap[title]
	if !ok {
		log.Println("Couldn't find id for", title)
		return
	}
	log.Println("Found offset:", offsetAndId.Offset, "and id:", offsetAndId.Id)
	content, err := extractArticleMediawiki(h.contentFilePath, offsetAndId)
	if err != nil {
		log.Println(err)
		return
	}
	io.WriteString(w, content)
}

func main() {
	flag.Parse()
	indexFile, err := os.Open(indexFilePath)
	if err != nil {
		log.Fatal(err)
	}
	offsetMap, err := readBzip2StreamOffsetAndId(indexFile)
	indexFile.Close()
	if err != nil {
		log.Fatal(err)
	}

	wikiHandler := NewTinyWikiHandler(offsetMap, contentFilePath)
	http.Handle("/wiki/", http.StripPrefix("/wiki/", wikiHandler))
	http.Handle("/", http.FileServer(http.Dir("static")))
	log.Fatal(http.ListenAndServe(":8080", nil))

}

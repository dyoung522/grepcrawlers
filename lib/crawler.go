package lib

import (
	"bufio"
	"fmt"
	"log"
	"regexp"
	"slices"
	"strconv"
	"strings"

	"github.com/taylorskalyo/goreader/epub"
)

type Crawlers map[string]Crawler

type Crawler struct {
	ID   string
	Name string
}

func (c Crawler) Key() string {
	return strings.ReplaceAll(strings.TrimSpace(c.ID), ",", "")
}

func (c Crawler) String() string {
	return fmt.Sprintf("Crawler #%s %q", c.ID, c.Name)
}

func (c Crawler) CSV() string {
	return fmt.Sprintf("%q,%q\n", c.ID, c.Name)
}

func (c Crawler) MarshalCSV() []byte {
	return []byte(c.CSV())
}

func (M *Crawlers) Add(crawler Crawler) {
	if M == nil {
		log.Fatal("Crawlers map is nil")
	}

	(*M)[crawler.Key()] = crawler
}

func (M *Crawlers) SortIDsNumerically() ([]string, error) {
	var (
		lastErr error
		keys    []string
	)

	for _, c := range *M {
		keys = append(keys, c.Key())
	}

	slices.SortFunc(keys, func(a, b string) int {
		aInt, err := strconv.ParseInt(a, 10, 64)
		if err != nil {
			lastErr = err // Store the error if conversion fails
			return 0      // Return 0 to indicate an issue or handle as needed
		}

		bInt, err := strconv.ParseInt(b, 10, 64)
		if err != nil {
			lastErr = err // Store the error if conversion fails
			return 0      // Return 0 to indicate an issue or handle as needed
		}

		if aInt > bInt {
			return 1
		}

		if aInt < bInt {
			return -1
		}

		return 0 // Numeric comparison
	})

	return keys, lastErr // Return the sorted slice and any error encountered
}

func ScanBook(ebook string, debug bool) Crawlers {
	crawlers := make(Crawlers)
	cregex := regexp.MustCompile(`(?i)crawler\s+#?([\d,]+)\.?\s+“([\w\s]+)\.?”`)
	rc, err := epub.OpenReader(ebook)
	if err != nil {
		panic(err)
	}
	defer rc.Close()

	// The rootfile (content.opf) lists all of the contents of an epub file.
	// There may be multiple rootfiles, although typically there is only one.
	for _, book := range rc.Rootfiles {
		if debug {
			log.Printf("Found rootfile: %s\n", book.Title)
		}
		// Print book title.
		log.Printf("Reading %s", book.Title)

		for _, item := range book.Itemrefs {
			efile, err := item.Open()
			if err != nil {
				log.Fatal("Error opening item: %w", item.HREF, err)
				continue
			}
			defer func() {
				if err := efile.Close(); err != nil {
					log.Fatal("Error closing item: %w", item.HREF, err)
				}
			}()
			scanner := bufio.NewScanner(efile)

			for scanner.Scan() {
				line := scanner.Text()
				if cregex.MatchString(line) {
					match := cregex.FindStringSubmatch(line)
					if len(match) < 3 {
						log.Printf("Invalid crawler format in line: %s", line)
						continue
					}
					ID, Name := strings.TrimSpace(match[1]), strings.TrimSpace(match[2])

					if debug {
						log.Println("Found crawler reference:", line)
						log.Printf("Crawler ID: %s, Name: %s\n", ID, Name)
					}

					crawlers.Add(Crawler{ID, Name})
				}
			}

			if err := scanner.Err(); err != nil {
				log.Fatal(err)
			}
		}
	}

	return crawlers
}

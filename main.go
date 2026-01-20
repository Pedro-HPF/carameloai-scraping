package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/chromedp"
)

type Article struct {
	Title      string    `json:"title"`
	Summary    string    `json:"summary"`
	URL        string    `json:"url"`
	DateString string    `json:"date_string"`
	ParsedDate time.Time `json:"-"` // This won't show in JSON, used for sorting
}

func main() {
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.UserAgent("Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36"),
		chromedp.Flag("headless", true),
	)

	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()
	ctx, cancel := chromedp.NewContext(allocCtx)
	defer cancel()
	ctx, cancel = context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	var nodes []*cdp.Node
	results := []Article{}

	fmt.Println("Scraping and Sorting Articles...")

	err := chromedp.Run(ctx,
		chromedp.Navigate("https://carameloai.com/en/news"),
		chromedp.WaitVisible(`article.node-type-news`, chromedp.ByQuery),
		chromedp.Sleep(2*time.Second),
		chromedp.Nodes(`article.node-type-news`, &nodes, chromedp.ByQueryAll),
	)

	if err != nil {
		log.Fatal(err)
	}

	for _, node := range nodes {
		var title, link, summary, dateStr string

		_ = chromedp.Run(ctx,
			chromedp.Text(`.node-title a span`, &title, chromedp.ByQuery, chromedp.FromNode(node)),
			chromedp.AttributeValue(`.node-title a`, "href", &link, nil, chromedp.ByQuery, chromedp.FromNode(node)),
			chromedp.Text(`.field--name-body`, &summary, chromedp.ByQuery, chromedp.FromNode(node)),
			chromedp.Text(`.node-date`, &dateStr, chromedp.ByQuery, chromedp.FromNode(node)),
		)

		if title != "" {
			if len(link) > 0 && link[0] == '/' {
				link = "https://carameloai.com" + link
			}

			dateStr = strings.TrimSpace(dateStr)

			parsedTime, _ := time.Parse("January 2, 2006", dateStr)

			results = append(results, Article{
				Title:      title,
				Summary:    summary,
				URL:        link,
				DateString: dateStr,
				ParsedDate: parsedTime,
			})
		}
	}

	// Sorts the slice in-place: Newest (latest date) first
	sort.Slice(results, func(i, j int) bool {
		return results[i].ParsedDate.After(results[j].ParsedDate)
	})

	saveToJSON(results)
}

func saveToJSON(data []Article) {
	file, _ := os.Create("results.json")
	defer file.Close()
	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	encoder.Encode(data)
	fmt.Printf("Success! %d sorted articles saved to results.json\n", len(data))
}

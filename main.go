package main

import (
	"fmt"
	"log"

	"github.com/gocolly/colly/v2"
)

func main() {
	c := colly.NewCollector(
		colly.AllowedDomains("carameloai.com"),
	)

	c.OnHTML("h1", func(e *colly.HTMLElement) {
		fmt.Println("Main Header found:", e.Text)
	})

	c.OnHTML(".post-title, h2", func(e *colly.HTMLElement) {
		fmt.Printf("Content Heading: %s \n", e.Text)
	})

	c.OnRequest(func(r *colly.Request) {
		fmt.Println("Visiting:", r.URL.String())
	})

	err := c.Visit("https://carameloai.com/")
	if err != nil {
		log.Fatal(err)
	}
}

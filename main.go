package main

import (
	"flag"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

type Node struct {
	word        string
	children    []*Node
	isSearched  bool
	isDuplicate bool
}

func (n *Node) appendChild(child *Node) {
	n.children = append(n.children, child)
}

func main() {
	// 入力ワード取得
	flag.Parse()
	word := flag.Arg(0)

	var queue []*Node
	var searchedWords []string

	// ルートノード作成＋キューに追加
	rootNode := Node{word: word, isSearched: false, isDuplicate: false}
	queue = append(queue, &rootNode)

	searchedCount := 0

	for len(queue) > 0 {
		node := queue[0]
		queue = queue[1:]

		if inArray(searchedWords, node.word) {
			node.isDuplicate = true
		} else if searchedCount < 20 {
			if searchedCount > 0 {
				time.Sleep(500 * time.Millisecond)
			}

			// wiki検索して子ノードのキーワードとして登録
			words := searchWiki(node)
			for _, childWord := range words {
				childNode := Node{word: childWord}
				node.appendChild(&childNode)
				queue = append(queue, &childNode)
			}
			node.isSearched = true

			// 検索済みワード登録
			searchedWords = append(searchedWords, word)
			searchedCount++
		}
	}

	printNode(&rootNode, 0)
}

func searchWiki(node *Node) (words []string) {
	// 検索
	res, err := http.Get("https://ja.wikipedia.org/wiki/" + url.QueryEscape(node.word))
	if err != nil {
		log.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != 200 {
		log.Fatalf("status code error: %d %s", res.StatusCode, res.Status)
	}
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		log.Fatal(err)
	}

	doc.Find(".mw-parser-output > p").First().Find("a").Each(func(i int, s *goquery.Selection) {
		// hrefが/wiki/XXXのものを取得※英語除く
		href, _ := s.Attr("href")
		href, _ = url.QueryUnescape(href)
		r := regexp.MustCompile("^/wiki/([^a-zA-z]*)$")
		if r.MatchString(href) {
			words = append(words, r.FindStringSubmatch(href)[1])
		}
	})
	return words
}

func printNode(node *Node, layer int) {
	postfix := ""
	if node.isDuplicate {
		postfix = "@"
	} else if !node.isSearched {
		postfix = "$"
	}

	prefix := strings.Repeat(" ", layer*2) + " - "
	fmt.Println(prefix + node.word + postfix)
	for _, child := range node.children {
		printNode(child, layer+1)
	}
}

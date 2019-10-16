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

	// wikiツリー作成＋表示
	wikiTree := createWikiTree(word)
	printWikiTree(wikiTree, 0)
}

func createWikiTree(word string) (node *Node) {
	var queue []*Node
	var searchedWords []string

	// ルートノード作成＋キューに追加
	rootNode := &Node{word: word, isSearched: false, isDuplicate: false}
	queue = append(queue, rootNode)

	searchedCount := 0

	for len(queue) > 0 {
		node := queue[0]
		queue = queue[1:]

		if inArray(searchedWords, node.word) {
			node.isDuplicate = true
		} else if searchedCount < 20 {
			if searchedCount > 0 {
				time.Sleep(1000 * time.Millisecond)
			}

			// wiki検索して子ノードのキーワードとして登録
			words := searchWiki(node)
			for _, childWord := range words {
				childNode := &Node{word: childWord}
				node.appendChild(childNode)
				queue = append(queue, childNode)
			}
			node.isSearched = true

			// 検索済みワード登録
			searchedWords = append(searchedWords, word)
			searchedCount++
		}
	}

	return rootNode
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

func printWikiTree(node *Node, layer int) {
	suffix := ""
	if node.isDuplicate {
		suffix = "@"
	} else if !node.isSearched {
		suffix = "$"
	}
	indent := strings.Repeat(" ", layer*2)
	fmt.Printf("%s - %s%s\n", indent, node.word, suffix)

	for _, child := range node.children {
		printWikiTree(child, layer+1)
	}
}

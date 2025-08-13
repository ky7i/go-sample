package main

import (
	"fmt"
	"net/http"
	"strings"
)

// reference: https://christina04.hatenablog.com/entry/routing-with-radix-tree
type Route struct {
	Path    string `json:"path"`
	handler http.HandlerFunc
}

type Node struct {
	Part     string  `json:"part"`
	Children []*Node `json:"children"`
	IsWild   bool    `json:"isWild"`
	Route    Route   `json:"route"`
}

// set pattern and route on Radix-tree
func (n *Node) insert(pattern string, route Route) {
	parts := strings.Split(pattern, "/")[1:]

	for _, part := range parts {
		child := n.matchChild(part)
		if child == nil {
			child = &Node{
				Part:   part,
				IsWild: part[0] == ':' || part[0] == '*' || part[0] == '{',
			}
			n.Children = append(n.Children, child)
		}
		// down a pattern layer
		n = child
	}
	// handler sets only lowest layer
	n.Route = route
}

func (n *Node) search(path string) Route {
	parts := strings.Split(path, "/")[1:]

	for _, part := range parts {
		child := n.matchChild(part)
		if child == nil {
			return Route{}
		}
		n = child
	}
	return n.Route
}

func (n *Node) matchChild(part string) *Node {
	// why doesn't "range" return i, child ?
	for i := range n.Children {
		if n.Children[i].Part == part || n.Children[i].IsWild {
			return n.Children[i]
		}
	}
	return nil
}

func Index(w http.ResponseWriter, r *http.Request) {
	fmt.Fprint(w, "Welcome!\n")
}

func Hello(w http.ResponseWriter, r *http.Request) {
	name := strings.TrimPrefix(r.URL.Path, "/hello/")
	fmt.Fprintf(w, "Hello, %s!\n", name)
}

func Hello2(w http.ResponseWriter, r *http.Request) {
	name := strings.TrimPrefix(r.URL.Path, "/hello/")
	fmt.Fprintf(w, "Hello2, %s!\n", name)
}

func main() {
	routes := []Route{
		{Path: "/hello", handler: Index},
		{Path: "/hello/:name", handler: Hello},
		{Path: "/hello/:name/foo", handler: Hello2},
		{Path: "/foo", handler: Index},
	}

	tree := &Node{}
	for _, route := range routes {
		tree.insert(route.Path, route)
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		path := r.URL.Path
		route := tree.search(path)

		route.handler(w, r)
	})

	http.ListenAndServe(":8080", nil)
}

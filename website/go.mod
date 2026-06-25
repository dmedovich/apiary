// The Docusaurus site is its own (codeless) module purely so it is excluded
// from the apiary module zip that `go get`/`go install` downloads. Without this,
// every consumer would also pull website/ (including the large npm lockfile).
// There is no Go code here; the directory is built with npm, not go.
module github.com/yaop-labs/apiary/website

go 1.22.0

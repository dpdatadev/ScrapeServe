Simple executable that I keep running to query in my other scripts and programs.

Feed it a url, get DOM elements or converted text back:

Examples:

```sh
#all links returned in a collection
curl -s 'http://127.0.0.1:7171/links?url=http://go-colly.org/'

#all plain text in the document
curl -s 'http://127.0.0.1:7171/text?url=http://go-colly.org/'

#all table elements on a page
curl -s 'http://127.0.0.1:7171/table?url=http://go-colly.org/'

#the whole document, links and all, converted to markdown
curl -s 'http://127.0.0.1:7171/md?url=http://go-colly.org/'

```

Typical Golang. Fast single file network programs that do useful stuff.

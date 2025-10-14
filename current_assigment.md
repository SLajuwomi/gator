RSS
The whole point of the gator program is to fetch the RSS feed of a website and store its content in a structured format in our database. That way we can display it nicely in our CLI.

RSS stands for "Really Simple Syndication" and is a way to get the latest content from a website in a structured format. It's fairly ubiquitous on the web: most content sites have an RSS feed.

Structure of an RSS Feed
RSS is a specific structure of XML (I know, gross). We will keep it simple and only worry about a few fields. Here's an example of the documents we'll parse:

<rss xmlns:atom="http://www.w3.org/2005/Atom" version="2.0">
<channel>
  <title>RSS Feed Example</title>
  <link>https://www.example.com</link>
  <description>This is an example RSS feed</description>
  <item>
    <title>First Article</title>
    <link>https://www.example.com/article1</link>
    <description>This is the content of the first article.</description>
    <pubDate>Mon, 06 Sep 2021 12:00:00 GMT</pubDate>
  </item>
  <item>
    <title>Second Article</title>
    <link>https://www.example.com/article2</link>
    <description>Here's the content of the second article.</description>
    <pubDate>Tue, 07 Sep 2021 14:30:00 GMT</pubDate>
  </item>
</channel>
</rss>

We'll then directly unmarshal this kind of document into structs like this:

type RSSFeed struct {
Channel struct {
Title string `xml:"title"`
Link string `xml:"link"`
Description string `xml:"description"`
Item []RSSItem `xml:"item"`
} `xml:"channel"`
}

type RSSItem struct {
Title string `xml:"title"`
Link string `xml:"link"`
Description string `xml:"description"`
PubDate string `xml:"pubDate"`
}

If there are any extra fields in the XML, the parser will just discard them, and if any are missing, the parser will leave them as their zero value.

Assignment
Write a func fetchFeed(ctx context.Context, feedURL string) (\*RSSFeed, error) function. It should fetch a feed from the given URL, and, assuming that nothing goes wrong, return a filled-out RSSFeed struct. Here are some useful docs (be sure to check the Overviews for examples if the entry lacks any):
http.NewRequestWithContext
http.Client.Do
I set the User-Agent header to gator in the request with request.Header.Set. This is a common practice to identify your program to the server.
io.ReadAll
xml.Unmarshal (works the same as json.Unmarshal)
Use the html.UnescapeString function to decode escaped HTML entities (like &ldquo;). You'll need to run the Title and Description fields (of both the entire channel as well as the items) through this function.
Add an agg command. Later this will be our long-running aggregator service. For now, we'll just use it to fetch a single feed and ensure our parsing works. It should fetch the feed found at https://www.wagslane.dev/index.xml and print the entire struct to the console.
Run and submit the CLI tests.

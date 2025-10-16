Aggregate
Feeds are essentially just lists of posts. A post represents a single web page. The entire point of the gator program is to fetch the actual posts from the feed URLs and store them in our database. That way we can display them nicely in our CLI.

Assignment
Enhance the agg command to actually fetch the RSS feeds, parse them, and print the posts to the console--all in a long-running loop.

Create a new migration that adds a last_fetched_at column to the feeds table. It should be nullable.
Add a MarkFeedFetched SQL query. It should simply set the last_fetched_at and updated_at columns to the current time for a given feed (probably by ID is simplest).
Add a GetNextFeedToFetch SQL query. It should return the next feed we should fetch posts from. We want to scrape all the feeds in a continuous loop. A simple approach is to keep track of when a feed was last fetched, and always fetch the oldest one first (or any that haven't ever been fetched). SQL has a NULLS FIRST clause that can help with this.
Write an aggregation function, I called mine scrapeFeeds. It should:
Get the next feed to fetch from the DB.
Mark it as fetched.
Fetch the feed using the URL (we already wrote this function)
Iterate over the items in the feed and print their titles to the console.
Update the agg command to now take a single argument: time_between_reqs.
time_between_reqs is a duration string, like 1s, 1m, 1h, etc. I used the time.ParseDuration function to parse it into a time.Duration value.

It should print a message like Collecting feeds every 1m0s when it starts.
Use a time.Ticker to run your scrapeFeeds function once every time_between_reqs. I used a for loop to ensure that it runs immediately (I don't like waiting) and then every time the ticker ticks:
ticker := time.NewTicker(timeBetweenRequests)
for ; ; <-ticker.C {
	scrapeFeeds(s)
}

Do NOT DOS the servers you're fetching feeds from. Anytime you write code that makes a request to a third party server you should be sure that you are not making too many requests too quickly. That's why I recommend printing to the console for each request, and being ready with a quick Ctrl+C to stop the program if you see something going wrong.

The agg command should now be a never-ending loop that fetches feeds and prints posts to the console. The intended use case is to leave the agg command running in the background while you interact with the program in another terminal.

You should be able to kill the program with Ctrl+C.

There are no CLI tests for this lesson, test your own program and make sure everything behaves as expected. Here are a few RSS feeds to get you started:

TechCrunch: https://techcrunch.com/feed/
Hacker News: https://news.ycombinator.com/rss
Boot.dev Blog: https://blog.boot.dev/index.xml
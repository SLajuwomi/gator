Feeds
Now that we have a way to fetch feeds from the internet, we need to store them in our database.

Assignment
Create a feeds table.
Like any table in our DB, we'll need the standard id, created_at, and updated_at fields. We'll also need a few more:

name: The name of the feed (like "The Changelog, or "The Boot.dev Blog")
url: The URL of the feed
user_id: The ID of the user who added this feed
Make the url field unique so that in the future we aren't downloading duplicate posts.

Use an ON DELETE CASCADE constraint on the user_id foreign key so that if a user is deleted, all of their feeds are automatically deleted as well. This will ensure we have no orphaned records and that deleting the users in the reset command also deletes all of their feeds.

Write the appropriate migrations and run them.

Add a new query to create a feed, then use sqlc generate to generate the Go code.
Add a new command called addfeed. It takes two args:
name: The name of the feed
url: The URL of the feed
At the top of the handler, get the current user from the database and connect the feed to that user.

If everything goes well, print out the fields of the new feed record.

Run and submit the CLI tests.
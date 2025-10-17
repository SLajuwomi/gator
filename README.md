# Gator Blog Aggregator

## Description

This project is a blog aggregator that gets posts from RSS feeds and aggregates them for easy viewing. Users can follow and add new feeds at will.

## Requirements

You will need Postgres and Golang installed on your PC to use this program.

1. How to install Postgres:

   - macOS with brew

   ```sh
   brew install postgresql@15
   ```

   - Linux / WSL (Debian). For Windows users, these [docs](https://learn.microsoft.com/en-us/windows/wsl/tutorials/wsl-database#install-postgresql) contain more detail. But simply:

   ```sh
   sudo apt update
   sudo apt install postgresql postgresql-contrib
   ```

2. How to install Golang:

   - macOS with brew

   ```sh
   brew install go
   ```

   - Linux (Debian)

   ```sh
   sudo apt-get update
   sudo apt-get install golang-go
   ```

   - Windows
     - Go [here](https://go.dev/dl/), download the .msi and follow the installation instructions.

You will also need to install the gator CLI using [go install](https://pkg.go.dev/cmd/go#hdr-Compile_and_install_packages_and_dependencies)

You can do this in one of two ways.

1. Without cloning the repo
   - Simply run
     ```sh
     go install github.com/slajuwomi/gator@latest
     ```
   - You will then be able to run `gator` anywhere on your machine.
2. If you choose to clone the repo
   - Run the following from the root of the repository (`gator/`)
   ```sh
   go install .
   ```
   - You will now be able to run `gator` inside the repository.

## Using the CLI

Gator works by using a config file to store the currently logged in user and register users. You will need to create the config file in the home directory on your machine before continuing.

In your home directory create a file named `.gatorconfig.json` and put the following contents in it:

```json
{
  "db_url": "postgres://example"
}
```

Once the config file is created, you can run the CLI. Here are some of the available commands and their usage.

Replace text surrounded by `<>` with your custom options.

| Command                                 | Description                                                                                                                                                                                                                                     |
| --------------------------------------- | ----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| `gator register <name>`                 | Register a new user with the passed name                                                                                                                                                                                                        |
| `gator login <name>`                    | Login with the designated username                                                                                                                                                                                                              |
| `gator reset`                           | Clear the databade and reset it                                                                                                                                                                                                                 |
| `gator users`                           | Print all users that are currently registered                                                                                                                                                                                                   |
| `gator agg <time_between_reqs>`         | Scrape posts from followed RSS feeds and add them to the database. This command runs infinitely, please do not DOS websites. Do `Ctrl-C` to stop the loop after some time. Use 1h1m1s format for time. ex: to set time as 1m, do `gator agg 1m` |
| `gator addfeed <url_name> <actual_url>` | Add feed to database. ex: `gator addfeed TechCrunch https://techcrunch.com/feed/`                                                                                                                                                               |
| `gator feeds`                           | Prints all feeds that have been added                                                                                                                                                                                                           |
| `gator follow <url>`                    | Follows a designated feed ex: `gator follow https://techcrunch.com/feed/`                                                                                                                                                                       |
| `gator following`                       | Print all feeds you are currently following to the console.                                                                                                                                                                                     |
| `gator unfollow <feed_url>`             | Unfollows a specified feed. es. `gator unfollow https://techcrunch.com/feed/`                                                                                                                                                                   |
| `gator browse <optional_limt>`          | Prints posts from followed feeds. Can optionally specify how many feeds to browse. If no limit is given, 2 posts will be returned ex: `gator browse 4`                                                                                          |

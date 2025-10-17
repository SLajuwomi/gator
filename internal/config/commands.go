package config

import (
	"context"
	"database/sql"
	"encoding/xml"
	"errors"
	"fmt"
	"html"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/slajuwomi/gator/internal/database"
)

type State struct {
	Db  *database.Queries
	Cfg *Config
	Ctx context.Context
}

type Command struct {
	CommandName string
	Arguments   []string
}

type Commands struct {
	AllCommands map[string]func(*State, Command) error
}

type RSSFeed struct {
	Channel struct {
		Title       string    `xml:"title"`
		Link        string    `xml:"link"`
		Description string    `xml:"description"`
		Item        []RSSItem `xml:"item"`
	} `xml:"channel"`
}

type RSSItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
}

func (c *Commands) Run(s *State, cmd Command) error {
	err := c.AllCommands[cmd.CommandName](s, cmd)
	if err != nil {
		return err
	}
	return nil
}

func (c *Commands) RegisterNewCommand(name string, f func(*State, Command) error) {
	c.AllCommands[name] = f
}

func HandlerLogin(s *State, cmd Command) error {
	if len(cmd.Arguments) == 0 {
		return errors.New("username is required")
	}
	_, err := s.Db.GetUser(context.Background(), cmd.Arguments[0])
	if err != nil {
		return fmt.Errorf("user not found in database (maybe try to register first): %v", err)
	}
	err = s.Cfg.SetUser(cmd.Arguments[0])
	if err != nil {
		return err
	}
	fmt.Printf("%v has been logged in!", cmd.Arguments[0])
	return nil
}

func HandlerRegister(s *State, cmd Command) error {
	if len(cmd.Arguments) == 0 {
		return errors.New("username is required")
	}
	dbUser, err := s.Db.CreateUser(context.Background(), database.CreateUserParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name:      cmd.Arguments[0]},
	)
	if err != nil {
		return fmt.Errorf("could not register: %v", err)
	}
	err = s.Cfg.SetUser(cmd.Arguments[0])
	if err != nil {
		return err
	}
	fmt.Printf("User has been set! You are now logged in as %v\n%+v", cmd.Arguments[0], dbUser)
	return nil
}

func HandleReset(s *State, cmd Command) error {
	err := s.Db.Clear(context.Background())
	if err != nil {
		return fmt.Errorf("reset failed: %v", err)
	} else {
		fmt.Println("Reset successful!")
	}
	return nil
}

func HandleGetAllUsers(s *State, cmd Command) error {
	dbUsers, err := s.Db.GetUsers(context.Background())
	if err != nil {
		return fmt.Errorf("error getting all users: %v", err)
	}
	for _, user := range dbUsers {
		if user.Name == s.Cfg.CurrentUserName {
			fmt.Printf("* %v (current)\n", user.Name)
		} else {
			fmt.Printf("* %v\n", user.Name)
		}
	}
	return nil
}

func FetchFeed(ctx context.Context, feedURL string) (*RSSFeed, error) {
	var newRSSFeed RSSFeed
	client := &http.Client{
		Timeout: 5 * time.Second,
	}
	req, err := http.NewRequestWithContext(ctx, "GET", feedURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create new request:\n%v", err)
	}
	req.Header.Set("User-Agent", "gator")
	res, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("FetchFeed client sending http request failed:\n%v", err)
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read request body:\n%v", err)
	}
	err = xml.Unmarshal(body, &newRSSFeed)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal XML into struct:\n%v", err)
	}
	newRSSFeed.Channel.Title = html.UnescapeString(newRSSFeed.Channel.Title)
	newRSSFeed.Channel.Description = html.UnescapeString(newRSSFeed.Channel.Description)
	for i, item := range newRSSFeed.Channel.Item {
		item.Title = html.UnescapeString(item.Title)
		item.Description = html.UnescapeString(item.Description)
		newRSSFeed.Channel.Item[i] = item
	}
	return &newRSSFeed, nil
}

func HandleAgg(s *State, cmd Command) error {
	if len(cmd.Arguments) < 1 {
		return fmt.Errorf("not enough arguments. expecting gator agg <time_between_reqs>")
	}
	timeBetweenReqs, err := time.ParseDuration(cmd.Arguments[0])
	if err != nil {
		return fmt.Errorf("failure getting time between requests: %v", err)
	}
	fmt.Printf("Collecting feeds every %v\n", cmd.Arguments[0])
	ticker := time.NewTicker(timeBetweenReqs)
	for ; ; <-ticker.C {
		scrapeFeeds(s)
	}
}
func parsePublishedAt(s string) (time.Time, error) {
	var t time.Time
	var err error
	layouts := []string{time.RFC1123Z, time.RFC1123, time.RFC850}
	for _, layout := range layouts {
		t, err = time.Parse(layout, s)
		if err == nil {
			return t, nil
		} else {
			fmt.Println("invalid layout")
		}
	}
	if t.IsZero() {
		return t, fmt.Errorf("no layout matched: %v", err)
	}
	return t, nil
}
func scrapeFeeds(s *State) error {
	nextFeedToFetch, err := s.Db.GetNextFeedToFetch(context.Background())
	if err != nil {
		return fmt.Errorf("failed to get next feed to fetch: %v", err)
	}
	s.Db.MarkFeedFetched(context.Background(), database.MarkFeedFetchedParams{
		LastFetchedAt: sql.NullTime{
			Time:  time.Now(),
			Valid: true,
		},
		UpdatedAt: time.Now(),
		ID:        nextFeedToFetch.ID,
	})
	feedStruct, err := FetchFeed(context.Background(), nextFeedToFetch.Url)
	if err != nil {
		return fmt.Errorf("error occurred running FetchFeed:\n%v", err)
	}
	for _, item := range feedStruct.Channel.Item {
		t, err := parsePublishedAt(item.PubDate)
		if err != nil {
			return fmt.Errorf("parsePublishedAt failed: %v", err)
		}
		post, err := s.Db.CreatePosts(context.Background(), database.CreatePostsParams{
			ID:          uuid.New(),
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
			Title:       item.Title,
			Url:         item.Link,
			Description: item.Description,
			PublishedAt: t,
			FeedID:      nextFeedToFetch.ID,
		})
		if err != nil {
			if strings.Contains(err.Error(), "duplicate key value violates unique constraint") {
				continue
			}
			return fmt.Errorf("error adding posts to database: %v", err)
		}
		fmt.Printf("Added %v post from %v to database. It can now be browsed\n", post.Title, nextFeedToFetch.Name)
	}
	return nil
}
func HandleAddFeed(s *State, cmd Command, user database.User) error {
	if len(cmd.Arguments) < 2 {
		return fmt.Errorf("not enough arguments. expecting addfeed url_name actual_url")
	}

	newFeed, err := s.Db.CreateFeed(context.Background(), database.CreateFeedParams{
		ID:            uuid.New(),
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
		LastFetchedAt: sql.NullTime{},
		Name:          cmd.Arguments[0],
		UserID:        user.ID,
		Url:           cmd.Arguments[1]})
	if err != nil {
		return fmt.Errorf("error creating feed: %v", err)
	}
	fmt.Printf("Created feed: %+v\n", newFeed)
	insertFeedFollow, err := s.Db.CreateFeedFollow(context.Background(), database.CreateFeedFollowParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		UserID:    user.ID,
		FeedID:    newFeed.ID,
	})
	if err != nil {
		return fmt.Errorf("error creating feed follow while adding feed: %v", err)
	}
	fmt.Printf("%v now following feed: %v\n", user.Name, insertFeedFollow.FeedName)
	return nil
}

func HandleGetAllFeeds(s *State, cmd Command) error {
	dbFeeds, err := s.Db.GetFeeds(context.Background())
	if err != nil {
		return fmt.Errorf("error getting all feeds: %v", err)
	}
	for _, feed := range dbFeeds {
		fmt.Printf("* ID: %v\n", feed.ID)
		fmt.Printf("* Created At: %v\n", feed.CreatedAt)
		fmt.Printf("* Updated At: %v\n", feed.UpdatedAt)
		fmt.Printf("* Name: %v\n", feed.Name)
		creatorUserName, err := s.Db.GetUserByID(context.Background(), feed.UserID)
		if err != nil {
			return fmt.Errorf("error getting name of user that created feed: %v", err)
		}
		fmt.Printf("* Creator of Feed: %v\n", creatorUserName.Name)
		fmt.Printf("* URL of Feed: %v\n", feed.Url)
		fmt.Println()
	}
	return nil
}

func HandleFeedFollow(s *State, cmd Command, user database.User) error {
	if len(cmd.Arguments) < 1 {
		return fmt.Errorf("not enough arguments. expecting gator follow <url>")
	}
	feedToFollow, err := s.Db.GetFeedByURL(context.Background(), cmd.Arguments[0])
	if err != nil {
		return fmt.Errorf("error getting feed to follow: %v", err)
	}
	insertFeedFollow, err := s.Db.CreateFeedFollow(context.Background(), database.CreateFeedFollowParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		UserID:    user.ID,
		FeedID:    feedToFollow.ID,
	})
	if err != nil {
		return fmt.Errorf("error creating feed follow: %v", err)
	}
	fmt.Printf("Name of Feed: %v\n", insertFeedFollow.FeedName)
	fmt.Printf("Current user: %v\n", insertFeedFollow.UserName)
	fmt.Printf("Successfully followed %v\n", insertFeedFollow.FeedName)
	return nil
}

func HandleFollowing(s *State, cmd Command, user database.User) error {
	allFollowing, err := s.Db.GetFeedFollowsForUser(context.Background(), user.Name)
	if err != nil {
		return fmt.Errorf("error getting all feeds followed by current user: %v", err)
	}
	fmt.Println("Currently following:")
	for _, feed := range allFollowing {
		fmt.Printf("* %v\n", feed.FeedName)
	}
	return nil
}

func HandleUnfollow(s *State, cmd Command, user database.User) error {
	if len(cmd.Arguments) < 1 {
		return fmt.Errorf("too few arguments. expected gator unfollow <feed_url>")
	}
	feed, err := s.Db.GetFeedByURL(context.Background(), cmd.Arguments[0])
	if err != nil {
		return fmt.Errorf("couldn't get feed with url: %v", err)
	}
	err = s.Db.DeleteFeedFollow(context.Background(), database.DeleteFeedFollowParams{
		UserID: user.ID,
		FeedID: feed.ID,
	})
	if err != nil {
		return fmt.Errorf("error deleting feed follow: %v", err)
	} else {
		fmt.Printf("successfully unfollowed %v\n", feed.Url)
	}
	return nil
}

func HandleBrowse(s *State, cmd Command, user database.User) error {
	var limit int64
	limit = 2
	var err error
	if len(cmd.Arguments) != 0 {
		if cmd.Arguments[0] != "" {
			limit, err = strconv.ParseInt(cmd.Arguments[0], 10, 64)
			if err != nil {
				return fmt.Errorf("failed to convert to int: %v", err)
			}
		}
	}
	posts, err := s.Db.GetPostsForUser(context.Background(), database.GetPostsForUserParams{
		UserID: user.ID,
		Limit:  int32(limit),
	})
	if err != nil {
		return fmt.Errorf("error getting posts for user: %v", err)
	}
	fmt.Printf("Found %d posts for user %s:\n", len(posts), user.Name)
	for _, post := range posts {
		fmt.Printf("%s from %s\n", post.PublishedAt.Format("Mon Jan 2"), post.FeedName)
		fmt.Printf("--- %s ---\n", post.Title)
		fmt.Printf("    %v\n", post.Description)
		fmt.Printf("Link: %s\n", post.Url)
	}
	return nil
}

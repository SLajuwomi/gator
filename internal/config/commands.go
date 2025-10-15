package config

import (
	"context"
	"encoding/xml"
	"errors"
	"fmt"
	"html"
	"io"
	"net/http"
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
	feedStruct, err := FetchFeed(context.Background(), "https://www.wagslane.dev/index.xml")
	if err != nil {
		return fmt.Errorf("error occurred running FetchFeed:\n%v", err)
	}
	fmt.Printf("%+v", feedStruct)
	return nil
}

func HandleAddFeed(s *State, cmd Command, user database.User) error {
	if len(cmd.Arguments) < 2 {
		return fmt.Errorf("not enough arguments. expecting addfeed url_name actual_url")
	}
	
	newFeed, err := s.Db.CreateFeed(context.Background(), database.CreateFeedParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name:      cmd.Arguments[0],
		UserID:    user.ID,
		Url:       cmd.Arguments[1]})
	if err != nil {
		return fmt.Errorf("error creating feed: %v", err)
	}
	fmt.Printf("Created feed: %+v\n", newFeed)
	insertFeedFollow, err := s.Db.CreateFeedFollow(context.Background(), database.CreateFeedFollowParams{
		ID: uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		UserID: user.ID,
		FeedID: newFeed.ID,
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
		return fmt.Errorf("not enough arguments. expecting go run . follow <url>")
	}
	feedToFollow, err := s.Db.GetFeedByURL(context.Background(), cmd.Arguments[0])
	if err != nil {
		return fmt.Errorf("error getting feed to follow: %v", err)
	}
	insertFeedFollow, err := s.Db.CreateFeedFollow(context.Background(), database.CreateFeedFollowParams{
		ID: uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		UserID: user.ID,
		FeedID: feedToFollow.ID,
	})
	if err != nil {
		return fmt.Errorf("error creating feed follow: %v", err)
	}
	fmt.Printf("Name of Feed: %v\n", insertFeedFollow.FeedName)
	fmt.Printf("Current user: %v\n", insertFeedFollow.UserName)
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

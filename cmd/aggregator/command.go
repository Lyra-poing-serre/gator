package aggregator

import (
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/Lyra-poing-serre/gator/internal/database"

	"github.com/google/uuid"
)

type command struct {
	name string
	args []string
}

type commands struct {
	registry map[string]func(*state, command) error
}

func (c *commands) run(s *state, cmd command) error {
	currentCommand, exist := c.registry[cmd.name]
	if !exist {
		return errors.New("unkown command")
	}
	return currentCommand(s, cmd)
}

func (c *commands) register(name string, f func(*state, command) error) {
	c.registry[name] = f
}

func handlerLogin(s *state, cmd command) error {
	if len(cmd.args) == 0 {
		return errors.New("the login handler expects a single argument, the username")
	}
	_, err := s.db.GetUser(context.Background(), cmd.args[0])
	if err != nil {
		return fmt.Errorf("user %s doesn't exist in DB", cmd.args[0])
	}

	err = s.config.SetUser(cmd.args[0])
	if err != nil {
		return err
	}
	fmt.Println("A new user has been set.")
	return nil
}

func handlerRegister(s *state, cmd command) error {
	if len(cmd.args) == 0 {
		return errors.New("the login handler expects a single argument, the username")
	}
	_, err := s.db.GetUser(context.Background(), cmd.args[0])
	if err == nil {
		return fmt.Errorf("user %s already in DB", cmd.args[0])
	}
	userParams := database.CreateUserParams{

		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name:      cmd.args[0],
	}
	usr, err := s.db.CreateUser(context.Background(), userParams)
	if err != nil {
		return err
	}
	s.config.SetUser(usr.Name)
	fmt.Printf("User %s created.\n", usr.Name)
	return nil
}

func handlerReset(s *state, cmd command) error {
	err := s.db.ResetUsers(context.Background())
	if err != nil {
		return err
	}
	fmt.Println("Users table is clean.")
	return nil
}

func handlerUsers(s *state, cmd command) error {
	users, err := s.db.GetUsers(context.Background())
	if err != nil {
		return err
	}
	for _, usr := range users {
		fmt.Printf("* %s", usr.Name)
		if usr.Name == s.config.CurrentUserName {
			fmt.Print(" (current)")
		}
		fmt.Println()
	}
	return nil
}

func handlerAgg(s *state, cmd command) error {
	if len(cmd.args) == 0 {
		return errors.New("missing arguments")
	}
	timeBetweenReqs, err := time.ParseDuration(cmd.args[0])
	if err != nil {
		return err
	}
	ticker := time.NewTicker(timeBetweenReqs)
	defer ticker.Stop()
	for ; ; <-ticker.C {
		scrapeFeed(s)
	}
}

func handlerAddFeed(s *state, cmd command, usr database.User) error {
	if len(cmd.args) < 2 {
		return errors.New("missing arguments")
	}

	feedParams := database.CreateFeedParams{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Name:      cmd.args[0],
		Url:       cmd.args[1],
		UserID:    usr.ID,
	}
	f, err := s.db.CreateFeed(context.Background(), feedParams)
	if err != nil {
		return err
	}
	fmt.Printf("User %s subscribed to %s.\n", usr.Name, f.Name)
	err = handlerFollow(s, command{args: []string{f.Url}}, usr)
	if err != nil {
		return err
	}
	return nil
}

func handlerFeeds(s *state, cmd command) error {
	feeds, err := s.db.ListFeeds(context.Background())
	if err != nil {
		return err
	}
	for _, feed := range feeds {
		fmt.Printf("* %s (%s) -> %s\n", feed.Name, feed.UserName, feed.Url)
	}
	return nil
}

func handlerFollow(s *state, cmd command, usr database.User) error {
	if len(cmd.args) == 0 {
		return errors.New("missing arguments")
	}

	feed, err := s.db.GetFeedFromUrl(context.Background(), cmd.args[0])
	if err != nil {
		return err
	}

	newFeedFollow, err := s.db.CreateFeedFollow(
		context.Background(), database.CreateFeedFollowParams{
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
			UserID:    usr.ID,
			FeedID:    feed.ID,
		})
	if err != nil {
		return err
	}
	fmt.Printf("%s (current) follow %s\n", newFeedFollow.UserName, newFeedFollow.FeedName)
	return nil
}

func handlerFollowing(s *state, cmd command, usr database.User) error {
	usrFeeds, err := s.db.GetFeedFollowsForUser(context.Background(), usr.Name)
	if err != nil {
		return err
	}
	for _, feed := range usrFeeds {
		fmt.Printf("* %s\n", feed.FeedName)
	}
	return nil
}

func handlerUnfollow(s *state, cmd command, usr database.User) error {
	if len(cmd.args) == 0 {
		return errors.New("missing arguments")
	}
	feed, err := s.db.GetFeedFromUrl(context.Background(), cmd.args[0])
	if err != nil {
		return err
	}
	err = s.db.DeleteFeedFollowsByIds(
		context.Background(), database.DeleteFeedFollowsByIdsParams{UserID: usr.ID, FeedID: feed.ID})
	if err != nil {
		return err
	}
	return nil
}

func handlerBrowse(s *state, cmd command, usr database.User) error {
	var limit int = 2
	if len(cmd.args) != 0 {
		limit, _ = strconv.Atoi(cmd.args[0])
	}
	posts, err := s.db.GetPostsForUser(context.Background(), database.GetPostsForUserParams{
		UserID: usr.ID,
		Limit:  int32(limit),
	})
	if err != nil {
		return err
	}

	for _, post := range posts {
		fmt.Printf("%s: %s\n", post.Name, post.Title.String)
		fmt.Printf("Published at: %s\n", post.PublishedAt)
		fmt.Println(post.Description.String)
	}
	return nil
}

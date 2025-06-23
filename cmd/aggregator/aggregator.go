package aggregator

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/Lyra-poing-serre/gator/internal/database"
	"github.com/Lyra-poing-serre/gator/internal/settings"
)

type state struct {
	db     *database.Queries
	config *settings.Config
}

func LaunchAggregator(c *settings.Config) {
	db, err := sql.Open("postgres", c.DbURL)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer db.Close()
	s := state{db: database.New(db), config: c}
	cmds := commands{registry: make(map[string]func(*state, command) error)}

	if len(os.Args) < 2 {
		fmt.Println("too few arguments")
		os.Exit(1)
	}
	registerAggregatorCommands(&cmds)

	cmd := command{
		name: os.Args[1],
		args: os.Args[2:],
	}
	err = cmds.run(&s, cmd)
	if err != nil {
		log.Fatalln(err)
	}
	os.Exit(0)
}

func registerAggregatorCommands(c *commands) {

	c.register("login", handlerLogin)
	c.register("register", handlerRegister)
	c.register("reset", handlerReset)
	c.register("users", handlerUsers)
	c.register("agg", handlerAgg)
	c.register("addfeed", middlewareLoggedIn(handlerAddFeed))
	c.register("feeds", handlerFeeds)
	c.register("follow", middlewareLoggedIn(handlerFollow))
	c.register("following", middlewareLoggedIn(handlerFollowing))
	c.register("unfollow", middlewareLoggedIn(handlerUnfollow))
	c.register("browse", middlewareLoggedIn(handlerBrowse))
}

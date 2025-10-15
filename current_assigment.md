Middleware
We have 3 command handlers (and we'll add more) that all start by ensuring that a user is logged in.

addfeed
follow
following
They all share this code (or something similar):

user, err := s.db.GetUser(context.Background(), s.cfg.CurrentUserName)
if err != nil {
	return err
}

Let's create some "middleware" that abstracts this away for us. In addition, if we need to modify this code for any reason later, there will be only one place that must be edited.

Middleware is a way to wrap a function with additional functionality. It is a common pattern that allows us to write DRY code.

Assignment
Create logged-in middleware. It will allow us to change the function signature of our handlers that require a logged in user to accept a user as an argument and DRY up our code. Here's the function signature of my middleware:

middlewareLoggedIn(handler func(s *state, cmd command, user database.User) error) func(*state, command) error

You'll notice it's a higher order function that takes a handler of the "logged in" type and returns a "normal" handler that we can register. I used it like this:

cmds.register("addfeed", middlewareLoggedIn(handlerAddFeed))

Test your code before and after this refactor to make sure that everything still works.

Run and submit the CLI tests.
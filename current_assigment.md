SQLC
SQLC is an amazing Go program that generates Go code from SQL queries. It's not exactly an ORM, but rather a tool that makes working with raw SQL easy and type-safe.

We will use Goose to manage our database migrations (the schema), and SQLC to generate Go code that our application can use to interact with the database (run queries).

Assignment
Install SQLC.
SQLC is just a command line tool, it's not a package that we need to import. I recommend installing it using go install. Installing Go CLI tools with go install is easy and ensures compatibility with your Go environment.

go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest

Then run sqlc version to make sure it's installed correctly.

Configure SQLC. You'll always run the sqlc command from the root of your project. Create a file called sqlc.yaml in the root of your project. Here is mine:
version: "2"
sql:
  - schema: "sql/schema"
    queries: "sql/queries"
    engine: "postgresql"
    gen:
      go:
        out: "internal/database"

We're telling SQLC to look in the sql/schema directory for our schema structure (which is the same set of files that Goose uses, but sqlc automatically ignores "down" migrations), and in the sql/queries directory for queries. We're also telling it to generate Go code in the internal/database directory.

Write a query to create a user. Inside the sql/queries directory, create a file called users.sql. Here's the format:
-- name: CreateUser :one
INSERT INTO users (id, created_at, updated_at, name)
VALUES (
    $1,
    $2,
    $3,
    $4
)
RETURNING *;

$1, $2, $3, and $4 are parameters that we'll be able to pass into the query in our Go code. The :one at the end of the query name tells SQLC that we expect to get back a single row (the created user).

Keep the SQLC postgres docs handy, you'll probably need to refer to them again later.

Generate the Go code. Run sqlc generate from the root of your project. It should create a new package of go code in internal/database. You'll notice that the generated code relies on Google's uuid package, so you'll need to add that to your module:
go get github.com/google/uuid

Write another query to get a user by name. I used the comment -- name: GetUser :one to tell SQLC that I expect to get back a single row. Again, generate the Go code to ensure that it works.
Import a PostgreSQL driver.
We need to add and import a Postgres driver so our program knows how to talk to the database. Install it in your module:

go get github.com/lib/pq

Add this import to the top of your main.go file:

import _ "github.com/lib/pq"

This is one of my least favorite things working with SQL in Go currently. You have to import the driver, but you don't use it directly anywhere in your code. The underscore tells Go that you're importing it for its side effects, not because you need to use it.

Open a connection to the database, and store it in the state struct:
In main(), load in your database URL to the config struct and sql.Open() a connection to your database:

db, err := sql.Open("postgres", dbURL)

Use your generated database package to create a new *database.Queries, and store it in your state struct:

dbQueries := database.New(db)

type state struct {
	db  *database.Queries
	cfg *config.Config
}

Create a register handler and register it with the commands. Usage:
go run . register lane

It should:

Ensure that a name was passed in the args.
Create a new user in the database. It should have access to the CreateUser query through the state -> db struct.
Pass context.Background() to the query to create an empty Context argument.
Use the uuid.New() function to generate a new UUID for the user.
created_at and updated_at should be the current time.
Use the provided name.
Exit with code 1 if a user with that name already exists.
Set the current user in the config to the given name.
Print a message that the user was created, and log the user's data to the console for your own debugging.
Test the register command by running it with a name:

go run . register lane

Use psql to verify that the user was created:

Mac: psql postgres
Linux: sudo -u postgres psql
\c gator

SELECT * FROM users;

Update the login command handler to error (and exit with code 1) if the given username doesn't exist in the database. You can't login to an account that doesn't exist!
Take a good look at the tests and run them before submitting.

Be sure to migrate down and back up with goose before each run/submit, because the tests assume a clean database.

Submit the CLI tests after a final down/up migration.
# Contents
1. [JSON](#json)
    - [Sending JSON Responses](#sending-json-responses)
    - [Parsing JSON Requests](#parsing-json-requests)
2. [Database Connection Pool](#database-connection-pool)
    - [Configuring the Connection Pool](#configuring-the-connection-pool)
3. [golang-migrate](#using-golang-migrate-module)
    - [Executing SQL Migrations](#executing-sql-migrations)
    - [Rolling-back DB Versions](#rolling-back-database-version)
    - [Fixing Errors in SQL Migrations](#fixing-errors-in-sql-migrations)
4. [PostgreSQL CRUD](#postgresql-json-crud-operations)
    - [Create/Insert](#createinsert)
    - [Read/Fetch](#readfetch)
    - [Put/Patch](#updatepatch)
    - [Delete](#delete)
5. [Concurrency Control](#concurrency-control)
    - [Optimistic Locking](#optimistic-locking)
    - [Pessimistic Locking](#pessmistic-locking)
    - [Round-Trip Locking](#round-trip-locking)
6. [Listing Records](#listing-records)
    - [Parsing Query String Params](#parsing-query-string-parameters)
    - [Validating Query String Params](#validating-query-string-parameters)
    - [Listing Data](#listing-data)
    - [Filtering Lists](#filtering-lists)
    - [Full-Text Search](#full-text-search)
    - [Sorting Lists](#sorting-lists)
    - [Pagination](#pagination)
 
# JSON

### Sending JSON Responses

- Two methods to construct JSON objects in Go via the `encoding/json` package:
    1. `json.Marshal(data)`
        - Can opt for the `json.MarshalIndent(data, "", "\t")` method instead, which automatically formats the JSON with whitespace, but tradeoff is increased memory usage and slower return times
    2. `json.NewEncoder(w).Encode(data)`

- <strong>If constructing a JSON object from a struct, field names *MUST* be exported in order for them to be recognized by the `encoding/json` package</strong>

- JSON Struct Tags:
    - Most frequently used to change the name of keys in the JSON object
    ```
    # Change field to snake_case
    type Movie struct {
        ID        int64     `json:"id"`
    }
    ```
    - JSON struct tag directives:
        - `omitempty` directive hides the field *if and only if* the struct field value is empty
        - `- (hypen)` directive *NEVER* shows the field in the JSON output
            - using the hypen is highly preferred over un-exporting the struct field, as it explicitly marks the field
        - `string` directive forces data to be represented as a string
            - only works on `int*`, `uint*`, `float*`, and `bool` types
        
- Advanced JSON customization
    - Under the hood, when Go encodes a particular *type* to JSON, it looks if that type has a method which satifies the `json.Marshaler` interface
    ```
    type Marshaler interface {
        MarshalJSON() ([]byte, error)
    }
    ```
    - If the type satfies the interface, Go will call the method to determine how to encode the data, if not, it will fallback to its own internal set of rules
    - An example of this can be found in Go's [`time.Time`](https://github.com/golang/go/blob/73d213708e3186b48d5147b8eb939fdfd51f1f8d/src/time/time.go#L1267) package
        - `time.Time`'s is actually a struct with a `MarshalJSON()` method that outputs a *RFC 3339* format representation of itself

### Parsing JSON Requests

- Two methods to parse JSON objects into Go objects via the `encoding/json` package:
    1. `json.Decoder` type
    2. `json.Unmarshal()` method

- The code uses the `json.Decoder` method, heres an example using the `json.Unmarshal()` method for reference:
```
func (app *application) exampleHandler(w http.ResponseWriter, r *http.Request) {
    var input struct {
        Foo string `json:"foo"`
    }

    // Use io.ReadAll() to read the entire request body into a []byte slice.
    body, err := io.ReadAll(r.Body)
    if err != nil {
        app.serverErrorResponse(w, r, err)
        return
    }
    
    // Use the json.Unmarshal() function to decode the JSON in the []byte slice to the
    // input struct. Again, notice that we are using a *pointer* to the input
    // struct as the decode destination.
    err = json.Unmarshal(body, &input)
    if err != nil {
        app.errorResponse(w, r, http.StatusBadRequest, err.Error())
        return
    }

    fmt.Fprintf(w, "%+v\n", input)
}
```

 - For decoding JSON from a HTTP request body, using `json.Decoder` is the better choice since it is more efficient and requires less code

- When decoding a JSON object into a struct, the key/value pairs in the JSON are mapped to the struct fields based on the struct tag names. If there is no matching struct tag, Go will attempt to decode the value into a field that matches the key name (exact matches are preferred, but it will fall back to a case-insensitive match). Any JSON key/value pairs which cannot be successfully mapped to the struct fields will be silently ignored

- `json.Decoder` is designed to support streams of JSON data. When we call `Decode()` on our request body, it actually reads the first JSON value only from the body and decodes it. If we made a second call to `Decode()`, it would read and decode the second value and so on
    - If `Decode()` is called a second time and it doesn't return an `io.EOF` error, that means theres multiple JSON/non-JSON values in the request body

- Customizing How JSON is Parsed
    - When Go is decoding a type, it first looks to see if that type contains a method which satifies the `json.Unmarshaler()` interface
    ```
    type Unmarshaler interface {
        UnmarshalJSON([]byte) error
    }
    ```
    - Go will first call that method to decode the data

#### JSON items with `null` values will be ignored and will remain unchanged

# Database Connection Pool

- There are two types of connections:
    1. In-use connections: Connections actively used to execute SQL queries or DB operations
    2. Idle connections: Connections avaliable for use

- If there are no connections avaliable, Go will spawn a new connection
- If a connection is bad, Go will re-try the connection *twice* before removing the connection and creating a new one

### Configuring the Connection Pool

- The connection pool has 4 methods to customize its behavior:

1. `SetMaxOpenConns()` method
    - Sets a limit on how many "open" (in-use + idle) connections are avaliable in the pool
    - PostgreSQL sets a default limit of 100 connections
        - Can be overridden by the `max_connections` setting in the `postgres.conf` file
        - To avoid an error, the limit in our application should be comfortably below PostgreSQL's default/custom limit
    - Setting a limit comes with a caveat, if all connections are used up, new DB operations are left to hang (potentially indefinite) while waiting for a new connection to be free'd up
        - Mitigate this by always setting a timeout on database tasks using a `context.Context` object

2. `SetMaxIdleConns()` method
    - Sets a limit on the number of "idle" connections in the pool
        - Default max idle connections is 2
    - `MaxIdleCons` should always be less than or equal to `MaxOpenConns`, Go automatically enforces this
    - Keeping too many idle connections open can consume too much memory

3. `SetConnMaxLifetime()` method
    - Sets the maximum length of time that a connection can be reused for
    - Not an idle timeout!

4. `SetConnMaxIdleTimeout()` method
    - Sets the maximum amount of time a connection can idle before its marked as expired
        - By default, this has no limit
    - Can be used in combination with `SetMaxIdleCons()` to set a high number of idle connections and perform a cleanup operation if any connection hasn't been used in a while

# Using `golang-migrate` module

```
# Linux installation
$ cd /tmp
$ curl -L https://github.com/golang-migrate/migrate/releases/download/v4.16.2/migrate.linux-amd64.tar.gz | tar xvz
$ mv migrate ~/go/bin/
$ migrate --version
```
- Use the `migrate create` command to generate a pair of *migration files*
```
$ migrate create -seq -ext=.sql -dir=./migrations create_movies_table
/home/tlei_dev/greenlight/migrations/000001_create_movies_table.up.sql
/home/tlei_dev/greenlight/migrations/000001_create_movies_table.down.sql
```
- `-seq` flag instructs migrate to use sequential numbering (00001, 00002, ..etc) for the migration files instead of a Unix timestamp
- `-ext` flag allows us to specify the file extension of our migration files
- `-dir` flag specifies the directory we want our migration files to be created
    - if directory doesn't already exists, it will be created for us
- `create_movies_table` is a descriptive label we give to the migration files signifying their contents

### Executing SQL migrations

```
$ migrate -path=./migrations -database=$GREENLIGHT_DB_DSN up
1/u create_movies_table (38.19761ms)
2/u add_movies_check_constraints (63.284269ms)
```

### Rolling-back database version

- To see which migration version your database is currently on using `golang-migrate` tool's `version` command:
```
$ migrate -path=./migrations -database=$EXAMPLE_DSN version
```

- To migrate up or down to a specific version use `goto` command:
```
$ migrate -path=./migrations -database=$EXAMPLE_DSN goto {version}
```

- To roll-back *all* migrations, use the `down` command:
```
$ migrate -path=./migrations -database=$EXAMPLE_DSN down
Are you sure you want to apply all down migrations? [y/N]
y
Applying all down migrations
2/d create_bar_table (39.988791ms)
1/d create_foo_table (59.460276ms)
```

### Fixing errors in SQL migrations

- When a migration which contains an error is ran, all SQL statements up to the erroneous one will be applied and then the migrate tool will exit with a message describing the error
    - So if a migration file contains *multiple* SQL statements, its possible the migration file was *partially* applied
    - This will leave the database in an *unknown* state
    - Further signified by the database version displaying a *dirty* field

- Fix:
    1. Investigate the original error and figure out if the migration file was partially applied
    2. Manually roll-back the partially applied migration
    3. Force the `version` number in the `schema_migrations` table to the correct value using the `force` command:
        ```
        $ migrate -path=./migrations -database=$EXAMPLE_DSN force 1
        ```
    4. Once `force` is applied, the database will be considered "clean" and migrations should be able to run again

# PostgreSQL JSON CRUD Operations

### Create/Insert

- PostgreSQL specific `RETURNING` clause returns values from any record being manipulated by an `INSERT`, `UPDATE`, OR `DELETE` statement
    - If we use this clause and are expecting the query to return exactly *one* row, we must execute the query using the `sql.DB`'s `QueryRow()` method

- Placeholder parameter inputs are denoted by `$N` 

- To store Go arrays/slices, we need to pass the array into the `pq.Array()` adapter
    - `pq.Array()` adapter converts the `[]string` slice to a `pq.StringArray` type
        - `pq.StringArray` type implements the [`driver.Valuer`](https://pkg.go.dev/database/sql/driver#Valuer) and [`sql.Scanner`](https://pkg.go.dev/database/sql#Scanner) interfaces which are necessary to translate the type so our PostgreSQL database can understand it
    - Same goes for `[]bool`, `[]int]`, `[]float64`...

- In our handler, set a `Location` header directing the requesting user where to find our newly created resource
    - Also in the event of our resource being created successfully the appropriate status code to use is `201 (StatusCreated)`


### Read/Fetch

- Similar to Create/Insert method, when reading the PSQL column type `[]text`, the `pq.Array()` method is required to translate the array to a slice which Go can recognize

- Q: Why not use an unsigned integer to type the `ID` field when we know it will never be a negative number?
    - PostgreSQL does *NOT* have an unsigned integer type, it is best to align Go and database integer types as closely as possible
    - Go's `database/sql` package doesn't actually support integer values greater than 9223372036854775807 (max value for an integer of type `int64`) 
        - Its possible for a `uint64` value to be greater than this, which would lead to Go generating a runtime error


### Update/Patch

For our app's `updateMovieHandler`, we'll specifically:
1. Extract the movie ID from the URL using the app.readIDParam() helper.
2. Fetch the corresponding movie record from the database using the Get() method that we made in the previous chapter.
3. Read the JSON request body containing the updated movie data into an input struct.
4. Copy the data across from the input struct to the movie record.
5. Check that the updated movie record is valid using the data.ValidateMovie() function.
6. Call the Update() method to store the updated movie record in our database.
7. Write the updated movie data in a JSON response using the app.writeJSON() helper.

#### Handling Partial Updates aka Patch

- When decoding the request body, any fields in our input which *don't* have corresponding JSON key-value pairs will retain their *zero* value
- This causes a problem as we cannot tell the difference between a key/value pair that needs to be updated with their zero value versus a key/value pair that was omitted
- We take advantage of the fact that the zero value for pointers is `nil`
- By making all the fields in our input struct pointer types, we are able to check for the case that the input field is omitted

### Delete

- Use the `db.Exec()` query method if our SQL statement does not return any rows
    - It also conveniently returns a `sql.Result` object that contains the `RowsAffected()` method
    - We can use this method to check if 0 rows have been affected and return a error not found in response to that

- Depending on if a human or a machine makes a request to our endpoint
    - Respond with a `200` status if its a human for UX 
    - Response with a `204 No Content` status if a machine is hitting the endpoint
 
# Concurrency Control

- Concurrency control are methods which prevent data races
- A data race is when two user requests to update the same resource at the exact same time, which in turn, forces the server to race the two requests

### Optimistic Locking

- Assumes nothing is going to change while reading a record, many collisions are not expected
- Implementation: Make a check which makes sure the database row hasn't changed before writing the record
    - The check can be in the form of an auto-incremented `version` column
    - A UUID should be used if its important that the identifier cannot be guessed

### Pessmistic Locking

- Assumes something will change as a record is being read and locks the record, a collision is expected and anticipated

### Round-trip Locking

- An extension to optimisic locking where the *client* can pass the version number *they* expect in an `X-Expected-Version` header 
- This can be useful to help ensure the client is not sending their update based on outdated information
```
// Rough implementation
if r.Header.Get("X-Expected-Version") != "" {
    if strconv.FormatInt(int64(movie.Version), 32) != r.Header.Get("X-Expected-Version") {
        app.editConflictResponse(w, r)
        return
    }
}
```

# Listing Records

### Parsing Query String Parameters

1. Created helper methods to parse query string paramters to their desired types
    - Note: Query string parameters are of type `url.Values` from the `net/url` package
    - Helper methods can potentially accept a `validator` as an argument and add errors to the validator map as needed
2. In the handler, we created an `Input` struct which holds all our expected query string *keys*
3. Grab the `url.Values` map with the `r.URL.Query()` method
4. We parse all the query strings into the `Input` struct

### Validating Query String Parameters

1. Created a struct which holds our filter query string values in a separate file
2. Added a `ValidateFilters(v *validator.Validator, f Filters)` method which validates our filter query strings
    - Utilized the `v.Check()` method to perform the checks as well as add errors to the errors map
3. Embedded the filters struct in our `listMovieHandlers` input struct
4. Added the query string values to the filters struct and validated the values

### Listing Data

1. Created a `GetAll()` method to serve as the model for our `listMovieHandlers` handler
2. In `GetAll()`, we use the `db.QueryContext()` method to query *multiple* rows
    - Be sure to defer a call to `rows.Close()` to close the result set before exiting the fn
3. Initalize an empty slice of *Movies
4. Iterate over`rows.Next()` and scan each movie into a struct and finally append the struct to the *Movies slice
    - When scanning slices/arrays, don't forget to use the `pq.Array()` adapter
5. After the `rows.Next()` loop has finished, call `rows.Err()` to collect any errors that occurred during iteration
```
    if rows.Err() != nil {}
```
6. If all goes well, return the slice of Movies to the handler

### Filtering Lists

As opposed to dynamically generating a SQL query at runtime by concatenating the filter clauses, we opt to use conditional clauses where we can control whats being filtered by forcing the clause to return true.

```
SELECT id, created_at, title, year, runtime, genres, version
FROM movies
WHERE (LOWER(title) = LOWER($1) OR $1 = '') 
AND (genres @> $2 OR $2 = '{}') 
ORDER BY id
```

`(LOWER(title) = LOWER($1) OR $1 = '')`
- Returns rows where the column title matches the parameter OR skips the filter entirely if the parameter equals an empty string

`(genres @> $2 OR $2 = '{}')`
- @> symbol denotes "contains"
- Return rows where the genres column contains one or more values from the parameter or skip the filter entirely if the paramter is an empty array

- [PostgreSQL Array Operators](https://www.postgresql.org/docs/9.6/functions-array.html)

### Full-Text Search

PostgreSQL provides powerful full-text search capabilities through the use of configurable *lexemes*

```
# Consider this WHERE clause
WHERE (to_tsvector('simple', title) @@ plainto_tsquery('simple', $1) OR $1 = '') 
```

- The `to_tsvector('simple', title)` function takes in a text column and splits it into lexemes
    - We specify the *simple* configuration which means that the lexemes are just lowercase versions of the words in the title
    - Title: "The Breakfast Club" | Lexemes: "the" "breakfast" "club"
    - There are vast configuration options, such as the removal of common words or applying language-specific stemming
- The `plainto_tsquery('simple', $1)` function takes in a search term and normalizes it into a formatted *query term*
    - It strips any special characters and inserts the `&` operator between the words
    - Search Term: "The Club" | Query term: "the" & "club" - Matches rows which contain both lexemes "the" and "club"
- The `@@` operator is the *matches* operator
    - We use it in the query to check whether our *query term matches the lexeme* 

Adding DB Indexes

- Database indexes allow our SQL queries to perform quickly as our dataset grows
- Indexes helps us avoid full table scans and avoid re-generating lexemes for our columns every time a query is ran
- For a full-text search, it makes sense to utilize PostgreSQL's `GIN` index type
    - `GIN` indexes efficiently handle full-text search involving arrays and other advanced queries

- Few other notable text-search methods include the `STRPOS()` function and `ILIKE` operator
    - STRPOS() is a sub-string search
    - ILIKE matches case-insensitive patterns

### Sorting Lists

When working with PostgreSQL, its important to remember that the order of returned rows is only guaranteed by the rules that your `ORDER BY` clause imposes
- This can be problematic when it comes to paginating rows with the same sort relevance (i.e. sorting two records with the same year)
- Gurantee the order by including a primary key/unique constraint on the `ORDER BY` clause

1. Implemented methods against our `Filters` struct: `sortColumn()` and `sortDirection()`
2. Added string interpolation to our SQL query and dynamically added the filter conditions with `fmt.Sprintf`
3. We also added a second constraint on our `ORDER BY` clause
    - Added the `id ASC` to guarantee our records always remain the same order for pagination

### Pagination

- Utilize the `LIMIT` and `OFFSET` clauses along with some simple math to paginate

```
LIMIT = page_size
OFFSET = (page - 1) * page_size
```

- We can utilize parameterized queries since numbers are not SQL keywords

Pagination Metadata

- Metadata such as current and last page numbers and total number of avaliable records can help give the client context about the response and make navigating through pages easier


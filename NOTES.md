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

- Executing SQL migrations
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


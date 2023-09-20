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
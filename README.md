# pilosa

Pilosa is a bitmap index database.

[![Build Status](https://travis-ci.com/pilosa/pilosa.svg?token=Peb4jvQ3kLbjUEhpU5aR&branch=master)](https://travis-ci.com/pilosa/pilosa)

## Getting Started

Pilosa requires Go 1.7 or greater.

You can download the source by running `go get`:

```sh
$ go get github.com/pilosa/pilosa
```

Now you can install the `pilosa` binary:

```sh
$ go install github.com/pilosa/pilosa/cmd/...
```

Now run a single pilosa node with the default configuration:

```sh
pilosa server
```

## Configuration

You can specify a configuration by setting the `-config` flag when running `pilosa`.


```sh
pilosa server --config custom-config-file.cfg
```

The config file uses the [TOML](https://github.com/toml-lang/toml) configuration file format,
and should look like:

```
data-dir = "/tmp/pil0"
host = "127.0.0.1:15000"

[cluster]
replicas = 2

[[cluster.node]]
host = "127.0.0.1:15000"

[[cluster.node]]
host = "127.0.0.1:15001"
```

You can generate a template config file with default values with:

```sh
pilosa config
```

The first two configuration options will be unique to each node in the cluster:

`data-dir`: directory in which data is stored to disk

`host`: IP and port of the pilosa node

The remaining configuration options should be the same on every node in the cluster.

`replicas`: the number of replicas within the cluster

`[[cluster.node]]`: specifies each node within the cluster

## Docker

You can create a Pilosa container using `make docker` or equivalently:
```
docker build -t pilosa:latest .
```

You can run a temporary container using:
```
docker run -it --rm --name pilosa -p 15000:15000 pilosa:latest
```

When you click `Ctrl+C` to stop the container, the container and the data in the container will be erased. You can leave out `--rm` flag to keep the data in the container. See [Docker documentation](https://docs.docker.com) for other options.

## Usage

You can interact with Pilosa via HTTP requests to the host:port on which you have Pilosa running.
The following examples illustrate how to do this using `curl` with a Pilosa cluster running on
127.0.0.1 port 15000.

Return the version of Pilosa:
```sh
$ curl "http://127.0.0.1:15000/version"
```

Return a list of all databases and frames in the index:
```sh
$ curl "http://127.0.0.1:15000/schema"
```


### Queries

Queries to Pilosa require sending a POST request where the query itself is sent as POST data.
You specify the database on which to perform the query with a URL argument `db=database-name`.

A query sent to database `exampleDB` will have the following format:

```sh
$ curl -X POST "http://127.0.0.1:15000/query?db=exampleDB" -d 'Query()'
```

The `Query()` object referenced above should be made up of one or more of the query types listed below.
So for example, a SetBit() query would look like this:
```sh
$ curl -X POST "http://127.0.0.1:15000/query?db=exampleDB" -d 'SetBit(id=10, frame="foo", profileID=1)'
```

Query results have the format `{"results":[]}`, where `results` is a list of results for each `Query()`. This
means that you can provide multiple `Query()` objects with each HTTP request and `results` will contain
the results of all of the queries.

```sh
$ curl -X POST "http://127.0.0.1:15000/query?db=exampleDB" -d 'Query() Query() Query()'
```

---
#### SetBit()
```
SetBit(id=10, frame="foo", profileID=1)
```
A return value of `{"results":[true]}` indicates that the bit was toggled from 0 to 1.
A return value of `{"results":[false]}` indicates that the bit was already set to 1 and therefore nothing changed.

SetBit accepts an optional `timestamp` field:
```
SetBit(id=10, frame=f, profileID=2, timestamp="2016-12-11T10:09:07")
```

---
#### ClearBit()
```
ClearBit(id=10, frame="foo", profileID=1)
```
A return value of `{"results":[true]}` indicates that the bit was toggled from 1 to 0.
A return value of `{"results":[false]}` indicates that the bit was already set to 0 and therefore nothing changed.

---
#### SetBitmapAttrs()
```
SetBitmapAttrs(id=10, frame="foo", category=123, color="blue", happy=true)
```
Returns `{"results":[null]}`

---
#### SetProfileAttrs()
---
```
SetProfileAttrs(id=10, category=123, color="blue", happy=true)
```

Returns `{"results":[null]}`

---
#### Bitmap()
```
Bitmap(id=10, frame="foo")
```
Returns `{"results":[{"attrs":{"category":123,"color":"blue","happy":true},"bits":[1,2]}]}` where `attrs` are the
attributes set using `SetBitmapAttrs()` and `bits` are the bits set using `SetBit()`.

In order to return profile attributes attached to the profiles of a bitmap, add `&profiles=true` to the query string. Sample response:
```
{"results":[{"attrs":{},"bits":[10]}],"profiles":[{"id":10,"attrs":{"category":123,"color":"blue","happy":true}}]}
```

---
#### Union()
```
Union(Bitmap(id=10, frame="foo"), Bitmap(id=20, frame="foo")))
```
Returns a result set similar to that of a `Bitmap()` query, only the `attrs` dictionary will be empty: `{"results":[{"attrs":{},"bits":[1,2]}]}`.
Note that a `Union()` query can be nested within other queries anywhere that you would otherwise provide a `Bitmap()`.

---
#### Intersect()
```
Intersect(Bitmap(id=10, frame="foo"), Bitmap(id=20, frame="foo")))
```
Returns a result set similar to that of a `Bitmap()` query, only the `attrs` dictionary will be empty: `{"results":[{"attrs":{},"bits":[1]}]}`.
Note that an `Intersect()` query can be nested within other queries anywhere that you would otherwise provide a `Bitmap()`.

---
#### Difference()
```
Difference(Bitmap(id=10, frame="foo"), Bitmap(id=20, frame="foo")))
```
`Difference()` represents all of the bits that are set in the first `Bitmap()` but are not set in the second `Bitmap()`.  It returns a result set similar to that of a `Bitmap()` query, only the `attrs` dictionary will be empty: `{"results":[{"attrs":{},"bits":[2]}]}`.
Note that a `Difference()` query can be nested within other queries anywhere that you would otherwise provide a `Bitmap()`.

---
#### Count()
```
Count(Bitmap(id=10, frame="foo"))
```
Returns the count of the number of bits set in `Bitmap()`: `{"results":[28]}`

---
#### Range()
```
Range(id=10, frame="foo", start="1970-01-01T00:00", end="2000-01-02T03:04")
```

---
#### TopN()
```
TopN(frame="bar")
```
Returns all Bitmaps in the cache from frame `bar` sorted by the count of bits.

```
TopN(frame="bar", n=20)
```
Returns the top 20 Bitmaps from frame `bar`.

```
TopN(Bitmap(id=10, frame="foo"), frame="bar", n=20)
```
Returns the top 20 Bitmaps from `bar` sorted by the count of bits in the intersection with `Bitmap(id=10)`.

```
TopN(Bitmap(id=10, frame="foo"), frame="bar", n=20, field="category", [81,82])
```
Returns the top 20 Bitmaps from `bar`in attribute `category` with values `81 or
82` sorted by the count of bits in the intersection with `Bitmap(id=10)`.

## Development

### Updating dependencies

To update dependencies, you'll need to install [Glide][].

Then add the new dependencies in your project:

```sh
$ glide get github.com/foo/bar
```

### Protobuf

If you update protobuf (pilosa/internal/internal.proto), then you need to run `go generate`
```sh
$ go generate
```

### Version

In order to set the version number, compile Pilosa with the following argument:
```sh
$ go install --ldflags="-X main.Version=1.0.0"
```

[Glide]: http://glide.sh/

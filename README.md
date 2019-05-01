# amznode

Amazing Co needed an piece of software to model their company tree structure.

amznode is a small HTTP API which can store nodetrees.

## Running the application

The application can be run locally using docker-compose `make compose`

The server now listens on **localhost:8080**

## Endpoints

The following are the endpoints exposed by amznode. All endpoints expect an
empty body and hence all input is passed directly in the URL.

### POST `/:rootName`

_Creates a new root node_

The created node is returned in the response body.

### POST `/:parentID/:childName`

_Adds a new child to a node_

The created node is returned in the response body.

A parentID of 0 will create a new root node as 0 is interpreted as not having a
parent.

### GET `/`

_Gets all registered root nodes_

The root nodes are returned as a list where each node also has its immediate
children.

### GET `/:id`

_Gets a node with the given id along with its immediate children_

Returns node as json in the response body.

An id of 0 will return the list of registered root nodes.

### PUT `/:id?parentID=:parentID`

_Changes the parent of a node._

Please note that Cycles are not allowed in the tree structure. That means that
you can't change the parent of a node to one of it's decendent children.

### DELETE `/:id`

_Deletes a node along with all its decendent children_

## Example

Start the server with with docker-compose using `make compose`

Given the tree

```
- root1
  - child3
  - child4
    - child7
      - child9
    - child8
  - child5
- root2
  - child6
```

Lets create the nodes:

```
$ curl -X POST "http://localhost:8080/root1" && \
  curl -X POST "http://localhost:8080/root2" && \
  curl -X POST "http://localhost:8080/1/child3" && \
  curl -X POST "http://localhost:8080/1/child4" && \
  curl -X POST "http://localhost:8080/1/child5" && \
  curl -X POST "http://localhost:8080/2/child6" && \
  curl -X POST "http://localhost:8080/4/child7" && \
  curl -X POST "http://localhost:8080/4/child8" && \
  curl -X POST "http://localhost:8080/7/child9"
{"id":1,"name":""root1","root_id":1,"height":0}
{"id":2,"name":"root2","root_id":2,"height":0}
{"id":3,"parent_id":1,"name":"child3","root_id":1,"height":1}
{"id":4,"parent_id":1,"name":"child4","root_id":1,"height":1}
{"id":5,"parent_id":1,"name":"child5","root_id":1,"height":1}
{"id":6,"parent_id":2,"name":"child6","root_id":2,"height":1}
{"id":7,"parent_id":4,"name":"child7","root_id":1,"height":2}
{"id":8,"parent_id":4,"name":"child8","root_id":1,"height":2}
{"id":9,"parent_id":7,"name":"child9","root_id":1,"height":3}
```

Lets get the roots:

```
$ curl -X GET -s "http://localhost:8080/" | jq
[
  {
    "id": 1,
    "name": "root1",
    "root_id": 1,
    "height": 0,
    "children": [
      {
        "id": 3,
        "parent_id": 1,
        "name": "child3",
        "root_id": 1,
        "height": 1
      },
      {
        "id": 4,
        "parent_id": 1,
        "name": "child4",
        "root_id": 1,
        "height": 1
      },
      {
        "id": 5,
        "parent_id": 1,
        "name": "child5",
        "root_id": 1,
        "height": 1
      }
    ]
  },
  {
    "id": 2,
    "name": "root2",
    "root_id": 2,
    "height": 0,
    "children": [
      {
        "id": 6,
        "parent_id": 2,
        "name": "child6",
        "root_id": 2,
        "height": 1
      }
    ]
  }
]
```

Lets get a specific node:

```
$ curl -X GET -s "http://localhost:8080/4" | jq
{
  "id": 4,
  "parent_id": 1,
  "name": "child4",
  "root_id": 1,
  "height": 1,
  "children": [
    {
      "id": 7,
      "parent_id": 4,
      "name": "child7",
      "root_id": 1,
      "height": 2
    },
    {
      "id": 8,
      "parent_id": 4,
      "name": "child8",
      "root_id": 1,
      "height": 2
    }
  ]
}
```

Lets change the parent of a node to the other root node and then get the root nodes:

```
$ curl -X PUT "http://localhost:8080/4?parentID=2"
$ curl -X GET -s "http://localhost:8080/" | jq
[
  {
    "id": 1,
    "name": "root1",
    "root_id": 1,
    "height": 0,
    "children": [
      {
        "id": 3,
        "parent_id": 1,
        "name": "child3",
        "root_id": 1,
        "height": 1
      },
      {
        "id": 5,
        "parent_id": 1,
        "name": "child5",
        "root_id": 1,
        "height": 1
      }
    ]
  },
  {
    "id": 2,
    "name": "root2",
    "root_id": 2,
    "height": 0,
    "children": [
      {
        "id": 4,
        "parent_id": 2,
        "name": "child4",
        "root_id": 2,
        "height": 1
      },
      {
        "id": 6,
        "parent_id": 2,
        "name": "child6",
        "root_id": 2,
        "height": 1
      }
    ]
  }
]
```

Lets delete a node or actually an entire tree, and then get the root nodes:

```
$ curl -X DELETE "http://localhost:8080/2"
$ curl -X GET -s "http://localhost:8080/" | jq
[
  {
    "id": 1,
    "name": "root1",
    "root_id": 1,
    "height": 0,
    "children": [
      {
        "id": 3,
        "parent_id": 1,
        "name": "child3",
        "root_id": 1,
        "height": 1
      },
      {
        "id": 5,
        "parent_id": 1,
        "name": "child5",
        "root_id": 1,
        "height": 1
      }
    ]
  }
]
```

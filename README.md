# amznode

Amazing Co needed an piece of software to model their company tree structure.

amznode is a small HTTP API which can store nodetrees. It has the following endpoints:

## Endpoints

The following are the endpoints exposed by amznode. All endpoints expect an
empty body and hence all input is passed directly in the URL.

### POST `/:parentID/:childName`

*Adds a new child to a node*

This endpoint returns the newly created child node as json in the response
body.

The server supports multiple root nodes. In order to create a root node post to
the special `parentID` of `0`.

### GET `/:id`

*Gets a node with the given id along with its immediate children*

This endpoint returns node as json in the response body.

### PUT `/:id?parentID=:parentID`

*Changes the parent of a node.*

This endpoint returns an empty response body.

Please note that Cycles are not allowed in the treestructure. That means that
you can't change the parent of a node to one of it's decendent children.

### DELETE `/:id` 

*Deletes a node along with all its decendent children*

## Example

Run the code with `docker-compose`

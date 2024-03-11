# go-BenXAdmin

BenXAdmin is a prototype Benefit Administration.

The goal is to create a model for a next-generation benefit administration system.
This is based on the authors past experience in building benefit administration systems combining all the best learnings from examining several benefit administration systems.

It has the following key advantages over existing systems:
* Next Generation Benefits Administration Domain Model for:
    * Person
    * Person Roles
        * Employee/Worker
        * Benefit Participant
        * Covered Person/Dependent
    * Benefit Plan
* Next Generation Business Process Model
    * Business Process Definition (Template)
    * Person Business Process
* Modern Run time Architecture 
    * Mongo for data storage
    * Event Architecture using Kafka
    * REST API
* Modern Programming Model - Golang
* Based on Solid Design Standards/Best Practices


## Environment setup

You need to have:
 [Go](https://golang.org/),
[Docker](https://www.docker.com/), and
[Docker Compose](https://docs.docker.com/compose/)


Verify the tools by running the following commands:

```sh
go version
docker --version
docker-compose --version
```


## Start in development mode

In the project directory run the command (you might
need to prepend it with `sudo` depending on your setup):
```sh
docker-compose -f docker-compose-dev.yml up
```

This starts a local MongoDB on `localhost:27017`.


Navigate to the `server` folder and start the back end:

```sh
cd server
go run server.go
```
The back end will serve on http://localhost:8080.

Navigate to the `client/webapp` folder, and start the front end development server by running:

```sh
cd client/webapp
go run server.go

```
The application will be available on http://localhost:3000.
 
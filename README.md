# To GPX

<p align="center">
    <img src="https://i.imgur.com/9uvTXeh.png" title="To GPX" />
</p>

This is the server code for the "To GPX" web application.

The client code can be found at [Garee/to-gpx-client](https://github.com/Garee/to-gpx-client).

## Quick Start

Create the docker container:

`$ docker build -t to-gpx-server`

Run the docker container:

`$ docker run -d -p 8000:8080 to-gpx-server`

The service will be available at http://localhost:8000

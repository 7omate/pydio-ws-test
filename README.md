# pydio-ws-test
Simple program to test the websockects are up and correctly configured

# Build

	git clone https://github.com/7omate/pydio-ws-test
	go build
	./simpleWSclient -h

# Sample usage

## Connect with credentials

The following will proceed to a basic authentication, get the server configuration for pydio-booster and try to connect to the websocket.

	./simpleWSclient -server http://SERVERURL/ -user USERNAME -password THEPASSWORD

## Connect with signed wsurl

The server parameter is required for proper origin checks.

	./simpleWSclient -server http://SERVERURL/ -wsURL ws://127.0.0.1:8090/ws?auth_hash=8e677357cef0f3119e290bb094038aca8f9d913c:3f3e92d56576f30a539d2dd7a12d85eca0af94102df4c4876c915d4800529b81&auth_token=BggSqV6BBW9RXot6W9W3te12

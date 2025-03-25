# Chirpy
Chirpy is a twitter clone made for the boot.dev [Learn HTTP Servers course](https://www.boot.dev/courses/learn-http-servers-golang). It has an API for authentication, managing 'chirps' and users, and a webhook for a mock payment system.

## Motivation
This repository might be of service if you'd like to look at a minimalistic and functional implementation of a website's backend. Key consideration has been granted to security principles (API keys and password hashing) as well as stateless (JSON web tokens) and stateful (refresh tokens) authentication management. The rudimentary middleware (site visit tracking) may also be of interest.

## Getting Started
Everything you need to get this repository up and running.
### Installation:
Clone this repository onto your machine. You'll need to create a `.env`	file with the following fields:
```
DB_URL="" //insert database url here
PLATFORM="dev"
TOKEN_SECRET="" //generate a random 256 bit string
POLKA_KEY="mock payment api key - unavailable"
```


### Requisites:
* [Go toolchain](https://go.dev/doc/install)
* [PostgreSQL](https://www.postgresql.org/download/)
	You'll need to initialize a database.
* [Goose](https://github.com/pressly/goose) 
	Once Postgres is up and running, you'll need to [apply all available migrations](https://github.com/pressly/goose?tab=readme-ov-file#up).
* [SQLC](https://github.com/sqlc-dev/sqlc)
	No generated files are included in this repo. These can be created with `sqlc generate`.

### Running
After ensuring the dependences are installed, run `go mod download` to import all of the required packages.

To launch the program, use `go run .` from the root folder of the project.

The default port is `8080`, you can check if the site is running by visiting [http://localhost:8080/app](http://localhost:8080/app)

## API
### Health Check

-   **GET**  `/api/healthz` - Checks the health of the API.
		Can be viewed in browser: http://localhost:8080/api/healthz
    

### Admin Endpoints

-   **GET**  `/admin/metrics` - Retrieves metrics for site visits.
		Can be viewed in browser: http://localhost:8080/admin/metrics
    
-   **POST**  `/admin/reset` - Removes all users from the database.
    

### User Management

-   **POST**  `/api/users` - Creates a new user.
		- Expects a JSON body with `email` and `password` fields.
		- Returns a JSON body with all user field except the hashed password. JWT and Refresh token are empty as they aren't generated until login.
    
-   **PUT**  `/api/users` - Updates an existing user.
		- Expects an `Authentication` header with a `Bearer [refresh token]` value.
		- Expects a JSON body with `email` and `password` fields.
    
-   **POST**  `/api/login` - Authenticates a user and returns a JWT.
		- Expects a JSON body with `email` and `password` fields.
		- Returns a JSON body with all user field except the hashed password.
    
-   **POST**  `/api/refresh` - Generate a new JWT for user.
		- Expects an `Authentication` header with a `Bearer [refresh token]` value.
		- Expects an **empty** body.
		- Returns a JSON body with the new JWT token in the `token` field
    
-   **POST**  `/api/revoke` - Revokes a refresh token.	
		- Expects an `Authentication` header with a `Bearer [refresh token]` value.
		- Expects an **empty** body.
    

### Chirps Endpoints

-   **POST**  `/api/chirps` - Creates a new chirp.
		- Expects an `Authorization` header with a `Bearer [JWT token]` value. 
		- Expects a JSON body with a `body` field (max 140 characters).
		- Returns a JSON body with the created chirp.
        
-   **GET**  `/api/chirps` - Retrieves a list of chirps.
		- Returns a JSON array of chirps.
		- Supports optional query parameters:
	-	`author_id`: Filters chirps by a specific user.
	-	`sort=desc`: Returns chirps in descending order by creation date.
        
-   **GET**  `/api/chirps/{chirpID}` - Retrieves a specific chirp by ID.
		-  Returns a JSON object of the chirp if found
        
-   **DELETE**  `/api/chirps/{chirpID}` - Deletes a chirp by ID.
		-   Expects an `Authorization` header with a `Bearer [JWT token]` value.
    
### Webhooks

-   **POST**  `/api/polka/webhooks` - Upgrades a user to 'chirpy red' if mock payment webhook is successful.


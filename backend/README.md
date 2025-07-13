# Backend Usage Instructions

This document provides instructions on how to set up, configure, and run the backend of the Async Multiplayer Management System.

## Prerequisites

- **Go**: Ensure you have Go installed on your system. You can download it from [https://golang.org/dl/](https://golang.org/dl/).

## Installation

1.  **Navigate to the backend directory**:
    ```bash
    cd backend
    ```

2.  **Install dependencies**:
    ```bash
    go mod tidy
    ```

## Configuration

This application uses Google OAuth2 for authentication. To run the backend, you must configure the following environment variables with your Google OAuth2 credentials:

- `GOOGLE_OAUTH_CLIENT_ID`: Your Google OAuth2 client ID.
- `GOOGLE_OAUTH_CLIENT_SECRET`: Your Google OAuth2 client secret.
- `GOOGLE_OAUTH_REDIRECT_URL`: The callback URL registered with your Google OAuth2 application (e.g., `http://localhost:8080/auth/google/callback`).

You can set these variables in your shell or using a `.env` file.

## Running the Application

To run the backend server, execute the following command from the `backend` directory:

```bash
go run main.go
```

The server will start on `0.0.0.0:8080` by default.

## API Endpoints

### Game Management

- **Create Game**: `POST /create-game`
  - Creates a new game. Requires a JSON body with a `name` field.
- **Join Game**: `POST /join-game/:id`
  - Allows a user to join an existing game. Requires a `User-ID` header.
- **Get Game Details**: `GET /games/:id`
  - Retrieves detailed information for a specific game, including a list of its players.

### Save File Handling

- **Upload Save**: `POST /games/:id/saves`
  - Uploads a new save file for a game. Requires a `User-ID` header and a multipart form with a `file` field.
- **Get Latest Save**: `GET /games/:id/saves/latest`
  - Downloads the most recent save file for a game. Requires a `User-ID` header.

### Authentication

- **Google Login**: `GET /auth/google/login`
  - Redirects the user to the Google OAuth2 consent page.
- **Google Callback**: `GET /auth/google/callback`
  - Handles the callback from Google after the user has granted consent.

## Testing

To run the test suite, execute the following command from the `backend` directory:

```bash
go test ./...
```

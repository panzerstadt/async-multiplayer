# Backend

This document provides instructions on how to set up and run the backend for local development. For deployment instructions, please see the main [DEPLOY.md](../../DEPLOY.md) file.

## Prerequisites

- **Go**: Ensure you have Go installed on your system. You can download it from [https://golang.org/dl/](https://golang.org/dl/).

## Getting Started

1.  **Navigate to the backend directory**:

    ```bash
    cd backend
    ```

2.  **Install dependencies**:

    ```bash
    go mod tidy
    ```

3.  **Configure Environment Variables**:

    This application uses Google OAuth2 for authentication. You must configure the following environment variables:

    - `GOOGLE_OAUTH_CLIENT_ID`: Your Google OAuth2 client ID.
    - `GOOGLE_OAUTH_CLIENT_SECRET`: Your Google OAuth2 client secret.
    - `GOOGLE_OAUTH_REDIRECT_URL`: The callback URL for local development (e.g., `http://localhost:8080/auth/google/callback`).

4.  **Run the Application**:
    ```bash
    go run main.go
    ```
    The server will start on `0.0.0.0:8080` by default.

## Testing

To run the test suite, execute the following command from the `backend` directory:

```bash
go test ./...
```

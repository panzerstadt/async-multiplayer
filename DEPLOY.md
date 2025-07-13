# Deployment Instructions

This document provides instructions for deploying the backend and frontend of the Async Multiplayer Management System.

## Backend Deployment

The backend is a Go application that handles game logic, user authentication, and file storage.

### Prerequisites

- Go installed on the deployment server.
- Access to a server or container environment where you can run the Go application.

### Configuration

The backend requires the following environment variables to be set for Google OAuth2 authentication:

- `GOOGLE_OAUTH_CLIENT_ID`: Your Google OAuth2 client ID.
- `GOOGLE_OAUTH_CLIENT_SECRET`: Your Google OAuth2 client secret.
- `GOOGLE_OAUTH_REDIRECT_URL`: The callback URL registered with your Google OAuth2 application. This should point to your deployed backend's callback endpoint (e.g., `https://your-backend-domain.com/auth/google/callback`).
- `FRONTEND_URL`: client, so that google callback goes to the right frontend

### Build and Run

1.  **Navigate to the backend directory**:

    ```bash
    cd backend
    ```

2.  **Build the application**:

    ```bash
    go build -o main .
    ```

3.  **Run the application**:
    ```bash
    ./main
    ```

The server will start on port `8080`. It is recommended to run the application behind a reverse proxy like Nginx or Caddy to handle HTTPS and load balancing.

## Frontend Deployment

The frontend is a Next.js application. The easiest way to deploy it is by using a platform like Vercel or Netlify, which are optimized for Next.js applications.

### Prerequisites

- Node.js and npm (or yarn/pnpm) installed on the deployment environment.
- A Vercel or Netlify account (recommended).

### Configuration

The frontend application needs to know the URL of the backend API. You must set the following environment variable:

- `NEXT_PUBLIC_API_URL`: The URL of your deployed backend (e.g., `https://your-backend-domain.com`).

When deploying to Vercel or Netlify, you can set this environment variable in the project settings.

### Deployment with Vercel

1.  Push your code to a Git repository (e.g., GitHub, GitLab, Bitbucket).
2.  Import your project into Vercel.
3.  Configure the project settings:
    - **Framework Preset**: Next.js
    - **Root Directory**: `frontend`
    - **Environment Variables**: Add `NEXT_PUBLIC_API_URL` with the URL of your backend.
4.  Deploy the application.

Vercel will automatically handle the build process and deploy your frontend.

### Manual Deployment

If you prefer to deploy the frontend manually, follow these steps:

1.  **Navigate to the frontend directory**:

    ```bash
    cd frontend
    ```

2.  **Install dependencies**:

    ```bash
    npm install
    ```

3.  **Build the application**:

    ```bash
    npm run build
    ```

4.  **Start the server**:
    ```bash
    npm start
    ```

The application will start on port `3000`. You should run it behind a reverse proxy for production use.

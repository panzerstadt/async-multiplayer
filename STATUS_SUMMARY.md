# Project Status Summary

This document summarizes the work completed on the backend and outlines the remaining tasks for both the backend and frontend.

## Backend Status

The backend has been significantly refactored for better organization, testability, and maintainability. The codebase is now stable, with a full suite of passing tests.

### Completed Tasks

- [x] **Modular Codebase Refactoring**: The initial monolithic `main.go` file has been broken down into a more organized and maintainable structure with dedicated packages for `game` logic, `handlers`, and `models`.
- [x] **Comprehensive Test Suite**: The entire test suite was refactored from a single file into smaller, feature-focused files (`game_creation`, `game_joining`, `game_details`, `oauth`, `notifications`, and `saves`). A centralized test setup utility was created to ensure test isolation and consistency. All tests are currently passing.
- [x] **Game Management API**: Implemented core API endpoints for game management:
  - `POST /create-game`: Creates a new game.
  - `POST /join-game/:id`: Allows a user to join an existing game.
  - `GET /games/:id`: Retrieves detailed information for a specific game, including a list of its players.
- [x] **Save File Handling API**: Implemented endpoints for managing game saves:
  - `POST /games/:id/saves`: Uploads a new save file for a game.
  - `GET /games/:id/saves/latest`: Downloads the most recent save file for a game.
- [x] **Dynamic Turn Order**: The `joinGameHandler` now automatically calculates and assigns the correct `TurnOrder` to new players based on the number of existing players in the game.
- [x] **Foundational OAuth2**: Implemented the initial Google OAuth2 authentication flow using the `golang.org/x/oauth2` library. This includes:
  - A `/auth/google/login` route to redirect users to the Google consent page.
  - A `/auth/google/callback` route to handle the callback from Google.
- [x] **Mock Notification System**: A `Notifier` interface has been defined, and a mock `EmailNotifier` has been implemented. This system is integrated into the `UploadSaveHandler` to log a notification when a new save is uploaded.

### Remaining Tasks

- [ ] **Full Authentication & Authorization**:
  - Implement session management (e.g., JWTs or secure cookies) to persist user sessions after a successful OAuth callback.
  - Create and apply an authentication middleware to secure all protected API endpoints, replacing the current placeholder `User-ID` header.
- [ ] **Production-Ready Notifications**:
  - Replace the current mock `EmailNotifier` with a real email service integration (e.g., SendGrid, Amazon SES) as outlined in the project instructions.
- [ ] **Deployment**:
  - Containerize the backend application using Docker.
  - Deploy the container to the specified Homelab environment and configure external access via Tailscale Funnel.
- [ ] **Configuration Management**:
  - Implement a more robust configuration management system to handle different environments (development, staging, production) instead of relying solely on `os.Getenv` calls within the code.

## Frontend Status

The frontend has not yet been started. The following tasks remain, based on the project instructions:

- [ ] **Initial UI Setup**: Scaffold the Next.js application.
- [ ] **Component Implementation**: Build out the necessary UI components using ShadCN for pages such as game creation, joining, and the main dashboard.
- [ ] **API Integration**: Connect the frontend to the newly created backend API endpoints.
- [ ] **Real-time Notifications**: Implement a client-side solution (e.g., WebSockets) to handle real-time turn notifications from the backend.
- [ ] **Deployment**: Deploy the completed frontend application to Vercel.

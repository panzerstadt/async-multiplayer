# Frontend

This document provides instructions on how to set up and run the frontend for local development. For deployment instructions, please see the main [DEPLOY.md](../../DEPLOY.md) file.

## Prerequisites

- [Node.js](https://nodejs.org/) (v18 or later)
- [npm](https://www.npmjs.com/) (or yarn/pnpm)

## Getting Started

1.  **Navigate to the frontend directory**:

    ```bash
    cd frontend
    ```

2.  **Install dependencies**:

    ```bash
    npm install
    ```

3.  **Configure Environment Variables**:

    Create a `.env.local` file in the `frontend` directory and add the following environment variable:

    ```
    NEXT_PUBLIC_API_URL=http://localhost:8080
    ```

4.  **Run the development server**:
    ```bash
    npm run dev
    ```

Open [http://localhost:3000](http://localhost:3000) with your browser to see the result.

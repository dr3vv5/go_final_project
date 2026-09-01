# Task Scheduler — Go Final Project

A graduation project built with Go, featuring a task management system with JWT-based authentication, SQLite persistence, and Docker deployment.

## Overview

This application is a simple yet robust task scheduler designed to demonstrate core backend development skills in Go. It allows users to:
*   Authenticate via a password stored in environment variables.
*   Create, read, update, and delete tasks (CRUD operations).
*   Set deadlines and recurrence rules for tasks.
*   Calculate the next execution date for recurring tasks.

The project showcases proficiency in:
*   **HTTP Routing:** Using `go-chi` for efficient request handling.
*   **Database Interaction:** Working with SQLite via the `database/sql` package.
*   **Security:** Implementing JWT authentication with password hashing.
*   **Containerization:** Multi-stage Docker builds for production-ready images.
*   **Testing:** Unit and integration tests for API endpoints.

## Features & Completed Challenges

### Core Features
*   **Password-Protected Access:** Authentication is enforced via the `TODO_PASSWORD` environment variable.
*   **JWT Token System:** Stateless authentication using signed tokens; tokens are invalidated automatically if the master password changes.
*   **Middleware Protection:** All API routes are secured by an auth middleware layer.
*   **Persistent Storage:** Tasks are stored in a local SQLite database file.

### Advanced Tasks (Starred Challenges)
*   **Docker Deployment:** Implemented a multi-stage `Dockerfile` to create a lightweight (~15MB) production image.
*   **Data Persistence Strategy:** Configured Docker volume mounting to ensure the SQLite database file (`scheduler.db`) persists on the host machine even after the container is removed.
*   **Environment-Based Configuration:** The application fully relies on environment variables (`TODO_PASSWORD`, `JWT_SECRET`, etc.) for configuration, following the 12-Factor App methodology.

---

## Prerequisites

To run this project, you need:
*   **Go** version 1.26 or higher.
*   **Docker** (optional, only for container deployment).
*   A terminal with shell access.

---

## Local Development Setup

### 1. Set Environment Variables

Before running the server, you must set the required environment variables.

**Option A: Using a `.env` file (Recommended)**
Create a file named `.env` in the root directory and add:
```bash
TODO_PASSWORD=SuperSecret123
JWT_SECRET=AnotherSuperSecretKey
TODO_PORT=7540
TODO_DBFILE=scheduler.db
```
*Note: Ensure `.env` is added to your `.gitignore` to prevent committing secrets.*

**Option B: Exporting in Terminal**

**Linux / macOS:**
```bash
export TODO_PASSWORD="SuperSecret123"
export JWT_SECRET="AnotherSuperSecretKey"
export TODO_PORT="7540"
export TODO_DBFILE="scheduler.db" 
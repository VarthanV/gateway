# Gateway

This project is a Gateway service implemented in Go using the HTTP module, Reverse proxy Gin framework. It includes features like load balancing, JWT authentication, logging, middleware handling, and dynamic management of backend servers. The service allows for adding, updating, and deleting backend servers via HTTP endpoints.

## Features

- **Load Balancing Logic**: Distributes traffic across multiple backend servers.
- **JWT Authentication**: Provides secure endpoints with JWT-based authentication.
- **Logging**: Logs requests and responses with custom log management, including writing logs with semaphores for concurrency control.
- **Middleware**: Includes default middleware for logging, error handling, and request validation.
- **Server Management**: Dynamically add, update, or delete backend servers.
- **Error Handling**: Handles various error cases like missing parameters, unsupported HTTP methods, and internal server errors.

## Getting Started

### Prerequisites

- Go 1.18 or higher
- Gin framework

### Installation

1. Clone the repository:
   ```bash
   git clone https://github.com/yourusername/your-repository.git
   cd your-repository
   ```

2. Install dependencies:
   ```bash
   go mod tidy
   ```

3. Run the application:
   ```bash
   go run main.go
   ```

The service will be available on `http://localhost:8080`.

## API Endpoints

### Add a Server
- **POST** `/servers`
  - Adds a new backend server to the specified path.
  - Request body: JSON representation of the server.

### Delete a Server
- **DELETE** `/servers`
  - Deletes an existing backend server.
  - Query Parameters: `path` (required), `url` (required).

### Update a Server
- **PUT** `/servers`
  - Updates an existing backend server.
  - Request body: JSON representation of the updated server.
  - Query Parameter: `path` (required).

### List Servers
- **GET** `/servers`
  - Lists all servers for the specified path.
  - Query Parameter: `path` (required).

### View Logs
- **GET** `/logs`
  - Fetches the log file content.
  - Note: The log file must be initialized for this endpoint to work.

## Configuration
The gateway and server management behavior is configured via environment variables and config files. For example, the JWT configuration and log file paths can be set in the config file. The config file is in ``toml`` format

### JWT Config
- **Secret Key**: Used to sign and verify JWT tokens.
- **Token Expiration**: Set the expiration time for the JWT tokens.

### Logging Configuration
- The logs are written using semaphores to handle concurrency.
- Log data includes request details and error messages.

## Contributing
If you find a bug or want to add new features, feel free to open an issue or create a pull request
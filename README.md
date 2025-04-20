# Microblogging

A simple microblogging service built in Go that allows users to create and manage posts. The project is designed to be easy to understand and extensible.

## Description

Microblogging is an application that allows users to register and manage posts on a platform similar to Twitter, with a focus on simplicity and best practices in backend development.

## Features

- **User registration**: Users can register on the platform.
- **Posts**: Users can create, read, edit, and delete posts.
- **Unit tests**: Unit tests have been implemented to ensure the system behaves as expected.

## Technologies Used

- **Go**: The primary language used for backend development.
- **PostgreSQL**: Relational database used for storing users and posts.
- **Docker**: For containerizing the application and the database, simplifying the development and deployment process.

## Installation

To run the project locally, follow these steps:

### Prerequisites

- Docker should be installed on your machine.
- Go 1.24 or higher should be installed.

### Installation Steps

1. Clone the repository:

    ```bash
    git clone https://github.com/RomiBareiro/microblogging.git
    cd microblogging
    ```

2. Build the Docker images:

    ```bash
    docker-compose build
    ```

3. Start the application services (Go and PostgreSQL) in Docker containers:

    ```bash
    docker-compose up
    ```

## API Usage

The application exposes several endpoints that allow users to interact with the service. Below are some of the main API endpoints, see swagger file

## Tests

To run the tests, use the following command:

```bash
go test ./...
```

## Contributing
If you would like to contribute to this project, you can do so by following these steps:

- Fork this repository.

- Create a branch for your new feature or bug fix (git checkout -b feature/new-feature).

- Make your changes and commit them (git commit -m 'Add new feature').

- Push your changes to your fork (git push origin feature/new-feature).

- Open a pull request for your changes to be reviewed and merged.

- Direct pushes to the `main` branch are not allowed.

- All changes must go through a Pull Request (PR).

- Every PR requires **at least one approval** before it can be merged.

- If new commits are pushed after approval, the PR must be reviewed again.

Take a look to the first PR I created to merge main to add evidences and get an approval to merge new functionalities: https://github.com/RomiBareiro/microblogging/pull/1


## Feel free to contact me

bareiro.romina@gmail.com

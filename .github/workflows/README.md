# Workflows

### Pull Request

This runs tests when a PR is created, typically from your working branch to develop or main.

### Merge

When you merge a PR to develop then this will build and push a _develop_ image to DockerHub.

### Release

When you create a new tag on main then this will build and push a _latest_ image to DockerHub.

# git fit

As a GitHub user, perhaps you have tried to change your GitHub avatar and realized there is a 1MB limit for images you can upload. Git Fit's purpose is to provide a tool in your command line (CLI) to compress your avatar while maintaining high quality output.

> NOTE!
> GitHub does not allow updating avatars through the GitHub API. GitHub avatars are generated and hosted from Gravatar (https://gravatar.com). The only other option is to use Gravatar's API endpoint to upload new avatar images (https://api.gravatar.com/v3/me/avatars).

- [ ] **TO DO: Connect to Gravatar API to update avatars.**

## useful gravatar links:
1. https://docs.gravatar.com/rest/getting-started/
2. https://gravatar.com/developers/console

## using the tool in its current state:
gitfit -input input.jpeg -output output.jpeg -maxsize <max bytes> -quality <1-100 for jpeg> -v [for verbose output]
```

## Running the Web App

You can run the full stack application using the Makefile:

1. ** start the backend server **:
   ```bash
   make server
   ```
   The server will start on `http://localhost:8080`.

2. ** start the frontend development server **:
   ```bash
   make web
   ```
   The frontend will start on `http://localhost:5173`.

3. ** run tests **:
   ```bash
   make test
   ```

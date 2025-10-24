# git fit

As a GitHub user, perhaps you have tried to change your GitHub avatar and realized there is a 1MB limit for images you can upload. Git Fit's purpose is to provide a tool in your command line interface (CLI) to compress your avatar of choice while maintaining high quality.

> NOTE!
> GitHub does not allow updating avatars through the GitHub API. GitHub avatars are generated and hosted from Gravatar (https://gravatar.com). The only other option is to use Gravatar's API endpoint to upload new avatar images (https://api.gravatar.com/v3/me/avatars).

- [ ] **TO DO: Connect to Gravatar API to update avatars.**

## useful gravatar links:
1. https://docs.gravatar.com/rest/getting-started/
2. https://gravatar.com/developers/console

## using the tool in its current state:
```bash
go run ./cmd/gitfit/main.go -input input.jpeg -output output.jpeg -quality <1-100 for jpeg> -v [for verbose output]

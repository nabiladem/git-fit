# git fit

![Git Fit](https://img.shields.io/badge/git--fit-online-brightgreen?logo=vercel&logoColor=white)

As a GitHub user, you may have tried to update your avatar and realized there is a strict 1MB file size limit. Git Fit provides a convenient way to compress your images while maintaining high-quality output, and is available as both a command-line tool (CLI) and a web application.

![Preview](preview.png)

## Stack

- ![Go](https://img.shields.io/badge/go-%2300ADD8.svg?style=for-the-badge&logo=go&logoColor=white)
- ![React](https://img.shields.io/badge/react-%2320232a.svg?style=for-the-badge&logo=react&logoColor=%2361DAFB)
- ![Vite](https://img.shields.io/badge/vite-%23646CFF.svg?style=for-the-badge&logo=vite&logoColor=white)
- ![TailwindCSS](https://img.shields.io/badge/tailwindcss-%2338B2AC.svg?style=for-the-badge&logo=tailwind-css&logoColor=white)

## Using the Command Line Tool

```bash
gitfit -input input.jpeg -output output.jpeg -maxsize <max bytes> -quality <1-100 for jpeg> -v [for verbose output]
```

## Running the Web App

You can run the fullstack application using the provided `Makefile`:

1. **Start the backend server**:
   ```bash
   make server
   ```
   The server will start on `http://localhost:8080`.

2. **Start the frontend development server**:
   ```bash
   make web
   ```
   The frontend will start on `http://localhost:5173`.

3. **Run tests**:
   ```bash
   make test
   ```

# RootAccess CLI Guide

The **rootaccess** CLI allows you to interact with the platform directly from your terminal.

## Installation

You can install the CLI using the following command:

```bash
curl -s https://ctf.rootaccess.live/install.sh | bash
```

Alternatively, you can build it from source:

```bash
cd cli
go build -o rootaccess main.go
sudo mv rootaccess /usr/local/bin/
```

## Authentication

RootAccess uses Google OAuth for authentication. To log in from the CLI:

1. Run the login command:
   ```bash
   rootaccess login
   ```
2. Your browser will open to the Google Login page.
3. After successful login, you will see a success page with your access token.
4. Copy the token and paste it back into your terminal.

## Commands

### List Challenges
```bash
rootaccess challenges
```

### View Challenge Details
```bash
rootaccess open <ID>
```

### Submit a Flag
```bash
rootaccess submit <ID> <FLAG>
```

### View Scoreboard
```bash
rootaccess scoreboard
```

### User Info
```bash
rootaccess whoami
```

### Logout
```bash
rootaccess logout
```

## Developer Notes

### Binaries
When deploying, build binaries for all major platforms:

```bash
GOOS=linux GOARCH=amd64 go build -o rootaccess-linux-amd64
GOOS=linux GOARCH=arm64 go build -o rootaccess-linux-arm64
GOOS=darwin GOARCH=amd64 go build -o rootaccess-darwin-amd64
GOOS=darwin GOARCH=arm64 go build -o rootaccess-darwin-arm64
GOOS=windows GOARCH=amd64 go build -o rootaccess-windows-amd64.exe
```

Upload these to your hosting provider under the `/bin/` path.

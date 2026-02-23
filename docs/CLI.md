# RootAccess CLI

`rootaccess` is the official command-line tool for the RootAccess CTF platform. It lets you browse challenges, submit flags, and check the scoreboard without leaving your terminal.

---

## Installation

### One-line installer (Linux / macOS)

```bash
curl -s https://ctf.rootaccess.live/install.sh | bash
```

The script downloads the right binary for your OS/architecture and places it in `/usr/local/bin/rootaccess`.

### Build from source

```bash
git clone https://github.com/Uttam-Mahata/RootAccess.git
cd RootAccess/cli
go build -o rootaccess main.go
sudo mv rootaccess /usr/local/bin/
```

### Cross-compile for other platforms

```bash
GOOS=linux   GOARCH=amd64 go build -o rootaccess-linux-amd64
GOOS=linux   GOARCH=arm64 go build -o rootaccess-linux-arm64
GOOS=darwin  GOARCH=amd64 go build -o rootaccess-darwin-amd64
GOOS=darwin  GOARCH=arm64 go build -o rootaccess-darwin-arm64
GOOS=windows GOARCH=amd64 go build -o rootaccess-windows-amd64.exe
```

---

## Configuration

The CLI stores its config (token, username, server URLs) at:

```
~/.config/rootaccess/config.json
```

You never need to edit this file directly — `login` and `logout` manage it automatically. Default API endpoint is `https://ctf.rootaccess.live/api`.

---

## Commands

### `login` — Authenticate via browser

```bash
rootaccess login
```

Opens your browser to the platform's OAuth login page. After you authenticate, the browser sends the session token back to a short-lived local server started by the CLI. No copy-pasting required.

**What happens step by step:**

1. CLI starts a temporary HTTP server on a random local port (e.g. `41253`).
2. Prints the auth URL and attempts to open your default browser:
   ```
   Opening your browser to authenticate...
   If the browser doesn't open automatically, visit: https://ctf.rootaccess.live/cli/auth?port=41253
   ```
3. You log in (or are already logged in) on the web page.
4. The browser navigates to `http://127.0.0.1:41253/callback?token=<JWT>`.
5. CLI verifies the token with the API, saves it, and prints:
   ```
   Successfully logged in as Uttam435!
   ```

The login attempt times out after **5 minutes** if the browser flow is not completed.

---

### `whoami` — Show current user

```bash
rootaccess whoami
```

Verifies the stored token with the API and prints your account details.

**Example output:**

```
Logged in as:
Username: Uttam435
Email:    uttam@example.com
Role:     user
```

**Error — not logged in:**

```
You are not logged in. Run 'rootaccess login' first.
```

---

### `challenges` — List all challenges

```bash
rootaccess challenges
```

Fetches all available challenges and prints them as an aligned table. Solved challenges are marked with `●`.

**Example output:**

```
ID                        TITLE                  CATEGORY    POINTS   SOLVES   STATUS
64f1a2b3c4d5e6f7a8b9c0d1   Baby Buffer Overflow   pwn         450      12       ○
64f1a2b3c4d5e6f7a8b9c0d2   SQLi 101               web         200      47       ● SOLVED
64f1a2b3c4d5e6f7a8b9c0d3   Hidden in Plain Sight  forensics   350      8        ○

Use 'rootaccess open <ID>' to view challenge details.
```

Points shown are **dynamic** — they decrease as more teams solve the challenge (CTFd-style scoring).

---

### `open` — View challenge details

```bash
rootaccess open <ID>
```

Fetches the full challenge detail: description, tags, attached files, and current point value.

**Example:**

```bash
rootaccess open 64f1a2b3c4d5e6f7a8b9c0d1
```

**Example output:**

```
=== BABY BUFFER OVERFLOW ===
ID:         64f1a2b3c4d5e6f7a8b9c0d1
Category:   pwn
Difficulty: easy
Points:     450
Status:     UNSOLVED ○
Tags:       buffer-overflow, x86-64, ret2win

DESCRIPTION:
Can you smash the stack and redirect execution to win()?
The binary is running at chall.rootaccess.live:9001.

FILES:
- https://ctf.rootaccess.live/files/bof/vuln
- https://ctf.rootaccess.live/files/bof/vuln.c

Submit flag:
  rootaccess submit 64f1a2b3c4d5e6f7a8b9c0d1 <FLAG>
```

---

### `submit` — Submit a flag

```bash
rootaccess submit <ID> <FLAG>
```

Submits a flag for a challenge. Flags are rate-limited to **5 attempts per minute** per challenge.

**Example — correct flag (first solve):**

```bash
rootaccess submit 64f1a2b3c4d5e6f7a8b9c0d1 'rootaccess{buff3r_0v3rfl0w_ftw}'
```

```
● CORRECT! Flag is valid. You earned 450 points!
```

**Example — correct flag (already solved):**

```
● Correct! But you already solved this challenge.
```

**Example — wrong flag:**

```
○ INCORRECT. Try again!
```

**Tip:** Wrap flags in single quotes to prevent your shell from interpreting braces or special characters.

---

### `scoreboard` — View the live scoreboard

```bash
rootaccess scoreboard
```

Fetches the scoreboard for the currently active contest and prints a ranked table.

**Example output:**

```
=== SCOREBOARD: RootAccess Spring CTF 2026 ===

RANK   USERNAME      TEAM          POINTS
1      Uttam435      ShellStorm    1250
2      h4ck3rX       -             980
3      r3v3rs3r      ByteWolves    870
```

`-` in the Team column means the player is competing individually (no team registered for this contest).

---

### `logout` — Clear saved credentials

```bash
rootaccess logout
```

Clears the stored token and username from `~/.config/rootaccess/config.json`. Does not invalidate the session server-side.

**Example output:**

```
Logged out successfully.
```

---

## Typical workflow

```bash
# 1. Authenticate
rootaccess login

# 2. Browse challenges
rootaccess challenges

# 3. Read a challenge
rootaccess open 64f1a2b3c4d5e6f7a8b9c0d1

# 4. Work the challenge... then submit your flag
rootaccess submit 64f1a2b3c4d5e6f7a8b9c0d1 'rootaccess{y0ur_fl4g_h3r3}'

# 5. Check your ranking
rootaccess scoreboard

# 6. Confirm your identity at any time
rootaccess whoami

# 7. Log out when done
rootaccess logout
```

---

## Troubleshooting

| Symptom | Likely cause | Fix |
|---------|-------------|-----|
| `You must be logged in` on any command | Token missing or expired | Run `rootaccess login` |
| Browser doesn't open | No default browser configured | Copy the printed URL and open it manually |
| Login times out | Browser flow not completed within 5 min | Run `rootaccess login` again |
| `API error (401)` | Token expired | Run `rootaccess login` again |
| `API error (429)` on `submit` | Rate limit hit (5/min per challenge) | Wait a minute and try again |
| Flag with special characters rejected by shell | Shell interpreting `{` `}` | Wrap the flag in single quotes: `'rootaccess{...}'` |

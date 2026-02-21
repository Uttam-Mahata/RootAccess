# Admin Management Tool

## 🔐 Security Overview

**✅ By Default: All registered users are regular users (NOT admins)**

- Registration endpoint: `/auth/register` creates users with role `"user"`
- There is NO way to register as admin through the API
- Role field is hardcoded in the backend and cannot be overridden
- Admin promotion must be done manually using this tool

## 🛠️ Admin Tool Usage

This CLI tool allows you to manage admin users for your RootAccess CTF platform.

### Build the Tool

```bash
cd backend
go build -o admin-tool cmd/admin/main.go
```

### Run the Tool

```bash
./admin-tool
```

### Available Options

```
╔═══════════════════════════════════════════╗
║   RootAccess CTF - Admin Management      ║
╚═══════════════════════════════════════════╝

1. Create Admin User
2. Promote User to Admin
3. Demote Admin to User
4. List All Users
5. Exit
```

## 📋 Common Tasks

### 1. Create Initial Admin User

When setting up the platform for the first time:

```bash
./admin-tool
# Choose option 1
# Enter username, email, and password
```

**Example:**
```
Choose an option: 1

=== Create Admin User ===
Username: rootadmin
Email: admin@rootaccess.ctf
Password: YourSecurePassword123!

✅ Admin user 'rootadmin' created successfully!
   Email: admin@rootaccess.ctf
   Role: admin
   Email verified: true
```

### 2. Promote Existing User to Admin

If someone registered normally and you want to make them an admin:

```bash
./admin-tool
# Choose option 2
# Enter their username or email
```

**Example:**
```
Choose an option: 2

=== Promote User to Admin ===
Enter username or email: john.doe@example.com

✅ User 'johndoe' promoted to admin!
   Email: john.doe@example.com
   New Role: admin
```

### 3. Demote Admin to Regular User

To remove admin privileges:

```bash
./admin-tool
# Choose option 3
# Enter their username or email
```

**Example:**
```
Choose an option: 3

=== Demote Admin to User ===
Enter username or email: oldadmin

✅ User 'oldadmin' demoted to regular user!
   Email: old@example.com
   New Role: user
```

### 4. List All Users

View all users and their roles:

```bash
./admin-tool
# Choose option 4
```

**Example Output:**
```
=== All Users ===

USERNAME             EMAIL                          ROLE       VERIFIED       
────────────────────────────────────────────────────────────────────────────────
rootadmin            admin@rootaccess.ctf           🔑 admin   Yes            
johndoe              john@example.com               🔑 admin   Yes            
alice                alice@example.com              user       Yes            
bob                  bob@example.com                user       No             

Total users: 4
```

## 🔒 Security Best Practices

1. **Protect the Admin Tool**
   - Only run on trusted machines
   - Don't commit compiled binary to version control
   - Store securely with restricted permissions

2. **Limit Admin Accounts**
   - Only create admins when absolutely necessary
   - Use strong passwords for admin accounts
   - Regularly audit admin access

3. **Monitor Admin Actions**
   - Keep track of who has admin access
   - Review admin activity regularly
   - Remove admin access when no longer needed

## 🐳 Docker Usage

If running in Docker, you can execute the tool inside the container:

```bash
# Build the tool inside container
docker exec -it backend_container bash
cd /app
go build -o admin-tool cmd/admin/main.go
./admin-tool

# Or one-liner
docker exec -it backend_container sh -c "cd /app && go run cmd/admin/main.go"
```

## 📝 Environment Requirements

The tool reads from the same `.env` file as your backend:
- `MONGO_URI`: MongoDB connection string
- `DB_NAME`: Database name

Make sure these are configured before running the tool.

## 🔧 Technical Details

### Registration Security

**File: `internal/services/auth_service.go`**
```go
// Line 60-70: Role is HARDCODED to "user"
user := &models.User{
    Username:           username,
    Email:              email,
    PasswordHash:       string(hashedPassword),
    Role:               "user",  // ← Cannot be changed via API
    EmailVerified:      false,
    VerificationToken:  token,
    VerificationExpiry: s.emailService.GetVerificationExpiry(),
    CreatedAt:          time.Now(),
    UpdatedAt:          time.Now(),
}
```

**File: `internal/handlers/auth_handler.go`**
```go
// Line 20-24: No role field in registration request
type RegisterRequest struct {
    Username string `json:"username" binding:"required,min=3,max=50"`
    Email    string `json:"email" binding:"required,email"`
    Password string `json:"password" binding:"required,min=8"`
    // No role field - cannot be set by user
}
```

### Admin Creation

Admins can ONLY be created through:
1. This CLI tool
2. Direct database manipulation

There is NO API endpoint that allows admin creation.

## ⚠️ Important Notes

- Admin users created via this tool have their email **automatically verified**
- Regular registered users must verify their email before logging in
- Promoting a user to admin does NOT change their email verification status
- Always use this tool with appropriate database backups

## 🆘 Troubleshooting

### "User not found"
- Check spelling of username/email
- Use option 4 to list all users first

### "Failed to connect to MongoDB"
- Verify `.env` file exists and has correct MONGO_URI
- Ensure MongoDB is running
- Check network connectivity

### Permission Errors
- Ensure you have write permissions in the backend directory
- Run with appropriate user permissions

## 📞 Support

For issues or questions, refer to the main project documentation or create an issue in the repository.

#!/bin/bash

# Script to set up sample data for the CTF platform

echo "Setting up sample data..."

# Check if user exists, create if not, then set as admin
update_result=$(docker exec ctf-mongodb mongosh go_ctf --quiet --eval '
const existingUser = db.users.findOne({ username: "testuser" });
if (!existingUser) {
  print("USER_NOT_FOUND");
  quit(1);
}
const result = db.users.updateOne(
  { username: "testuser" },
  { $set: { role: "admin" } }
);
print("SUCCESS");
' 2>&1)

if echo "$update_result" | grep -q "USER_NOT_FOUND"; then
  echo ""
  echo "⚠️  User 'testuser' does not exist in the database."
  echo ""
  echo "Please follow these steps:"
  echo "  1. Open http://localhost:4200 in your browser"
  echo "  2. Register a new account with username: testuser"
  echo "  3. Choose a secure password (suggestion: generate with 'openssl rand -base64 12')"
  echo "  4. Run this script again to promote the user to admin"
  echo ""
  exit 1
fi

echo "✓ Made testuser an admin"

# Create sample challenges
docker exec ctf-mongodb mongosh go_ctf --eval '
db.challenges.insertMany([
  {
    title: "Welcome Challenge",
    description: "Find the flag hidden in plain sight. Hint: Check the HTML comments!",
    category: "Web",
    points: 100,
    flag: "FLAG{welcome_to_ctf}",
    files: []
  },
  {
    title: "Simple Caesar",
    description: "Decode this message: SYNT{pnrfne_pvcure}. ROT13 might help!",
    category: "Crypto",
    points: 150,
    flag: "FLAG{caesar_cipher}",
    files: []
  },
  {
    title: "Binary Basics",
    description: "Convert this binary to ASCII: 01000110 01001100 01000001 01000111 01111011 01100010 01101001 01101110 01100001 01110010 01111001 01111101",
    category: "Reverse Engineering",
    points: 200,
    flag: "FLAG{binary}",
    files: []
  },
  {
    title: "SQL Injection 101",
    description: "Find a way to bypass the login form. The flag is in the database!",
    category: "Web",
    points: 250,
    flag: "FLAG{sql_injection_master}",
    files: []
  },
  {
    title: "Hidden Message",
    description: "The flag is encoded in base64: RkxBR3toaWRkZW5fbWVzc2FnZX0=",
    category: "Forensics",
    points: 150,
    flag: "FLAG{hidden_message}",
    files: []
  }
])
'

echo "✓ Created sample challenges"
echo ""
echo "========================================="
echo "✅ Setup Complete!"
echo "========================================="
echo ""
echo "Sample challenges have been added to the database."
echo "User 'testuser' has been promoted to admin."
echo ""
echo "⚠️  SECURITY NOTICE:"
echo "This script is for DEVELOPMENT ONLY."
echo "For production:"
echo "  • Create admin accounts through secure channels"
echo "  • Use strong, unique passwords (not 'password123')"
echo "  • Never use default test credentials"
echo "========================================="

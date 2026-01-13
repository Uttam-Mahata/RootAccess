#!/bin/bash

# Script to set up sample data for the CTF platform

echo "Setting up sample data..."

# Create admin user (update existing testuser to admin)
docker exec ctf-mongodb mongosh go_ctf --eval '
db.users.updateOne(
  { username: "testuser" },
  { $set: { role: "admin" } }
)
'

echo "Made testuser an admin"

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

echo "Created sample challenges"
echo "Setup complete!"
echo ""
echo "Admin credentials:"
echo "Username: testuser"
echo "Password: password123"

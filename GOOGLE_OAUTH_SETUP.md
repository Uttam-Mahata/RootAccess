# Google OAuth Setup Guide

This guide will walk you through setting up Google OAuth authentication for the RootAccess CTF Platform.

## Prerequisites

- A Google Account
- Access to [Google Cloud Console](https://console.cloud.google.com)

## Step 1: Create a Google Cloud Project

1. Go to the [Google Cloud Console](https://console.cloud.google.com)
2. Click on the project dropdown at the top of the page
3. Click **"New Project"**
4. Enter a project name (e.g., "RootAccess CTF")
5. Click **"Create"**
6. Wait for the project to be created and select it

## Step 2: Enable Required APIs

1. In the Google Cloud Console, navigate to **"APIs & Services"** > **"Library"**
2. Search for **"Google+ API"** (or "Google People API")
3. Click on it and click **"Enable"**
4. Alternatively, you can use the userinfo endpoint which doesn't require enabling additional APIs

## Step 3: Configure OAuth Consent Screen

1. Navigate to **"APIs & Services"** > **"OAuth consent screen"**
2. Select **"External"** user type (unless you have a Google Workspace)
3. Click **"Create"**

### Fill in the App Information:

- **App name**: RootAccess CTF Platform
- **User support email**: Your email address
- **App logo**: (Optional) Upload your CTF platform logo
- **App domain**: (Optional) Your website domain
- **Authorized domains**: 
  - Add `localhost` for development (if allowed)
  - Add your production domain (e.g., `148.100.79.152` or your actual domain)
- **Developer contact information**: Your email address

4. Click **"Save and Continue"**

### Add Scopes:

1. Click **"Add or Remove Scopes"**
2. Select the following scopes:
   - `.../auth/userinfo.email` - View your email address
   - `.../auth/userinfo.profile` - See your personal info
3. Click **"Update"**
4. Click **"Save and Continue"**

### Add Test Users (Development):

1. Click **"Add Users"**
2. Add email addresses of users who can test the OAuth flow during development
3. Click **"Save and Continue"**

4. Review the summary and click **"Back to Dashboard"**

## Step 4: Create OAuth 2.0 Credentials

1. Navigate to **"APIs & Services"** > **"Credentials"**
2. Click **"Create Credentials"** > **"OAuth 2.0 Client ID"**
3. Select **"Web application"** as the application type

### Configure the OAuth Client:

**Name**: RootAccess CTF Backend

**Authorized JavaScript origins** (optional for backend):
- `http://localhost:4200` (development frontend)
- `http://148.100.79.152` (production frontend, if applicable)

**Authorized redirect URIs** (IMPORTANT):
- Development: `http://localhost:8080/auth/google/callback`
- Production: `http://148.100.79.152:8080/auth/google/callback`

> **Note**: The redirect URI must exactly match the `GOOGLE_REDIRECT_URL` in your backend `.env` file.

4. Click **"Create"**

## Step 5: Copy Credentials

After creating the OAuth client, you'll see a dialog with your credentials:

1. **Client ID**: Copy this (looks like: `123456789-abc123def456.apps.googleusercontent.com`)
2. **Client Secret**: Copy this (looks like: `GOCSPX-abc123def456ghi789`)

You can also find these credentials later:
- Go to **"APIs & Services"** > **"Credentials"**
- Click on your OAuth 2.0 Client ID name
- View the Client ID and Client Secret

## Step 6: Update Backend Environment Variables

Edit your backend `.env` file and add the following:

```env
GOOGLE_CLIENT_ID=your-client-id.apps.googleusercontent.com
GOOGLE_CLIENT_SECRET=your-client-secret
GOOGLE_REDIRECT_URL=http://localhost:8080/auth/google/callback
```

For production, update `GOOGLE_REDIRECT_URL` to your production backend URL:

```env
GOOGLE_REDIRECT_URL=http://148.100.79.152:8080/auth/google/callback
```

## Step 7: Test the OAuth Flow

1. Restart your backend server to load the new environment variables
2. Navigate to your frontend login page
3. Click **"Continue with Google"**
4. You should be redirected to Google's consent screen
5. Select your Google account and approve the permissions
6. You should be redirected back to your application and logged in

### Troubleshooting:

**Error: "redirect_uri_mismatch"**
- The redirect URI in your backend doesn't match what's configured in Google Cloud Console
- Check that `GOOGLE_REDIRECT_URL` in `.env` matches exactly with the authorized redirect URI in Google Cloud Console

**Error: "Access blocked: This app's request is invalid"**
- Your OAuth consent screen is not properly configured
- Make sure you've added the required scopes (email and profile)

**Error: "403: org_internal"**
- Your app is set to "Internal" but you're trying to sign in with an external Google account
- Change the OAuth consent screen to "External"

**Error: "403: access_denied"**
- If using "Internal" mode, the user is not part of your organization
- If using "External" mode in testing, add the user as a test user in the OAuth consent screen

**Error: Invalid state token**
- CSRF protection failed
- Check that Redis is running and properly configured
- Make sure the state token hasn't expired (10 minute expiry by default)

## Step 8: Publish Your App (Optional - For Production)

During development, your OAuth consent screen is in "Testing" mode, which limits OAuth to 100 test users.

To make your app available to all Google users:

1. Navigate to **"APIs & Services"** > **"OAuth consent screen"**
2. Review all the information to ensure it's accurate
3. Add links to your Privacy Policy and Terms of Service (required for publication)
4. Click **"Publish App"**
5. Your app will be submitted for verification by Google

> **Note**: For small CTF platforms, you may not need to publish the app. You can keep it in testing mode and just add test users as needed (up to 100 users).

## Security Best Practices

1. **Never commit credentials**: Keep your `GOOGLE_CLIENT_SECRET` in `.env` and never commit it to version control
2. **Use HTTPS in production**: Set `Secure: true` for cookies in production (requires HTTPS)
3. **Rotate secrets**: If your client secret is ever exposed, regenerate it in Google Cloud Console
4. **Limit scopes**: Only request the minimum scopes needed (email and profile)
5. **Monitor usage**: Check the Google Cloud Console regularly for any suspicious API usage

## Additional Resources

- [Google OAuth 2.0 Documentation](https://developers.google.com/identity/protocols/oauth2)
- [OAuth 2.0 for Web Server Applications](https://developers.google.com/identity/protocols/oauth2/web-server)
- [Google OAuth Playground](https://developers.google.com/oauthplayground/) - Test OAuth flows

## Support

If you encounter any issues:
1. Check the backend logs for detailed error messages
2. Verify that all redirect URIs match exactly
3. Ensure Redis is running (required for state token storage)
4. Check that your Google Cloud project has the required APIs enabled

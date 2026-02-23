import { Routes } from '@angular/router';
import { Router } from '@angular/router';
import { inject } from '@angular/core';
import { AuthService } from './services/auth';
import { map, filter, take } from 'rxjs/operators';

// Guard for protected routes - requires login
const authGuard = () => {
  const authService = inject(AuthService);
  const router = inject(Router);

  // Wait for auth check to complete, then check if user is logged in
  return authService.authCheckComplete$.pipe(
    filter(complete => complete),
    take(1),
    map(() => {
      if (authService.isLoggedIn()) {
        return true;
      }
      router.navigate(['/login']);
      return false;
    })
  );
};

// Guard for admin routes
const adminGuard = () => {
  const authService = inject(AuthService);
  const router = inject(Router);

  return authService.authCheckComplete$.pipe(
    filter(complete => complete),
    take(1),
    map(() => {
      if (authService.isAdmin()) {
        return true;
      }
      router.navigate(['/home']);
      return false;
    })
  );
};

// Guard for guest routes (login, register) - redirects logged-in users to home
const guestGuard = () => {
  const authService = inject(AuthService);
  const router = inject(Router);

  return authService.authCheckComplete$.pipe(
    filter(complete => complete),
    take(1),
    map(() => {
      if (authService.isLoggedIn()) {
        router.navigate(['/home']);
        return false;
      }
      return true;
    })
  );
};

// Guard for landing page - redirects authenticated users to home
const landingGuard = () => {
  const authService = inject(AuthService);
  const router = inject(Router);

  return authService.authCheckComplete$.pipe(
    filter(complete => complete),
    take(1),
    map(() => {
      if (authService.isLoggedIn()) {
        router.navigate(['/home']);
        return false;
      }
      return true;
    })
  );
};

export const routes: Routes = [
  // Guest routes - redirect to home if logged in
  { path: 'login', loadComponent: () => import('./components/login/login').then(m => m.LoginComponent), canActivate: [guestGuard] },
  { path: 'register', loadComponent: () => import('./components/register/register').then(m => m.RegisterComponent), canActivate: [guestGuard] },

  // Public routes (no guards)
  { path: 'verify-email', loadComponent: () => import('./components/verify-email/verify-email').then(m => m.VerifyEmailComponent) },
  { path: 'forgot-password', loadComponent: () => import('./components/forgot-password/forgot-password').then(m => m.ForgotPasswordComponent) },
  { path: 'reset-password', loadComponent: () => import('./components/reset-password/reset-password').then(m => m.ResetPasswordComponent) },
  { path: 'auth/callback', loadComponent: () => import('./components/oauth-callback/oauth-callback').then(m => m.OAuthCallbackComponent) },
  { path: 'cli', loadComponent: () => import('./components/cli-download/cli-download').then(m => m.CLIDownloadComponent) },
  { path: 'cli/auth', loadComponent: () => import('./components/cli-auth/cli-auth').then(m => m.CLIAuthComponent) },
  { path: 'legal', loadComponent: () => import('./components/legal/legal').then(m => m.LegalComponent) },
  { path: 'profile/:username', loadComponent: () => import('./components/user-profile/user-profile').then(m => m.UserProfileComponent) },

  // Protected routes - require login
  { path: 'home', loadComponent: () => import('./components/home/home').then(m => m.HomeComponent), canActivate: [authGuard] },
  { path: 'scoreboard', loadComponent: () => import('./components/scoreboard/scoreboard').then(m => m.ScoreboardComponent), canActivate: [authGuard] },
  { path: 'settings', loadComponent: () => import('./components/account-settings/account-settings').then(m => m.AccountSettingsComponent), canActivate: [authGuard] },
  { path: 'team', loadComponent: () => import('./components/team-dashboard/team-dashboard').then(m => m.TeamDashboardComponent), canActivate: [authGuard] },
  { path: 'challenges', loadComponent: () => import('./components/challenge-list/challenge-list').then(m => m.ChallengeListComponent), canActivate: [authGuard] },
  { path: 'challenges/:id', loadComponent: () => import('./components/challenge-detail/challenge-detail').then(m => m.ChallengeDetailComponent), canActivate: [authGuard] },
  { path: 'activity', loadComponent: () => import('./components/activity-dashboard/activity-dashboard').then(m => m.ActivityDashboardComponent), canActivate: [authGuard] },
  { path: 'notifications', loadComponent: () => import('./components/notifications/notifications').then(m => m.NotificationsComponent), canActivate: [authGuard] },

  // Admin routes
  { path: 'admin', loadComponent: () => import('./components/admin-dashboard/admin-dashboard').then(m => m.AdminDashboardComponent), canActivate: [adminGuard] },

  // Default route - show landing page for guests, redirect authenticated users to home
  { path: '', loadComponent: () => import('./components/landing/landing').then(m => m.LandingComponent), canActivate: [landingGuard] },
];

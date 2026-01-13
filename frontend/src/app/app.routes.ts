import { Routes } from '@angular/router';
import { Router } from '@angular/router';
import { LoginComponent } from './components/login/login';
import { RegisterComponent } from './components/register/register';
import { ChallengeListComponent } from './components/challenge-list/challenge-list';
import { ChallengeDetailComponent } from './components/challenge-detail/challenge-detail';
import { ScoreboardComponent } from './components/scoreboard/scoreboard';
import { AdminDashboardComponent } from './components/admin-dashboard/admin-dashboard';
import { VerifyEmailComponent } from './components/verify-email/verify-email';
import { AccountSettingsComponent } from './components/account-settings/account-settings';
import { ForgotPasswordComponent } from './components/forgot-password/forgot-password';
import { ResetPasswordComponent } from './components/reset-password/reset-password';
import { TeamDashboardComponent } from './components/team-dashboard/team-dashboard';
import { HomeComponent } from './components/home/home';
import { UserProfileComponent } from './components/user-profile/user-profile';
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
      router.navigate(['/']);
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
        router.navigate(['/']);
        return false;
      }
      return true;
    })
  );
};

export const routes: Routes = [
  // Guest routes - redirect to home if logged in
  { path: 'login', component: LoginComponent, canActivate: [guestGuard] },
  { path: 'register', component: RegisterComponent, canActivate: [guestGuard] },
  
  // Public routes (no guards)
  { path: 'verify-email', component: VerifyEmailComponent },
  { path: 'forgot-password', component: ForgotPasswordComponent },
  { path: 'reset-password', component: ResetPasswordComponent },
  { path: 'scoreboard', component: ScoreboardComponent },
  { path: 'profile/:username', component: UserProfileComponent },
  
  // Protected routes - require login
  { path: 'home', component: HomeComponent, canActivate: [authGuard] },
  { path: 'settings', component: AccountSettingsComponent, canActivate: [authGuard] },
  { path: 'team', component: TeamDashboardComponent, canActivate: [authGuard] },
  { path: 'challenges', component: ChallengeListComponent, canActivate: [authGuard] },
  { path: 'challenges/:id', component: ChallengeDetailComponent, canActivate: [authGuard] },
  
  // Admin routes
  { path: 'admin', component: AdminDashboardComponent, canActivate: [adminGuard] },
  
  // Default route - show home if logged in, login if not
  { path: '', component: HomeComponent, canActivate: [authGuard] },
];

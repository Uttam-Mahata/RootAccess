import { Routes } from '@angular/router';
import { LoginComponent } from './components/login/login';
import { RegisterComponent } from './components/register/register';
import { ChallengeListComponent } from './components/challenge-list/challenge-list';
import { ChallengeDetailComponent } from './components/challenge-detail/challenge-detail';
import { ScoreboardComponent } from './components/scoreboard/scoreboard';
import { AdminDashboardComponent } from './components/admin-dashboard/admin-dashboard';
import { inject } from '@angular/core';
import { AuthService } from './services/auth';

const authGuard = () => {
  const authService = inject(AuthService);
  if (authService.isLoggedIn()) {
    return true;
  }
  return false;
};

export const routes: Routes = [
  { path: 'login', component: LoginComponent },
  { path: 'register', component: RegisterComponent },
  { path: 'challenges', component: ChallengeListComponent, canActivate: [authGuard] },
  { path: 'challenges/:id', component: ChallengeDetailComponent, canActivate: [authGuard] },
  { path: 'scoreboard', component: ScoreboardComponent },
  { path: 'admin', component: AdminDashboardComponent, canActivate: [authGuard] },
  { path: '', redirectTo: '/login', pathMatch: 'full' }
];

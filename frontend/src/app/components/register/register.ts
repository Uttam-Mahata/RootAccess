import { Component } from '@angular/core';
import { RouterModule } from '@angular/router';
import { AuthService } from '../../services/auth';
import { CommonModule } from '@angular/common';

@Component({
  selector: 'app-register',
  standalone: true,
  imports: [
    CommonModule,
    RouterModule
  ],
  templateUrl: './register.html',
  styleUrls: ['./register.scss']
})
export class RegisterComponent {
  constructor(private authService: AuthService) {}

  signUpWithGoogle(): void {
    this.authService.loginWithGoogle();
  }

  // Email/password registration commented out for now - restore when needed
}

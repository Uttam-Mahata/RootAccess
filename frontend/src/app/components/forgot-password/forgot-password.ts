import { Component } from '@angular/core';
import { FormBuilder, FormGroup, Validators, ReactiveFormsModule } from '@angular/forms';
import { CommonModule } from '@angular/common';
import { RouterModule } from '@angular/router';
import { AuthService } from '../../services/auth';

@Component({
  selector: 'app-forgot-password',
  standalone: true,
  imports: [CommonModule, ReactiveFormsModule, RouterModule],
  templateUrl: './forgot-password.html',
  styleUrls: ['./forgot-password.scss']
})
export class ForgotPasswordComponent {
  forgotPasswordForm: FormGroup;
  error = '';
  success = '';
  isLoading = false;

  constructor(
    private fb: FormBuilder,
    private authService: AuthService
  ) {
    this.forgotPasswordForm = this.fb.group({
      email: ['', [Validators.required, Validators.email]]
    });
  }

  onSubmit(): void {
    if (this.forgotPasswordForm.valid && !this.isLoading) {
      this.isLoading = true;
      this.error = '';
      this.success = '';

      const { email } = this.forgotPasswordForm.value;

      this.authService.forgotPassword(email).subscribe({
        next: (response) => {
          this.isLoading = false;
          this.success = response.message || 'If your email is registered, you will receive a password reset link.';
          this.forgotPasswordForm.reset();
        },
        error: (err) => {
          this.isLoading = false;
          this.error = err.error?.error || 'Failed to process request. Please try again.';
        }
      });
    }
  }
}

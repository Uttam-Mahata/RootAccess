import { Component, OnInit } from '@angular/core';
import { FormBuilder, FormGroup, Validators, ReactiveFormsModule } from '@angular/forms';
import { CommonModule } from '@angular/common';
import { Router, ActivatedRoute, RouterModule } from '@angular/router';
import { AuthService } from '../../services/auth';

@Component({
  selector: 'app-reset-password',
  standalone: true,
  imports: [CommonModule, ReactiveFormsModule, RouterModule],
  templateUrl: './reset-password.html',
  styleUrls: ['./reset-password.scss']
})
export class ResetPasswordComponent implements OnInit {
  resetPasswordForm: FormGroup;
  error = '';
  success = '';
  isLoading = false;
  token: string = '';

  constructor(
    private fb: FormBuilder,
    private authService: AuthService,
    private router: Router,
    private route: ActivatedRoute
  ) {
    this.resetPasswordForm = this.fb.group({
      newPassword: ['', [Validators.required, Validators.minLength(8)]],
      confirmPassword: ['', Validators.required]
    }, { validators: this.passwordMatchValidator });
  }

  ngOnInit(): void {
    this.route.queryParams.subscribe(params => {
      this.token = params['token'] || '';
      if (!this.token) {
        this.error = 'Invalid or missing reset token. Please request a new password reset link.';
      }
    });
  }

  passwordMatchValidator(g: FormGroup) {
    const newPassword = g.get('newPassword')?.value;
    const confirmPassword = g.get('confirmPassword')?.value;
    return newPassword === confirmPassword ? null : { 'mismatch': true };
  }

  onSubmit(): void {
    if (this.resetPasswordForm.valid && !this.isLoading && this.token) {
      this.isLoading = true;
      this.error = '';
      this.success = '';

      const { newPassword } = this.resetPasswordForm.value;

      this.authService.resetPassword(this.token, newPassword).subscribe({
        next: (response) => {
          this.isLoading = false;
          this.success = response.message || 'Password reset successfully! Redirecting to login...';
          setTimeout(() => {
            this.router.navigate(['/login']);
          }, 3000);
        },
        error: (err) => {
          this.isLoading = false;
          this.error = err.error?.error || 'Failed to reset password. The link may have expired.';
        }
      });
    }
  }

  get passwordMismatch(): boolean {
    return this.resetPasswordForm.hasError('mismatch') && 
           this.resetPasswordForm.get('confirmPassword')?.touched || false;
  }
}

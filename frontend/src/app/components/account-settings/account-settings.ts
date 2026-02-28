import { Component, OnInit } from '@angular/core';
import { FormBuilder, FormGroup, Validators, ReactiveFormsModule } from '@angular/forms';
import { CommonModule } from '@angular/common';
import { Router, RouterModule } from '@angular/router';
import { AuthService } from '../../services/auth';

@Component({
  selector: 'app-account-settings',
  standalone: true,
  imports: [CommonModule, ReactiveFormsModule, RouterModule],
  templateUrl: './account-settings.html',
  styleUrls: ['./account-settings.scss']
})
export class AccountSettingsComponent implements OnInit {
  changePasswordForm: FormGroup;
  usernameForm: FormGroup;
  error = '';
  success = '';
  usernameError = '';
  usernameSuccess = '';
  isLoading = false;
  isUpdatingUsername = false;
  userInfo: any = null;

  constructor(
    private fb: FormBuilder,
    private authService: AuthService,
    private router: Router
  ) {
    this.changePasswordForm = this.fb.group({
      oldPassword: ['', Validators.required],
      newPassword: ['', [Validators.required, Validators.minLength(8)]],
      confirmPassword: ['', Validators.required]
    }, { validators: this.passwordMatchValidator });

    this.usernameForm = this.fb.group({
      username: ['', [Validators.required, Validators.minLength(3), Validators.maxLength(50)]]
    });
  }

  ngOnInit(): void {
    this.userInfo = this.authService.getCurrentUser();
    if (!this.userInfo) {
      this.router.navigate(['/login']);
    } else {
      this.usernameForm.patchValue({
        username: this.userInfo.username
      });
    }
  }

  passwordMatchValidator(g: FormGroup) {
    const newPassword = g.get('newPassword')?.value;
    const confirmPassword = g.get('confirmPassword')?.value;
    return newPassword === confirmPassword ? null : { 'mismatch': true };
  }

  onSubmit(): void {
    if (this.changePasswordForm.valid && !this.isLoading) {
      this.isLoading = true;
      this.error = '';
      this.success = '';

      const { oldPassword, newPassword } = this.changePasswordForm.value;

      this.authService.changePassword(oldPassword, newPassword).subscribe({
        next: (response) => {
          this.isLoading = false;
          this.success = response.message || 'Password changed successfully!';
          this.changePasswordForm.reset();
        },
        error: (err) => {
          this.isLoading = false;
          this.error = err.error?.error || 'Failed to change password. Please try again.';
        }
      });
    }
  }

  onUpdateUsername(): void {
    if (this.usernameForm.valid && !this.isUpdatingUsername) {
      const newUsername = this.usernameForm.get('username')?.value;
      if (newUsername === this.userInfo?.username) {
        return;
      }

      this.isUpdatingUsername = true;
      this.usernameError = '';
      this.usernameSuccess = '';

      this.authService.updateUsername(newUsername).subscribe({
        next: (response) => {
          this.isUpdatingUsername = false;
          this.usernameSuccess = response.message || 'Username updated successfully!';
          this.userInfo.username = newUsername;
        },
        error: (err) => {
          this.isUpdatingUsername = false;
          this.usernameError = err.error?.error || 'Failed to update username. It might be taken.';
        }
      });
    }
  }

  get passwordMismatch(): boolean {
    return this.changePasswordForm.hasError('mismatch') && 
           this.changePasswordForm.get('confirmPassword')?.touched || false;
  }
}

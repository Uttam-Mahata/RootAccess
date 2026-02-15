import { Component, OnInit } from '@angular/core';
import { FormBuilder, FormGroup, Validators, ReactiveFormsModule } from '@angular/forms';
import { CommonModule } from '@angular/common';
import { Router, RouterModule } from '@angular/router';
import { AuthService } from '../../services/auth';
import { ProfileService, UpdateProfileRequest } from '../../services/profile';

@Component({
  selector: 'app-account-settings',
  standalone: true,
  imports: [CommonModule, ReactiveFormsModule, RouterModule],
  templateUrl: './account-settings.html',
  styleUrls: ['./account-settings.scss']
})
export class AccountSettingsComponent implements OnInit {
  changePasswordForm: FormGroup;
  profileForm: FormGroup;
  error = '';
  success = '';
  profileError = '';
  profileSuccess = '';
  isLoading = false;
  isProfileLoading = false;
  userInfo: any = null;

  constructor(
    private fb: FormBuilder,
    private authService: AuthService,
    private profileService: ProfileService,
    private router: Router
  ) {
    this.changePasswordForm = this.fb.group({
      oldPassword: ['', Validators.required],
      newPassword: ['', [Validators.required, Validators.minLength(8)]],
      confirmPassword: ['', Validators.required]
    }, { validators: this.passwordMatchValidator });

    this.profileForm = this.fb.group({
      bio: ['', Validators.maxLength(500)],
      website: ['', Validators.maxLength(200)],
      github: [''],
      twitter: [''],
      discord: [''],
      linkedin: ['']
    });
  }

  ngOnInit(): void {
    this.userInfo = this.authService.getCurrentUser();
    if (!this.userInfo) {
      this.router.navigate(['/login']);
      return;
    }

    // Load profile data
    this.profileService.getUserProfile(this.userInfo.username).subscribe({
      next: (profile) => {
        this.profileForm.patchValue({
          bio: profile.bio || '',
          website: profile.website || '',
          github: profile.social_links?.github || '',
          twitter: profile.social_links?.twitter || '',
          discord: profile.social_links?.discord || '',
          linkedin: profile.social_links?.linkedin || ''
        });
      },
      error: (err) => {
        console.error('Error loading profile:', err);
      }
    });
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

  onProfileSubmit(): void {
    if (this.profileForm.valid && !this.isProfileLoading) {
      this.isProfileLoading = true;
      this.profileError = '';
      this.profileSuccess = '';

      const formValue = this.profileForm.value;
      const profileUpdate: UpdateProfileRequest = {
        bio: formValue.bio || '',
        website: formValue.website || '',
        social_links: {
          github: formValue.github || '',
          twitter: formValue.twitter || '',
          discord: formValue.discord || '',
          linkedin: formValue.linkedin || ''
        }
      };

      this.profileService.updateProfile(profileUpdate).subscribe({
        next: (response) => {
          this.isProfileLoading = false;
          this.profileSuccess = response.message || 'Profile updated successfully!';
        },
        error: (err) => {
          this.isProfileLoading = false;
          this.profileError = err.error?.error || 'Failed to update profile. Please try again.';
        }
      });
    }
  }

  get passwordMismatch(): boolean {
    return this.changePasswordForm.hasError('mismatch') && 
           this.changePasswordForm.get('confirmPassword')?.touched || false;
  }
}

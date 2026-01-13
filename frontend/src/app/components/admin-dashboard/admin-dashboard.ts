import { Component } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormBuilder, FormGroup, Validators, ReactiveFormsModule } from '@angular/forms';
import { MatCardModule } from '@angular/material/card';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatInputModule } from '@angular/material/input';
import { MatButtonModule } from '@angular/material/button';
import { ChallengeService } from '../../services/challenge';

@Component({
  selector: 'app-admin-dashboard',
  standalone: true,
  imports: [
    CommonModule,
    ReactiveFormsModule,
    MatCardModule,
    MatFormFieldModule,
    MatInputModule,
    MatButtonModule
  ],
  templateUrl: './admin-dashboard.html',
  styleUrls: ['./admin-dashboard.scss']
})
export class AdminDashboardComponent {
  challengeForm: FormGroup;
  message = '';

  constructor(
    private fb: FormBuilder,
    private challengeService: ChallengeService
  ) {
    this.challengeForm = this.fb.group({
      title: ['', Validators.required],
      description: ['', Validators.required],
      category: ['', Validators.required],
      points: [0, Validators.required],
      flag: ['', Validators.required],
      files: [''] // Comma separated for now
    });
  }

  onSubmit(): void {
    if (this.challengeForm.valid) {
      const formValue = this.challengeForm.value;
      const challenge = {
        ...formValue,
        files: formValue.files ? formValue.files.split(',').map((f: string) => f.trim()) : []
      };

      this.challengeService.createChallenge(challenge).subscribe({
        next: () => {
          this.message = 'Challenge created successfully';
          this.challengeForm.reset();
        },
        error: (err) => {
          this.message = 'Error creating challenge';
        }
      });
    }
  }
}

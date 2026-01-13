import { Component, OnInit, OnDestroy } from '@angular/core';
import { CommonModule } from '@angular/common';
import { ActivatedRoute, RouterModule } from '@angular/router';
import { FormBuilder, FormGroup, Validators, ReactiveFormsModule } from '@angular/forms';
import { ChallengeService } from '../../services/challenge';

@Component({
  selector: 'app-challenge-detail',
  standalone: true,
  imports: [
    CommonModule,
    RouterModule,
    ReactiveFormsModule
  ],
  templateUrl: './challenge-detail.html',
  styleUrls: ['./challenge-detail.scss']
})
export class ChallengeDetailComponent implements OnInit, OnDestroy {
  challenge: any;
  flagForm: FormGroup;
  message = '';
  isCorrect = false;
  isSubmitting = false;
  
  // Rate limiting
  isRateLimited = false;
  rateLimitSeconds = 0;
  private rateLimitInterval: any;

  constructor(
    private route: ActivatedRoute,
    private challengeService: ChallengeService,
    private fb: FormBuilder
  ) {
    this.flagForm = this.fb.group({
      flag: ['', Validators.required]
    });
  }

  ngOnInit(): void {
    const id = this.route.snapshot.paramMap.get('id');
    if (id) {
      this.challengeService.getChallenge(id).subscribe({
        next: (data) => this.challenge = data,
        error: (err) => console.error(err)
      });
    }
  }

  ngOnDestroy(): void {
    if (this.rateLimitInterval) {
      clearInterval(this.rateLimitInterval);
    }
  }

  onSubmit(): void {
    if (this.flagForm.valid && this.challenge && !this.isRateLimited && !this.isSubmitting) {
      this.isSubmitting = true;
      this.challengeService.submitFlag(this.challenge.id, this.flagForm.value.flag).subscribe({
        next: (res) => {
          this.message = res.message;
          this.isCorrect = res.correct;
          this.isSubmitting = false;
          
          // Update challenge points if correct
          if (res.correct && res.points) {
            this.challenge.current_points = res.points;
          }
          if (res.solve_count !== undefined) {
            this.challenge.solve_count = res.solve_count;
          }
        },
        error: (err) => {
          this.isSubmitting = false;
          
          // Handle rate limiting (429 Too Many Requests)
          if (err.status === 429) {
            const retryAfter = err.error?.retry_after || 60;
            this.startRateLimitCooldown(retryAfter);
            this.message = `Too many attempts! Please wait ${retryAfter} seconds before trying again.`;
            this.isCorrect = false;
          } else {
            this.message = err.error?.error || 'Error submitting flag';
            this.isCorrect = false;
          }
        }
      });
    }
  }

  private startRateLimitCooldown(seconds: number): void {
    this.isRateLimited = true;
    this.rateLimitSeconds = seconds;

    // Clear any existing interval
    if (this.rateLimitInterval) {
      clearInterval(this.rateLimitInterval);
    }

    this.rateLimitInterval = setInterval(() => {
      this.rateLimitSeconds--;
      if (this.rateLimitSeconds <= 0) {
        this.isRateLimited = false;
        this.rateLimitSeconds = 0;
        this.message = '';
        clearInterval(this.rateLimitInterval);
      }
    }, 1000);
  }
}

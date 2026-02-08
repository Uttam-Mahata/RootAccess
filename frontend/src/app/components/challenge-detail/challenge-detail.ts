import { Component, OnInit, OnDestroy } from '@angular/core';
import { CommonModule } from '@angular/common';
import { ActivatedRoute, RouterModule } from '@angular/router';
import { FormBuilder, FormGroup, Validators, ReactiveFormsModule } from '@angular/forms';
import { ChallengeService, HintResponse } from '../../services/challenge';

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
  writeupForm: FormGroup;
  message = '';
  isCorrect = false;
  isSubmitting = false;
  
  // Hints
  hints: HintResponse[] = [];
  revealingHint: string | null = null;

  // Writeups
  writeups: any[] = [];
  showWriteupForm = false;
  writeupMessage = '';
  isSubmittingWriteup = false;
  
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
    this.writeupForm = this.fb.group({
      content: ['', [Validators.required, Validators.minLength(50)]]
    });
  }

  ngOnInit(): void {
    const id = this.route.snapshot.paramMap.get('id');
    if (id) {
      this.challengeService.getChallenge(id).subscribe({
        next: (data) => {
          this.challenge = data;
          this.loadHints();
          this.loadWriteups();
        },
        error: (err) => console.error(err)
      });
    }
  }

  ngOnDestroy(): void {
    if (this.rateLimitInterval) {
      clearInterval(this.rateLimitInterval);
    }
  }

  loadHints(): void {
    if (!this.challenge) return;
    this.challengeService.getHints(this.challenge.id).subscribe({
      next: (hints) => this.hints = hints || [],
      error: () => this.hints = []
    });
  }

  revealHint(hintId: string): void {
    if (!this.challenge || this.revealingHint) return;
    
    const hint = this.hints.find(h => h.id === hintId);
    if (!hint || hint.revealed) return;

    if (!confirm(`Revealing this hint will cost ${hint.cost} points. Are you sure?`)) return;

    this.revealingHint = hintId;
    this.challengeService.revealHint(this.challenge.id, hintId).subscribe({
      next: (revealed) => {
        const index = this.hints.findIndex(h => h.id === hintId);
        if (index !== -1) {
          this.hints[index] = revealed;
        }
        this.revealingHint = null;
      },
      error: (err) => {
        console.error('Error revealing hint:', err);
        this.revealingHint = null;
      }
    });
  }

  loadWriteups(): void {
    if (!this.challenge) return;
    this.challengeService.getWriteups(this.challenge.id).subscribe({
      next: (writeups) => this.writeups = writeups || [],
      error: () => this.writeups = []
    });
  }

  submitWriteup(): void {
    if (!this.writeupForm.valid || !this.challenge || this.isSubmittingWriteup) return;

    this.isSubmittingWriteup = true;
    this.challengeService.submitWriteup(this.challenge.id, this.writeupForm.value.content).subscribe({
      next: (res) => {
        this.writeupMessage = res.message || 'Writeup submitted for review!';
        this.showWriteupForm = false;
        this.writeupForm.reset();
        this.isSubmittingWriteup = false;
      },
      error: (err) => {
        this.writeupMessage = err.error?.error || 'Error submitting writeup';
        this.isSubmittingWriteup = false;
      }
    });
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

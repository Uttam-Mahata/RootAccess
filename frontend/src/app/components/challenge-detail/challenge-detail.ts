import { Component, OnInit } from '@angular/core';
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
export class ChallengeDetailComponent implements OnInit {
  challenge: any;
  flagForm: FormGroup;
  message = '';
  isCorrect = false;

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

  onSubmit(): void {
    if (this.flagForm.valid && this.challenge) {
      this.challengeService.submitFlag(this.challenge.id, this.flagForm.value.flag).subscribe({
        next: (res) => {
          this.message = res.message;
          this.isCorrect = res.correct;
        },
        error: (err) => {
          this.message = 'Error submitting flag';
        }
      });
    }
  }
}

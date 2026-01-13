import { Component, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { RouterModule } from '@angular/router';
import { MatCardModule } from '@angular/material/card';
import { MatButtonModule } from '@angular/material/button';
import { MatGridListModule } from '@angular/material/grid-list';
import { ChallengeService } from '../../services/challenge';

@Component({
  selector: 'app-challenge-list',
  standalone: true,
  imports: [CommonModule, RouterModule, MatCardModule, MatButtonModule, MatGridListModule],
  templateUrl: './challenge-list.html',
  styleUrls: ['./challenge-list.scss']
})
export class ChallengeListComponent implements OnInit {
  challenges: any[] = [];

  constructor(private challengeService: ChallengeService) { }

  ngOnInit(): void {
    this.challengeService.getChallenges().subscribe({
      next: (data) => this.challenges = data,
      error: (err) => console.error(err)
    });
  }
}

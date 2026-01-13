import { Component, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { RouterModule } from '@angular/router';
import { ChallengeService } from '../../services/challenge';

@Component({
  selector: 'app-challenge-list',
  standalone: true,
  imports: [CommonModule, RouterModule],
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

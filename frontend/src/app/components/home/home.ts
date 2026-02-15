import { Component, inject, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { RouterModule } from '@angular/router';
import { AuthService } from '../../services/auth';
import { TeamService } from '../../services/team';
import { ChallengeService, ChallengeStats } from '../../services/challenge';

@Component({
  selector: 'app-home',
  standalone: true,
  imports: [CommonModule, RouterModule],
  templateUrl: './home.html',
  styleUrls: ['./home.scss']
})
export class HomeComponent implements OnInit {
  authService = inject(AuthService);
  teamService = inject(TeamService);
  challengeService = inject(ChallengeService);

  stats: ChallengeStats | null = null;
  statsLoading = true;

  ngOnInit(): void {
    this.challengeService.getChallengeStats().subscribe({
      next: (data) => {
        this.stats = data;
        this.statsLoading = false;
      },
      error: () => {
        this.statsLoading = false;
      }
    });
  }

  getTotalSolves(): number {
    if (!this.stats?.categories) return 0;
    return this.stats.categories.reduce((sum, cat) => sum + (cat.total_solves || 0), 0);
  }
}
